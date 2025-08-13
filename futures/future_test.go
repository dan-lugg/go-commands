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

func Test_New(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		f := Start[string](nil, func(ctx context.Context) string {
			return Result1
		})
		assert.NotNil(t, f)
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
			time.Sleep(100 * time.Millisecond)
			return Result1
		})
		f2 := Start[string](nil, func(ctx context.Context) string {
			time.Sleep(200 * time.Millisecond)
			return Result2
		})
		f3 := Start[string](nil, func(ctx context.Context) string {
			time.Sleep(300 * time.Millisecond)
			return Result3
		})
		result := RaceAll(f1, f2, f3).Wait()
		assert.Equal(t, Result1, result)
	})

}

func Test_WaitAll(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		f1 := Start[string](nil, func(ctx context.Context) string {
			time.Sleep(100 * time.Millisecond)
			return Result1
		})
		f2 := Start[string](nil, func(ctx context.Context) string {
			time.Sleep(200 * time.Millisecond)
			return Result2
		})
		f3 := Start[string](nil, func(ctx context.Context) string {
			time.Sleep(300 * time.Millisecond)
			return Result3
		})
		results := WaitAll(f1, f2, f3).Wait()
		assert.Len(t, results, 3)
		assert.Equal(t, Result1, results[0])
		assert.Equal(t, Result2, results[1])
		assert.Equal(t, Result3, results[2])
	})
}
