package evidence

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"
)

type Item struct {
	ID         string         `json:"id"`
	Source     string         `json:"source"`
	OK         bool           `json:"ok"`
	Data       map[string]any `json:"data,omitempty"`
	Error      string         `json:"error,omitempty"`
	DurationMs int64          `json:"durationMs"`
}

type Collector interface {
	Name() string
	Collect(context.Context) (map[string]any, error)
}

type Evidence struct {
	Items       []Item    `json:"items"`
	CollectedAt time.Time `json:"collected_at"`
}

func Run(ctx context.Context, collectors []Collector) Evidence {
	result := Evidence{Items: make([]Item, len(collectors)), CollectedAt: time.Now().UTC()}
	type completed struct {
		index int
		item  Item
	}
	results := make(chan completed, len(collectors))
	bundleCtx, cancelBundle := context.WithTimeout(ctx, 2*time.Second)
	defer cancelBundle()
	var wg sync.WaitGroup
	for i, collector := range collectors {
		wg.Add(1)
		go func(i int, collector Collector) {
			defer wg.Done()
			start := time.Now()
			item := Item{ID: fmt.Sprintf("E%d", i+1), Source: collector.Name()}
			defer func() {
				item.DurationMs = time.Since(start).Milliseconds()
				if recovered := recover(); recovered != nil {
					item.OK = false
					item.Error = fmt.Sprintf("panic: %v", recovered)
					_ = debug.Stack()
				}
				results <- completed{i, item}
			}()
			collectCtx, cancel := context.WithTimeout(bundleCtx, 1500*time.Millisecond)
			defer cancel()
			data, err := collector.Collect(collectCtx)
			if err != nil {
				item.Error = err.Error()
				return
			}
			item.OK = true
			item.Data = data
		}(i, collector)
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-bundleCtx.Done():
	case <-done:
	}
	for {
		select {
		case item := <-results:
			result.Items[item.index] = item.item
		default:
			goto complete
		}
	}
complete:
	for i := range result.Items {
		if result.Items[i].ID == "" {
			result.Items[i] = Item{ID: fmt.Sprintf("E%d", i+1), Source: collectors[i].Name(), Error: "timeout"}
		}
	}
	return result
}
