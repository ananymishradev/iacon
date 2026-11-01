package signals

import (
	"sync"
	"testing"
	"time"
)

func TestWindow(t *testing.T) {
	w := new(Window)
	base := time.Now()
	for i := 0; i < 129; i++ {
		w.Push(Sample{At: base.Add(time.Duration(i) * time.Second), State: StateOK})
	}
	if got := w.OKSpan(); got != 127*time.Second {
		t.Fatalf("OKSpan = %s", got)
	}
	w.Push(Sample{At: base.Add(129 * time.Second), State: StateUNKNOWN})
	if w.AllOKConsecutive(6) {
		t.Fatal("unknown must not count as OK")
	}
	for i := 0; i < 5; i++ {
		w.Push(Sample{At: base.Add(time.Duration(130+i) * time.Second), State: StateBAD})
	}
	bad, total := w.BADOfLast(6)
	if bad != 5 || total != 6 {
		t.Fatalf("BADOfLast = %d/%d", bad, total)
	}
	if w.AllOKConsecutive(6) {
		t.Fatal("short/non-OK window reported healthy")
	}
	if w.ConsecutiveBAD() != 5 {
		t.Fatalf("ConsecutiveBAD = %d", w.ConsecutiveBAD())
	}
}

func TestWindowRace(t *testing.T) {
	w := new(Window)
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				w.Push(Sample{At: time.Now(), State: SampleState(i % 3)})
				_ = w.Last()
				_, _ = w.BADOfLast(6)
				_ = w.ConsecutiveBADDefinitive()
			}
		}(i)
	}
	wg.Wait()
}
