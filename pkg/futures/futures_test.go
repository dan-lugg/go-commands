package futures

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/dan-lugg/go-commands/pkg/util"
	"github.com/stretchr/testify/assert"
)

const (
	Result1 = "result 1"
	Result2 = "result 2"
	Result3 = "result 3"
)

func Test_Multi(t *testing.T) {

}

func Test_Start(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		ft := Start[string](nil, func(ctx context.Context) string {
			return Result1
		})
		assert.NotNil(t, ft)
	})
}

func Test_Value(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		ft := Value[string](Result1)
		assert.NotNil(t, ft)
		result := ft.Wait()
		assert.Equal(t, Result1, result)
	})
}

func Test_Future_Wait(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		ft := Start[string](nil, func(ctx context.Context) string {
			time.Sleep(100 * time.Millisecond)
			return Result1
		})
		result := ft.Wait()
		assert.Equal(t, Result1, result)
	})

	t.Run("canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ft := Start[util.Tuple2[string, error]](ctx, func(ctx context.Context) util.Tuple2[string, error] {
			time.Sleep(100 * time.Millisecond)
			if ctx.Err() != nil {
				return util.Tuple2[string, error]{Val2: ctx.Err()}
			}
			return util.Tuple2[string, error]{Val1: Result2}
		})
		cancel()
		result := ft.Wait()
		assert.Zero(t, result.Val1)
		assert.Error(t, result.Val2)
		assert.ErrorIs(t, result.Val2, context.Canceled)
	})
}

func Test_WaitAll(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		fns := []FutureFunc[string]{
			func(ctx context.Context) string {
				for range 6 {
					time.Sleep(100 * time.Millisecond)
				}
				return Result1
			},
			func(ctx context.Context) string {
				for range 2 {
					time.Sleep(100 * time.Millisecond)
				}
				return Result2
			},
			func(ctx context.Context) string {
				for range 4 {
					time.Sleep(100 * time.Millisecond)
				}
				return Result3
			},
		}

		ctx := context.Background()
		results := WaitAll(ctx, fns...).Wait()
		assert.Len(t, results, 3)
		assert.Equal(t, Result1, results[0])
		assert.Equal(t, Result2, results[1])
		assert.Equal(t, Result3, results[2])
	})

	t.Run("empty", func(t *testing.T) {
		ctx := context.Background()
		results := WaitAll[string](ctx).Wait()
		assert.Len(t, results, 0)
	})

	t.Run("nested", func(t *testing.T) {
		fns := []FutureFunc[string]{
			func(ctx context.Context) string {
				f := Start[string](ctx, func(ctx context.Context) string {
					for range 6 {
						time.Sleep(100 * time.Millisecond)
					}
					return Result1
				})
				return f.Wait()
			},
			func(ctx context.Context) string {
				f := Start[string](ctx, func(ctx context.Context) string {
					for range 2 {
						time.Sleep(100 * time.Millisecond)
					}
					return Result2
				})
				return f.Wait()
			},
			func(ctx context.Context) string {
				f := Start[string](ctx, func(ctx context.Context) string {
					for range 4 {
						time.Sleep(100 * time.Millisecond)
					}
					return Result3
				})
				return f.Wait()
			},
		}

		ctx := context.Background()
		results := WaitAll(ctx, fns...).Wait()
		assert.Len(t, results, 3)
		assert.Equal(t, Result1, results[0])
		assert.Equal(t, Result2, results[1])
		assert.Equal(t, Result3, results[2])
	})
}

func Test_RaceAll(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		fns := []FutureFunc[string]{
			func(ctx context.Context) string {
				for range 6 {
					time.Sleep(100 * time.Millisecond)
				}
				return Result1
			},
			func(ctx context.Context) string {
				for range 2 {
					time.Sleep(100 * time.Millisecond)
				}
				return Result2
			},
			func(ctx context.Context) string {
				for range 4 {
					time.Sleep(100 * time.Millisecond)
				}
				return Result3
			},
		}

		ctx := context.Background()
		result := RaceAll(ctx, fns...).Wait()
		assert.Equal(t, Result2, result)
	})

	t.Run("empty", func(t *testing.T) {
		ctx := context.Background()
		result := RaceAll[string](ctx).Wait()
		assert.Equal(t, *new(string), result)
	})

	t.Run("canceled", func(t *testing.T) {
		wg := sync.WaitGroup{}
		wg.Add(3)

		fn1Canceled := false
		fn2Canceled := false
		fn3Canceled := false

		fns := []FutureFunc[string]{
			func(ctx context.Context) string {
				defer wg.Done()
				for range 6 {
					time.Sleep(100 * time.Millisecond)
					if ctx.Err() != nil {
						fn1Canceled = true
						return ""
					}
				}
				return Result1
			},
			func(ctx context.Context) string {
				defer wg.Done()
				for range 2 {
					time.Sleep(100 * time.Millisecond)
					if ctx.Err() != nil {
						fn2Canceled = true
						return ""
					}
				}
				return Result2
			},
			func(ctx context.Context) string {
				defer wg.Done()
				for range 4 {
					time.Sleep(100 * time.Millisecond)
					if ctx.Err() != nil {
						fn3Canceled = true
						return ""
					}
				}
				return Result3
			},
		}

		ctx := context.Background()
		result := RaceAll(ctx, fns...).Wait()
		assert.Equal(t, Result2, result)

		wg.Wait()
		assert.True(t, fn1Canceled)
		assert.True(t, fn3Canceled)
		assert.False(t, fn2Canceled)
	})
}

func Test_WaitAllMap(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		fnm := map[string]FutureFunc[string]{
			"f1": func(ctx context.Context) string {
				for range 6 {
					time.Sleep(100 * time.Millisecond)
				}
				return Result1
			},
			"f2": func(ctx context.Context) string {
				for range 2 {
					time.Sleep(100 * time.Millisecond)
				}
				return Result2
			},
			"f3": func(ctx context.Context) string {
				for range 4 {
					time.Sleep(100 * time.Millisecond)
				}
				return Result3
			},
		}

		ctx := context.Background()
		results := WaitAllMap(ctx, fnm).Wait()
		assert.Len(t, results, 3)
		assert.Equal(t, Result1, results["f1"])
		assert.Equal(t, Result2, results["f2"])
		assert.Equal(t, Result3, results["f3"])
	})

	t.Run("empty", func(t *testing.T) {
		ctx := context.Background()
		results := WaitAllMap[string, string](ctx, map[string]FutureFunc[string]{}).Wait()
		assert.Len(t, results, 0)
	})
}

func Test_RaceAllMap(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		fnm := map[string]FutureFunc[string]{
			"f1": func(ctx context.Context) string {
				for range 6 {
					time.Sleep(100 * time.Millisecond)
				}
				return Result1
			},
			"f2": func(ctx context.Context) string {
				for range 2 {
					time.Sleep(100 * time.Millisecond)
				}
				return Result3
			},
			"f3": func(ctx context.Context) string {
				for range 4 {
					time.Sleep(100 * time.Millisecond)
				}
				return Result2
			},
		}

		ctx := context.Background()
		result := RaceAllMap(ctx, fnm).Wait()
		assert.Equal(t, "f2", result.Val1)
		assert.Equal(t, Result3, result.Val2)
	})

	t.Run("empty", func(t *testing.T) {
		ctx := context.Background()
		result := RaceAllMap[string, string](ctx, map[string]FutureFunc[string]{}).Wait()
		assert.Equal(t, "", result.Val1)
		assert.Equal(t, *new(string), result.Val2)
	})

	t.Run("canceled", func(t *testing.T) {
		wg := sync.WaitGroup{}
		wg.Add(3)

		fn1Canceled := false
		fn2Canceled := false
		fn3Canceled := false

		fnm := map[string]FutureFunc[string]{
			"f1": func(ctx context.Context) string {
				defer wg.Done()
				for range 6 {
					time.Sleep(100 * time.Millisecond)
					if ctx.Err() != nil {
						fn1Canceled = true
						return ""
					}
				}
				return Result1
			},
			"f2": func(ctx context.Context) string {
				defer wg.Done()
				for range 2 {
					time.Sleep(100 * time.Millisecond)
					if ctx.Err() != nil {
						fn2Canceled = true
						return ""
					}
				}
				return Result3
			},
			"f3": func(ctx context.Context) string {
				defer wg.Done()
				for range 4 {
					time.Sleep(100 * time.Millisecond)
					if ctx.Err() != nil {
						fn3Canceled = true
						return ""
					}
				}
				return Result2
			},
		}

		ctx := context.Background()
		result := RaceAllMap(ctx, fnm).Wait()
		assert.Equal(t, "f2", result.Val1)
		assert.Equal(t, Result3, result.Val2)

		wg.Wait()
		assert.True(t, fn1Canceled)
		assert.True(t, fn3Canceled)
		assert.False(t, fn2Canceled)
	})
}
