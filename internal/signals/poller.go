package signals

import (
	"context"
	"time"
)

type Poller struct {
	Interval time.Duration
	Timeout  time.Duration
	Source   func(context.Context) Sample
	W        *Window
}

func (p Poller) Run(ctx context.Context) {
	if p.Interval <= 0 || p.Source == nil || p.W == nil {
		return
	}
	ticker := time.NewTicker(p.Interval)
	defer ticker.Stop()
	var running chan struct{}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if running != nil {
				select {
				case <-running:
					running = nil
				default:
					continue
				}
			}
			done := make(chan struct{})
			running = done
			go func() {
				defer close(done)
				probeCtx := ctx
				cancel := func() {}
				if p.Timeout > 0 {
					probeCtx, cancel = context.WithTimeout(ctx, p.Timeout)
				}
				sample := p.Source(probeCtx)
				cancel()
				p.W.Push(sample)
			}()
		}
	}
}
