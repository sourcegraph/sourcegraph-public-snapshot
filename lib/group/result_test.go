package group

import (
	"context"
	"sort"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestResultGroup(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		g := NewWithResults[int]()
		expected := []int{}
		for i := 0; i < 100; i++ {
			i := i
			expected = append(expected, i)
			g.Go(func() int {
				return i
			})
		}
		res := g.Wait()
		sort.Ints(res)
		require.Equal(t, expected, res)
	})

	t.Run("limit", func(t *testing.T) {
		g := NewWithResults[int]().WithMaxConcurrency(1)

		currentConcurrent := int64(0)
		errCount := int64(0)
		for i := 0; i < 10; i++ {
			g.Go(func() int {
				cur := atomic.AddInt64(&currentConcurrent, 1)
				if cur > 1 {
					atomic.AddInt64(&errCount, 1)
				}
				time.Sleep(time.Millisecond)
				atomic.AddInt64(&currentConcurrent, -1)
				return 0
			})
		}
		res := g.Wait()
		require.Len(t, res, 10)
		require.Equal(t, int64(0), errCount)
	})
}

func TestResultErrorGroup(t *testing.T) {
	err1 := errors.New("err1")
	err2 := errors.New("err2")

	t.Run("wait returns no error if no errors", func(t *testing.T) {
		g := NewWithResults[int]().WithErrors()
		g.Go(func() (int, error) { return 1, nil })
		res, err := g.Wait()
		require.NoError(t, err)
		require.Equal(t, []int{1}, res)
	})

	t.Run("wait error if func returns error", func(t *testing.T) {
		g := NewWithResults[int]().WithErrors()
		g.Go(func() (int, error) { return 0, err1 })
		res, err := g.Wait()
		require.Len(t, res, 0) // errored value is ignored
		require.ErrorIs(t, err, err1)
	})

	t.Run("WithCollectErrored", func(t *testing.T) {
		g := NewWithResults[int]().WithErrors().WithCollectErrored()
		g.Go(func() (int, error) { return 0, err1 })
		res, err := g.Wait()
		require.Len(t, res, 1) // errored value is collected
		require.ErrorIs(t, err, err1)
	})

	t.Run("wait error is all returned errors", func(t *testing.T) {
		g := NewWithResults[int]().WithErrors()
		g.Go(func() (int, error) { return 0, err1 })
		g.Go(func() (int, error) { return 0, nil })
		g.Go(func() (int, error) { return 0, err2 })
		res, err := g.Wait()
		require.Len(t, res, 1)
		require.ErrorIs(t, err, err1)
		require.ErrorIs(t, err, err2)
	})

	t.Run("limit", func(t *testing.T) {
		g := NewWithResults[int]().WithErrors().WithMaxConcurrency(1)

		currentConcurrent := int64(0)
		for i := 0; i < 10; i++ {
			g.Go(func() (int, error) {
				cur := atomic.AddInt64(&currentConcurrent, 1)
				if cur > 1 {
					return 0, errors.New("expected no more than 1 concurrent goroutine")
				}
				time.Sleep(time.Millisecond)
				atomic.AddInt64(&currentConcurrent, -1)
				return 0, nil
			})
		}
		res, err := g.Wait()
		require.Len(t, res, 10)
		require.NoError(t, err)
	})
}

func TestResultContextErrorGroup(t *testing.T) {
	err1 := errors.New("err1")
	err2 := errors.New("err2")

	t.Run("behaves the same as ErrorGroup", func(t *testing.T) {
		bgctx := context.Background()
		t.Run("wait returns no error if no errors", func(t *testing.T) {
			g := NewWithResults[int]().WithContext(bgctx)
			g.Go(func(context.Context) (int, error) { return 0, nil })
			res, err := g.Wait()
			require.Len(t, res, 1)
			require.NoError(t, err)
		})

		t.Run("wait error if func returns error", func(t *testing.T) {
			g := NewWithResults[int]().WithContext(bgctx)
			g.Go(func(context.Context) (int, error) { return 0, err1 })
			res, err := g.Wait()
			require.Len(t, res, 0)
			require.ErrorIs(t, err, err1)
		})

		t.Run("wait error is all returned errors", func(t *testing.T) {
			g := NewWithResults[int]().WithErrors().WithContext(bgctx)
			g.Go(func(context.Context) (int, error) { return 0, err1 })
			g.Go(func(context.Context) (int, error) { return 0, nil })
			g.Go(func(context.Context) (int, error) { return 0, err2 })
			res, err := g.Wait()
			require.Len(t, res, 1)
			require.ErrorIs(t, err, err1)
			require.ErrorIs(t, err, err2)
		})
	})

	t.Run("context cancel propagates", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		g := NewWithResults[int]().WithContext(ctx)
		g.Go(func(ctx context.Context) (int, error) {
			<-ctx.Done()
			return 0, ctx.Err()
		})
		cancel()
		res, err := g.Wait()
		require.Len(t, res, 0)
		require.ErrorIs(t, err, context.Canceled)
	})

	t.Run("cancel unblocks limiter", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		g := NewWithResults[int]().WithContext(ctx).WithMaxConcurrency(0)
		g.Go(func(context.Context) (int, error) { return 0, nil })
		res, err := g.Wait()
		require.Len(t, res, 0)
		require.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("CancelOnError", func(t *testing.T) {
		g := NewWithResults[int]().WithContext(context.Background()).WithCancelOnError()
		g.Go(func(ctx context.Context) (int, error) {
			<-ctx.Done()
			return 0, ctx.Err()
		})
		g.Go(func(ctx context.Context) (int, error) {
			return 0, err1
		})
		res, err := g.Wait()
		require.Len(t, res, 0)
		require.ErrorIs(t, err, context.Canceled)
		require.ErrorIs(t, err, err1)
	})

	t.Run("WithFirstError", func(t *testing.T) {
		g := NewWithResults[int]().WithContext(context.Background()).WithCancelOnError().WithFirstError()
		g.Go(func(ctx context.Context) (int, error) {
			<-ctx.Done()
			return 0, err2
		})
		g.Go(func(ctx context.Context) (int, error) {
			return 0, err1
		})
		res, err := g.Wait()
		require.Len(t, res, 0)
		require.ErrorIs(t, err, err1)
		require.NotErrorIs(t, err, context.Canceled)
	})

	t.Run("limit", func(t *testing.T) {
		ctx := context.Background()
		g := NewWithResults[int]().WithContext(ctx).WithMaxConcurrency(1)

		currentConcurrent := int64(0)
		for i := 0; i < 10; i++ {
			g.Go(func(context.Context) (int, error) {
				cur := atomic.AddInt64(&currentConcurrent, 1)
				if cur > 1 {
					return 0, errors.New("expected no more than 1 concurrent goroutine")
				}
				time.Sleep(time.Millisecond)
				atomic.AddInt64(&currentConcurrent, -1)
				return 0, nil
			})
		}
		res, err := g.Wait()
		require.Len(t, res, 10)
		require.NoError(t, err)
	})
}
