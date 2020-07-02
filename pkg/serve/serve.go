package serve

import (
	"context"
	"github.secureserver.net/digital-crimes/hashserve/pkg/backoff"
	"time"
)

// Server is the interface that wraps the basic Serve method.
type Server interface {
	Serve(ctx context.Context) error
}

// The Func type is an adapter to allow the use of ordinary functions as
// event servers. If f is a function with the appropriate signature,
// Func(f) is a Handler that calls f.
type Func func(context.Context) error

// Serve implements Server by calling f(ctx).
func (f Func) Serve(ctx context.Context) error { return f(ctx) }

// WithBackoff applies an exponential backoff strategy for Server w.
// The function will continue to backoff until err is nil.
func WithBackoff(w Server) Server {
	return Func(func(ctx context.Context) error {
		var last time.Time
		for attempt := 1; ; attempt++ {
			last = time.Now().In(time.UTC)

			err := w.Serve(ctx)
			if err == nil {
				return nil
			}

			// If we have already ran for at least the next backoff duration we can skip the timer.
			elapsed := time.Since(last)
			backoff := backoff.Get(attempt)

			if diff := elapsed - backoff; diff > 0 {
				if diff > backoff*10 {
					attempt = 0
				}
				continue
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}
	})
}