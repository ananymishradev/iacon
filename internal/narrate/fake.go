package narrate

import (
	"context"
	"sync/atomic"
)

type Fake struct {
	Fixed   Output
	Calls   atomic.Int64
	Blocked <-chan struct{}
}

func (f *Fake) Narrate(ctx context.Context, _ Input) (Output, error) {
	f.Calls.Add(1)
	if f.Blocked != nil {
		select {
		case <-ctx.Done():
			return Output{}, ctx.Err()
		case <-f.Blocked:
		}
	}
	return f.Fixed, nil
}
