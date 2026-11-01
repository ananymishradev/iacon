package signals

import (
	"sync"
	"time"
)

const capacity = 128

type Window struct {
	mu      sync.RWMutex
	samples []Sample
}

func (w *Window) Push(sample Sample) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if len(w.samples) == capacity {
		copy(w.samples, w.samples[1:])
		w.samples = w.samples[:capacity-1]
	}
	w.samples = append(w.samples, sample)
}

func (w *Window) Last() Sample {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if len(w.samples) == 0 {
		return Sample{State: StateUNKNOWN}
	}
	return w.samples[len(w.samples)-1]
}

func (w *Window) LastState() SampleState {
	return w.Last().State
}

func (w *Window) BADOfLast(n int) (bad, total int) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if n < 0 {
		n = 0
	}
	start := len(w.samples) - n
	if start < 0 {
		start = 0
	}
	for _, sample := range w.samples[start:] {
		total++
		if sample.State == StateBAD {
			bad++
		}
	}
	return bad, total
}

func (w *Window) ConsecutiveBAD() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	count := 0
	for i := len(w.samples) - 1; i >= 0 && w.samples[i].State == StateBAD; i-- {
		count++
	}
	return count
}

func (w *Window) ConsecutiveBADDefinitive() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	found := false
	for i := len(w.samples) - 1; i >= 0 && w.samples[i].State == StateBAD; i-- {
		found = found || w.samples[i].Definitive
	}
	return found
}

func (w *Window) AllOKConsecutive(n int) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if n <= 0 || len(w.samples) < n {
		return false
	}
	for _, sample := range w.samples[len(w.samples)-n:] {
		if sample.State != StateOK {
			return false
		}
	}
	return true
}

func (w *Window) OKSpan() time.Duration {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if len(w.samples) == 0 || w.samples[len(w.samples)-1].State != StateOK {
		return 0
	}
	start := len(w.samples) - 1
	for start > 0 && w.samples[start-1].State == StateOK {
		start--
	}
	return w.samples[len(w.samples)-1].At.Sub(w.samples[start].At)
}
