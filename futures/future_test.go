package futures

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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
}

func Test_RaceAll(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		f1 := Start[string](nil, func(ctx context.Context) string {
			for i := 1; i <= 7; i++ {
				time.Sleep(1 * time.Second)
				println("Processing in f1:", i)
			}
			return Result1
		})
		f2 := Start[string](nil, func(ctx context.Context) string {
			for i := 1; i <= 5; i++ {
				time.Sleep(1 * time.Second)
				println("Processing in f2:", i)
			}
			return Result2
		})
		f3 := Start[string](nil, func(ctx context.Context) string {
			for i := 1; i <= 3; i++ {
				time.Sleep(1 * time.Second)
				println("Processing in f3:", i)
			}
			return Result3
		})
		result := RaceAll(f1, f2, f3).Wait()
		assert.Equal(t, Result3, result)
	})

}

func Test_WaitAll(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		f1 := Start[string](nil, func(ctx context.Context) string {
			for i := 1; i <= 7; i++ {
				time.Sleep(1 * time.Second)
				println("Processing in f1:", i)
			}
			return Result1
		})
		f2 := Start[string](nil, func(ctx context.Context) string {
			for i := 1; i <= 5; i++ {
				time.Sleep(1 * time.Second)
				println("Processing in f2:", i)
			}
			return Result2
		})
		f3 := Start[string](nil, func(ctx context.Context) string {
			for i := 1; i <= 3; i++ {
				time.Sleep(1 * time.Second)
				println("Processing in f3:", i)
			}
			return Result3
		})
		results := WaitAll(f1, f2, f3).Wait()
		assert.Len(t, results, 3)
		assert.Equal(t, Result1, results[0])
		assert.Equal(t, Result2, results[1])
		assert.Equal(t, Result3, results[2])
	})

	t.Run("empty", func(t *testing.T) {
		results := WaitAll[string]().Wait()
		assert.Len(t, results, 0)
	})

	t.Run("nested", func(t *testing.T) {
		start := time.Now()

		f1 := Start[string](nil, func(ctx context.Context) string {
			f := Start[string](nil, func(ctx context.Context) string {
				for i := 1; i <= 2; i++ {
					time.Sleep(1 * time.Second)
					println("Processing in f1:", i)
				}
				return Result1
			})
			return f.Wait()
		})
		f2 := Start[string](nil, func(ctx context.Context) string {
			f := Start[string](nil, func(ctx context.Context) string {
				for i := 1; i <= 3; i++ {
					time.Sleep(1 * time.Second)
					println("Processing in f2:", i)
				}
				return Result2
			})
			return f.Wait()
		})
		f3 := Start[string](nil, func(ctx context.Context) string {
			f := Start[string](nil, func(ctx context.Context) string {
				for i := 1; i <= 4; i++ {
					time.Sleep(1 * time.Second)
					println("Processing in f3:", i)
				}
				return Result3
			})
			return f.Wait()
		})

		results := WaitAll(f1, f2, f3).Wait()

		duration := time.Since(start)

		assert.Less(t, duration, 5*time.Second)
		assert.Greater(t, duration, 3*time.Second)

		assert.Len(t, results, 3)
		assert.Equal(t, Result1, results[0])
		assert.Equal(t, Result2, results[1])
		assert.Equal(t, Result3, results[2])
	})
}
