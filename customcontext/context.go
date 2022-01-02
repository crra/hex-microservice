// package customcontext provide helper functions for context in the
// spirit of https://pkg.go.dev/golang.org/x/sync/errgroup
package customcontext

import (
	"context"
	"sync"
	"time"
)

type contextWithErrorCanceller struct {
	ctx    context.Context
	cancel context.CancelFunc

	err     error
	errOnce sync.Once
}

// WithErrorCanceller extends a context's cancel function with a custom error.
func WithErrorCanceller(parent context.Context) (context.Context, func(error)) {
	ctx, cancel := context.WithCancel(parent)

	ec := &contextWithErrorCanceller{
		ctx:    ctx,
		cancel: cancel,
	}

	return ec, func(e error) {
		ec.errOnce.Do(func() {
			ec.err = e
			cancel()
		})
	}
}

func (ec *contextWithErrorCanceller) Deadline() (deadline time.Time, ok bool) {
	return ec.ctx.Deadline()
}

func (ec *contextWithErrorCanceller) Done() <-chan struct{} {
	return ec.ctx.Done()
}

func (ec *contextWithErrorCanceller) Err() error {
	if ec.err != nil {
		return ec.err
	}

	return ec.ctx.Err()
}

func (ec *contextWithErrorCanceller) Value(key any) any {
	return ec.ctx.Value(key)
}
