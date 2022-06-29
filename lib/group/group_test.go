package group

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestGroup(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		g := New()
		var completed int64 = 0
		for i := 0; i < 100; i++ {
			g.Go(func() {
				time.Sleep(10 * time.Millisecond)
				atomic.AddInt64(&completed, 1)
			})
		}
		g.Wait()
		require.Equal(t, completed, int64(100))
	})

	t.Run("limit", func(t *testing.T) {
		g := New().WithLimit(1)

		currentConcurrent := int64(0)
		errCount := int64(0)
		for i := 0; i < 10; i++ {
			g.Go(func() {
				cur := atomic.AddInt64(&currentConcurrent, 1)
				if cur > 1 {
					atomic.AddInt64(&errCount, 1)
				}
				time.Sleep(time.Millisecond)
				atomic.AddInt64(&currentConcurrent, -1)
			})
		}
		g.Wait()
		require.Equal(t, int64(0), errCount)
	})
}

func TestErrorGroup(t *testing.T) {
	err1 := errors.New("err1")
	err2 := errors.New("err2")

	t.Run("wait returns no error if no errors", func(t *testing.T) {
		g := New().WithErrors()
		g.Go(func() error { return nil })
		require.NoError(t, g.Wait())
	})

	t.Run("wait error if func returns error", func(t *testing.T) {
		g := New().WithErrors()
		g.Go(func() error { return err1 })
		require.ErrorIs(t, g.Wait(), err1)
	})

	t.Run("wait error is all returned errors", func(t *testing.T) {
		g := New().WithErrors()
		g.Go(func() error { return err1 })
		g.Go(func() error { return nil })
		g.Go(func() error { return err2 })
		require.ErrorIs(t, g.Wait(), err1)
		require.ErrorIs(t, g.Wait(), err2)
	})

	t.Run("limit", func(t *testing.T) {
		g := New().WithErrors().WithLimit(1)

		currentConcurrent := int64(0)
		for i := 0; i < 10; i++ {
			g.Go(func() error {
				cur := atomic.AddInt64(&currentConcurrent, 1)
				if cur > 1 {
					return errors.New("expected no more than 1 concurrent goroutine")
				}
				time.Sleep(time.Millisecond)
				atomic.AddInt64(&currentConcurrent, -1)
				return nil
			})
		}
		require.NoError(t, g.Wait())
	})
}

func TestContextErrorGroup(t *testing.T) {
	err1 := errors.New("err1")
	err2 := errors.New("err2")

	t.Run("behaves the same as ErrorGroup", func(t *testing.T) {
		bgctx := context.Background()
		t.Run("wait returns no error if no errors", func(t *testing.T) {
			g := New().WithContext(bgctx)
			g.Go(func(context.Context) error { return nil })
			require.NoError(t, g.Wait())
		})

		t.Run("wait error if func returns error", func(t *testing.T) {
			g := New().WithContext(bgctx)
			g.Go(func(context.Context) error { return err1 })
			require.ErrorIs(t, g.Wait(), err1)
		})

		t.Run("wait error is all returned errors", func(t *testing.T) {
			g := New().WithContext(bgctx)
			g.Go(func(context.Context) error { return err1 })
			g.Go(func(context.Context) error { return nil })
			g.Go(func(context.Context) error { return err2 })
			require.ErrorIs(t, g.Wait(), err1)
			require.ErrorIs(t, g.Wait(), err2)
		})
	})

	t.Run("context cancel propagates", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		g := New().WithContext(ctx)
		g.Go(func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		})
		cancel()
		require.ErrorIs(t, g.Wait(), context.Canceled)
	})

	t.Run("cancel unblocks limiter", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		g := New().WithContext(ctx).WithLimit(0)
		g.Go(func(context.Context) error { return nil })
		require.ErrorIs(t, g.Wait(), context.DeadlineExceeded)
	})

	t.Run("limit", func(t *testing.T) {
		ctx := context.Background()
		g := New().WithContext(ctx).WithLimit(1)

		currentConcurrent := int64(0)
		for i := 0; i < 10; i++ {
			g.Go(func(context.Context) error {
				cur := atomic.AddInt64(&currentConcurrent, 1)
				if cur > 1 {
					return errors.New("expected no more than 1 concurrent goroutine")
				}
				time.Sleep(time.Millisecond)
				atomic.AddInt64(&currentConcurrent, -1)
				return nil
			})
		}
		require.NoError(t, g.Wait())
	})
}
