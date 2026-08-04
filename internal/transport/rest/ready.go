package rest

import "context"

type ReadyChecker interface {
	Ready(ctx context.Context) error
}

type ReadyFunc func(ctx context.Context) error

func (fn ReadyFunc) Ready(ctx context.Context) error {
	return fn(ctx)
}
