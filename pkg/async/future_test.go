package async

import (
	"context"
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

func Test_Start(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		f := Start[string](nil, func(ctx context.Context) string {
			return Result1
		})
		assert.NotNil(t, f)
	})
}

func Test_Value(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		f := Value[string](Result1)
		assert.NotNil(t, f)
		result := f.Wait()
		assert.Equal(t, Result1, result)
	})
}

func Test_Future_Wait(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		f := Start[string](nil, func(ctx context.Context) string {
			return Result1
		})
		result := f.Wait()
		assert.Equal(t, Result1, result)
	})

	t.Run("canceled", func(t *testing.T) {
		ctx, cfn := context.WithCancel(context.Background())
		f := Start[util.Tuple2[string, error]](ctx, func(ctx context.Context) util.Tuple2[string, error] {
			time.Sleep(500 * time.Millisecond)
			if ctx.Err() != nil {
				return util.Tuple2[string, error]{Val2: ctx.Err()}
			}
			return util.Tuple2[string, error]{Val1: Result2}
		})
		cfn()

		result := f.Wait()
		assert.Zero(t, result.Val1)
		assert.Error(t, result.Val2)
		assert.ErrorIs(t, result.Val2, context.Canceled)
	})
}

func Test_RaceAll(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		start := time.Now()

		ctx := context.Background()
		fns := []FutureFunc[string]{
			func(ctx context.Context) string {
				for i := 1; i <= 5; i++ {
					time.Sleep(100 * time.Millisecond)
				}
				return Result1
			},
			func(ctx context.Context) string {
				for i := 1; i <= 3; i++ {
					time.Sleep(100 * time.Millisecond)
				}
				return Result3
			},
			func(ctx context.Context) string {
				for i := 1; i <= 7; i++ {
					time.Sleep(100 * time.Millisecond)
				}
				return Result2
			},
		}

		result := RaceAll(ctx, fns...).Wait()
		duration := time.Since(start)

		assert.Less(t, duration, 350*time.Millisecond)
		assert.Greater(t, duration, 250*time.Millisecond)
		assert.Equal(t, Result3, result)
	})

	t.Run("empty", func(t *testing.T) {
		ctx := context.Background()
		result := RaceAll[string](ctx).Wait()
		assert.Equal(t, *new(string), result)
	})
}

func Test_WaitAll(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		start := time.Now()

		ctx := context.Background()
		fns := []FutureFunc[string]{
			func(ctx context.Context) string {
				for i := 1; i <= 5; i++ {
					time.Sleep(100 * time.Millisecond)
				}
				return Result1
			},
			func(ctx context.Context) string {
				for i := 1; i <= 7; i++ {
					time.Sleep(100 * time.Millisecond)
				}
				return Result2
			},
			func(ctx context.Context) string {
				for i := 1; i <= 3; i++ {
					time.Sleep(100 * time.Millisecond)
				}
				return Result3
			},
		}

		results := WaitAll(ctx, fns...).Wait()
		duration := time.Since(start)

		assert.Less(t, duration, 750*time.Millisecond)
		assert.Greater(t, duration, 650*time.Millisecond)
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
		start := time.Now()

		ctx := context.Background()
		fns := []FutureFunc[string]{
			func(ctx context.Context) string {
				f := Start[string](ctx, func(ctx context.Context) string {
					for i := 1; i <= 2; i++ {
						time.Sleep(100 * time.Millisecond)
					}
					return Result1
				})
				return f.Wait()
			},
			func(ctx context.Context) string {
				f := Start[string](ctx, func(ctx context.Context) string {
					for i := 1; i <= 3; i++ {
						time.Sleep(100 * time.Millisecond)
					}
					return Result2
				})
				return f.Wait()
			},
			func(ctx context.Context) string {
				f := Start[string](ctx, func(ctx context.Context) string {
					for i := 1; i <= 4; i++ {
						time.Sleep(100 * time.Millisecond)
					}
					return Result3
				})
				return f.Wait()
			},
		}

		results := WaitAll(ctx, fns...).Wait()
		duration := time.Since(start)

		assert.Less(t, duration, 450*time.Millisecond)
		assert.Greater(t, duration, 350*time.Millisecond)
		assert.Len(t, results, 3)
		assert.Equal(t, Result1, results[0])
		assert.Equal(t, Result2, results[1])
		assert.Equal(t, Result3, results[2])
	})
}

func Test_WaitAllMap(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		start := time.Now()

		ctx := context.Background()
		fnm := map[string]FutureFunc[string]{
			"f1": func(ctx context.Context) string {
				for i := 1; i <= 7; i++ {
					time.Sleep(100 * time.Millisecond)
				}
				return Result1
			},
			"f2": func(ctx context.Context) string {
				for i := 1; i <= 5; i++ {
					time.Sleep(100 * time.Millisecond)
				}
				return Result2
			},
			"f3": func(ctx context.Context) string {
				for i := 1; i <= 3; i++ {
					time.Sleep(100 * time.Millisecond)
				}
				return Result3
			},
		}

		results := WaitMap(ctx, fnm).Wait()
		duration := time.Since(start)

		assert.Less(t, duration, 750*time.Millisecond)
		assert.Greater(t, duration, 650*time.Millisecond)
		assert.Len(t, results, 3)
		assert.Equal(t, Result1, results["f1"])
		assert.Equal(t, Result2, results["f2"])
		assert.Equal(t, Result3, results["f3"])
	})

	t.Run("empty", func(t *testing.T) {
		ctx := context.Background()
		results := WaitMap[string, string](ctx, map[string]FutureFunc[string]{}).Wait()
		assert.Len(t, results, 0)
	})
}

func Test_RaceAllMap(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		start := time.Now()

		fnm := map[string]FutureFunc[string]{
			"f1": func(ctx context.Context) string {
				for i := 1; i <= 5; i++ {
					time.Sleep(100 * time.Millisecond)
				}
				return Result1
			},
			"f2": func(ctx context.Context) string {
				for i := 1; i <= 3; i++ {
					time.Sleep(100 * time.Millisecond)
				}
				return Result3
			},
			"f3": func(ctx context.Context) string {
				for i := 1; i <= 7; i++ {
					time.Sleep(100 * time.Millisecond)
				}
				return Result2
			},
		}

		result := RaceAllMap(context.Background(), fnm).Wait()
		duration := time.Since(start)

		assert.Less(t, duration, 350*time.Millisecond)
		assert.Greater(t, duration, 250*time.Millisecond)
		assert.Equal(t, "f2", result.Val1)
		assert.Equal(t, Result3, result.Val2)
	})

	t.Run("empty", func(t *testing.T) {
		result := RaceAllMap[string, string](context.Background(), map[string]FutureFunc[string]{}).Wait()
		assert.Equal(t, "", result.Val1)
		assert.Equal(t, *new(string), result.Val2)
	})
}
