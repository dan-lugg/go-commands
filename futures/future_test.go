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
		start := time.Now()

		fut1 := Start[string](nil, func(ctx context.Context) string {
			for i := 1; i <= 5; i++ {
				time.Sleep(100 * time.Millisecond)
			}
			return Result1
		})

		fut2 := Start[string](nil, func(ctx context.Context) string {
			for i := 1; i <= 3; i++ {
				time.Sleep(100 * time.Millisecond)
			}
			return Result3
		})

		fut3 := Start[string](nil, func(ctx context.Context) string {
			for i := 1; i <= 7; i++ {
				time.Sleep(100 * time.Millisecond)
			}
			return Result2
		})
		
		result := RaceAll(fut1, fut2, fut3).Wait()
		duration := time.Since(start)

		assert.Less(t, duration, 350*time.Millisecond)
		assert.Greater(t, duration, 250*time.Millisecond)
		assert.Equal(t, Result3, result)
	})

	t.Run("empty", func(t *testing.T) {
		result := RaceAll[string]().Wait()
		assert.Equal(t, *new(string), result)
	})
}

func Test_WaitAll(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		start := time.Now()

		fut1 := Start[string](nil, func(ctx context.Context) string {
			for i := 1; i <= 7; i++ {
				time.Sleep(100 * time.Millisecond)
			}
			return Result1
		})
		fut2 := Start[string](nil, func(ctx context.Context) string {
			for i := 1; i <= 5; i++ {
				time.Sleep(100 * time.Millisecond)
			}
			return Result2
		})
		fut3 := Start[string](nil, func(ctx context.Context) string {
			for i := 1; i <= 3; i++ {
				time.Sleep(100 * time.Millisecond)
			}
			return Result3
		})

		results := WaitAll(fut1, fut2, fut3).Wait()
		duration := time.Since(start)

		assert.Less(t, duration, 750*time.Millisecond)
		assert.Greater(t, duration, 650*time.Millisecond)
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

		fut1 := Start[string](nil, func(ctx context.Context) string {
			f := Start[string](nil, func(ctx context.Context) string {
				for i := 1; i <= 2; i++ {
					time.Sleep(100 * time.Millisecond)
				}
				return Result1
			})
			return f.Wait()
		})
		fut2 := Start[string](nil, func(ctx context.Context) string {
			f := Start[string](nil, func(ctx context.Context) string {
				for i := 1; i <= 3; i++ {
					time.Sleep(100 * time.Millisecond)
				}
				return Result2
			})
			return f.Wait()
		})
		fut3 := Start[string](nil, func(ctx context.Context) string {
			f := Start[string](nil, func(ctx context.Context) string {
				for i := 1; i <= 4; i++ {
					time.Sleep(100 * time.Millisecond)
				}
				return Result3
			})
			return f.Wait()
		})

		results := WaitAll(fut1, fut2, fut3).Wait()
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

		fut1 := Start[string](nil, func(ctx context.Context) string {
			for i := 1; i <= 7; i++ {
				time.Sleep(100 * time.Millisecond)
			}
			return Result1
		})
		fut2 := Start[string](nil, func(ctx context.Context) string {
			for i := 1; i <= 5; i++ {
				time.Sleep(100 * time.Millisecond)
			}
			return Result2
		})
		fut3 := Start[string](nil, func(ctx context.Context) string {
			for i := 1; i <= 3; i++ {
				time.Sleep(100 * time.Millisecond)
			}
			return Result3
		})

		futMap := map[string]Future[string]{
			"fut1": fut1,
			"fut2": fut2,
			"fut3": fut3,
		}

		results := WaitAllMap(futMap).Wait()
		duration := time.Since(start)

		assert.Less(t, duration, 750*time.Millisecond)
		assert.Greater(t, duration, 650*time.Millisecond)
		assert.Len(t, results, 3)
		assert.Equal(t, Result1, results["fut1"])
		assert.Equal(t, Result2, results["fut2"])
		assert.Equal(t, Result3, results["fut3"])
	})

	t.Run("empty", func(t *testing.T) {
		results := WaitAllMap[string, string](map[string]Future[string]{}).Wait()
		assert.Len(t, results, 0)
	})
}
