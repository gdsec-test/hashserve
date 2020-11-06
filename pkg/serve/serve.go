package serve

import (
	"context"
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
