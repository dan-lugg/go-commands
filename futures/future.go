package futures

import (
	"context"
	"sync"
)

// Future represents a computation that will produce a result of type R in the future.
// The result can be retrieved by calling the Wait method, which blocks until the computation is complete.
type Future[R any] interface {
	Wait() R
}

type future[R any] struct {
	result    R
	waitGroup sync.WaitGroup
}

// Wait blocks until the computation represented by the Future is complete
// and returns the result of the computation.
func (f *future[R]) Wait() R {
	f.waitGroup.Wait()
	return f.result
}

// Start begins a computation that runs the provided function fn in a separate goroutine.
// The computation's result of type R can be retrieved by calling the Wait method on the returned Future.
// The provided ctx is passed to the function fn to support context-aware operations.
func Start[R any](ctx context.Context, fn func(ctx context.Context) R) Future[R] {
	f := future[R]{}
	f.waitGroup.Add(1)
	go func() {
		defer f.waitGroup.Done()
		f.result = fn(ctx)
	}()
	return &f
}

// Value creates a Future that immediately resolves to the provided value.
// The computation runs in a separate goroutine and can be awaited using the Wait method.
func Value[R any](value R) Future[R] {
	return Start(context.Background(), func(ctx context.Context) R {
		return value
	})
}

// WaitAll takes multiple Future instances and returns a new Future that resolves
// to a slice of results once all the provided Future instances have completed.
// The results are returned in the same order as the input Future instances.
func WaitAll[R any](futures ...Future[R]) Future[[]R] {
	return Start(context.Background(), func(ctx context.Context) []R {
		r := make([]R, len(futures))
		for i, f := range futures {
			r[i] = f.Wait()
		}
		return r
	})
}

func WaitAllMap[K comparable, R any](m map[K]Future[R]) Future[map[K]R] {
	return Start(context.Background(), func(ctx context.Context) map[K]R {
		r := make(map[K]R, len(m))
		for k, f := range m {
			r[k] = f.Wait()
		}
		return r
	})
}

// RaceAll takes multiple Future instances and returns a new Future
// that resolves to the result of the first Future to complete.
// The remaining Future computations are not canceled and will continue
// to execute in the background.
func RaceAll[R any](futures ...Future[R]) Future[R] {
	if len(futures) == 0 {
		return Value(*new(R))
	}
	return Start(context.Background(), func(ctx context.Context) R {
		ch := make(chan R, len(futures))
		for i := 0; i < len(futures); i++ {
			i_ := i
			go func() {
				ch <- futures[i_].Wait()
			}()
		}
		return <-ch
	})
}
