package async

import (
	"context"
	"sync"

	"github.com/dan-lugg/go-commands/pkg/util"
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

type FutureFunc[R any] func(ctx context.Context) R

// Start begins a computation that runs the provided function fn in a separate goroutine.
// The computation's result of type R can be retrieved by calling the Wait method on the returned Future.
// The provided ctx is passed to the function fn to support context-aware operations.
func Start[R any](ctx context.Context, fn FutureFunc[R]) Future[R] {
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
	return Start(nil, func(ctx context.Context) R {
		return value
	})
}

// WaitAll takes multiple Future instances and returns a new Future that resolves
// to a slice of results once all the provided Future instances have completed.
// The results are returned in the same order as the input Future instances.
func WaitAll[R any](ctx context.Context, fns ...FutureFunc[R]) Future[[]R] {
	if len(fns) == 0 {
		return Value([]R{})
	}
	fts := make([]Future[R], len(fns))
	for i := 0; i < len(fns); i++ {
		fts[i] = Start(ctx, fns[i])
	}
	return Start(nil, func(ctx context.Context) []R {
		r := make([]R, len(fts))
		for i, f := range fts {
			r[i] = f.Wait()
		}
		return r
	})
}

// WaitMap takes a map of Future instances and returns a new Future
// that resolves to a map containing the results of all the provided
// Future instances. The keys in the resulting map correspond to the
// keys in the input map, and the values are the results of the
// respective Future computations.
func WaitMap[K comparable, R any](ctx context.Context, fnm map[K]FutureFunc[R]) Future[map[K]R] {
	if len(fnm) == 0 {
		return Value(map[K]R{})
	}
	ftm := make(map[K]Future[R], len(fnm))
	for k, f := range fnm {
		ftm[k] = Start(ctx, f)
	}
	return Start(nil, func(ctx context.Context) map[K]R {
		r := make(map[K]R, len(ftm))
		for k, f := range ftm {
			r[k] = f.Wait()
		}
		return r
	})
}

// RaceAll takes multiple Future instances and returns a new Future
// that resolves to the result of the first Future to complete.
// The remaining Future computations are not canceled and will continue
// to execute in the background.
func RaceAll[R any](ctx context.Context, fns ...FutureFunc[R]) Future[R] {
	if len(fns) == 0 {
		return Value(*new(R))
	}
	fts := make([]Future[R], len(fns))
	for i := 0; i < len(fns); i++ {
		fts[i] = Start(ctx, fns[i])
	}
	return Start(nil, func(ctx context.Context) R {
		ch := make(chan R, len(fts))
		for i := 0; i < len(fts); i++ {
			i_ := i
			go func() {
				ch <- fts[i_].Wait()
			}()
		}
		return <-ch
	})
}

// RaceAllMap takes a map of Future instances and returns a new Future
// that resolves to a tuple containing the key and result of the first
// Future to complete. The remaining Future computations are not canceled
// and will continue to execute in the background.
func RaceAllMap[K comparable, R any](ctx context.Context, fnm map[K]FutureFunc[R]) Future[util.Tuple2[K, R]] {
	if len(fnm) == 0 {
		return Value(util.Tuple2[K, R]{})
	}
	ftm := make(map[K]Future[R], len(fnm))
	for k, f := range fnm {
		ftm[k] = Start(ctx, f)
	}
	return Start(nil, func(ctx context.Context) util.Tuple2[K, R] {
		ch := make(chan util.Tuple2[K, R], len(fnm))
		for k := range fnm {
			k_ := k
			f_ := Start(ctx, fnm[k_])
			go func() {
				ch <- util.Tuple2[K, R]{
					Val1: k_,
					Val2: f_.Wait(),
				}
			}()
		}
		return <-ch
	})
}
