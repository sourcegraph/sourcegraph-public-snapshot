package group

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestGroup(t *testing.T) {
	t.Parallel()
	t.Run("basic", func(t *testing.T) {
		g := New()
		completed := atomic.NewInt64(0)
		for i := 0; i < 100; i++ {
			g.Go(func() {
				time.Sleep(10 * time.Millisecond)
				completed.Inc()
			})
		}
		g.Wait()
		require.Equal(t, completed.Load(), int64(100))
	})

	t.Run("limit", func(t *testing.T) {
		t.Parallel()
		for _, maxConcurrent := range []int{1, 10, 100} {
			t.Run(strconv.Itoa(maxConcurrent), func(t *testing.T) {
				g := New().WithMaxConcurrency(maxConcurrent)

				currentConcurrent := atomic.NewInt64(0)
				errCount := atomic.NewInt64(0)
				taskCount := maxConcurrent * 10
				for i := 0; i < taskCount; i++ {
					g.Go(func() {
						cur := currentConcurrent.Inc()
						if cur > int64(maxConcurrent) {
							errCount.Inc()
						}
						time.Sleep(time.Millisecond)
						currentConcurrent.Dec()
					})
				}
				g.Wait()
				require.Equal(t, int64(0), errCount.Load())
				require.Equal(t, int64(0), currentConcurrent.Load())
			})
		}
	})

	t.Run("propagate panic", func(t *testing.T) {
		g := New()
		for i := 0; i < 10; i++ {
			i := i
			g.Go(func() {
				if i == 5 {
					panic(i)
				}
			})
		}
		require.Panics(t, func() { g.Wait() })
	})
}

func TestErrorGroup(t *testing.T) {
	t.Parallel()
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
		t.Parallel()
		for _, maxConcurrent := range []int{1, 10, 100} {
			t.Run(strconv.Itoa(maxConcurrent), func(t *testing.T) {
				g := New().WithErrors().WithMaxConcurrency(maxConcurrent)

				currentConcurrent := atomic.NewInt64(0)
				taskCount := maxConcurrent * 10
				for i := 0; i < taskCount; i++ {
					g.Go(func() error {
						cur := currentConcurrent.Inc()
						if cur > int64(maxConcurrent) {
							return errors.Newf("expected no more than %d concurrent goroutine", maxConcurrent)
						}
						time.Sleep(time.Millisecond)
						currentConcurrent.Dec()
						return nil
					})
				}
				require.NoError(t, g.Wait())
				require.Equal(t, int64(0), currentConcurrent.Load())
			})
		}
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
			g := New().WithErrors().WithContext(bgctx)
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
		g := New().WithContext(ctx).WithMaxConcurrency(0)
		g.Go(func(context.Context) error { return nil })
		require.ErrorIs(t, g.Wait(), context.DeadlineExceeded)
	})

	t.Run("CancelOnError", func(t *testing.T) {
		g := New().WithContext(context.Background()).WithCancelOnError()
		g.Go(func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		})
		g.Go(func(ctx context.Context) error {
			return err1
		})
		require.ErrorIs(t, g.Wait(), context.Canceled)
		require.ErrorIs(t, g.Wait(), err1)
	})

	t.Run("WithFirstError", func(t *testing.T) {
		g := New().WithContext(context.Background()).WithCancelOnError().WithFirstError()
		g.Go(func(ctx context.Context) error {
			<-ctx.Done()
			return err2
		})
		g.Go(func(ctx context.Context) error {
			return err1
		})
		require.ErrorIs(t, g.Wait(), err1)
		require.NotErrorIs(t, g.Wait(), context.Canceled)
	})

	t.Run("limit", func(t *testing.T) {
		t.Parallel()
		for _, maxConcurrent := range []int{1, 10, 100} {
			t.Run(strconv.Itoa(maxConcurrent), func(t *testing.T) {
				t.Parallel()
				ctx := context.Background()
				g := New().WithContext(ctx).WithMaxConcurrency(maxConcurrent)

				currentConcurrent := atomic.NewInt64(0)
				for i := 0; i < 100; i++ {
					g.Go(func(context.Context) error {
						cur := currentConcurrent.Inc()
						if cur > int64(maxConcurrent) {
							return errors.Newf("expected no more than %d concurrent goroutine", maxConcurrent)
						}
						time.Sleep(time.Millisecond)
						currentConcurrent.Dec()
						return nil
					})
				}
				require.NoError(t, g.Wait())
				require.Equal(t, int64(0), currentConcurrent.Load())
			})
		}
	})
}
