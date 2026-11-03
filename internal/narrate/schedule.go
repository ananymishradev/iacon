package narrate

import (
	"context"
	"time"
)

func RunAsync(ctx context.Context, n Narrator, input Input, done func(Output, error)) {
	go func() {
		callCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
		defer cancel()
		out, err := n.Narrate(callCtx, input)
		done(out, err)
	}()
}
