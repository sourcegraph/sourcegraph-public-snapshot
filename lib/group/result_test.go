package group

import (
	"context"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestResultGroup(t *testing.T) {
	t.Parallel()
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
		t.Parallel()
		for _, maxConcurrent := range []int{1, 10, 100} {
			t.Run(strconv.Itoa(maxConcurrent), func(t *testing.T) {
				g := NewWithResults[int]().WithMaxConcurrency(maxConcurrent)

				currentConcurrent := atomic.NewInt64(0)
				errCount := atomic.NewInt64(0)
				taskCount := maxConcurrent * 10
				expected := make([]int, taskCount)
				for i := 0; i < taskCount; i++ {
					i := i
					expected[i] = i
					g.Go(func() int {
						cur := currentConcurrent.Inc()
						if cur > int64(maxConcurrent) {
							errCount.Inc()
						}
						time.Sleep(time.Millisecond)
						currentConcurrent.Dec()
						return i
					})
				}
				res := g.Wait()
				sort.Ints(res)
				require.Equal(t, expected, res)
				require.Equal(t, int64(0), errCount.Load())
				require.Equal(t, int64(0), currentConcurrent.Load())
			})
		}
	})
}

func TestResultErrorGroup(t *testing.T) {
	t.Parallel()
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

	t.Run("WithFirstError", func(t *testing.T) {
		t.Parallel()
		g := NewWithResults[int]().WithErrors().WithFirstError()
		synchronizer := make(chan struct{})
		g.Go(func() (int, error) {
			<-synchronizer
			// This test has an intrinsic race condition that can be reproduced
			// by adding a `defer time.Sleep(time.Second)` before the `defer
			// close(synchronizer)`. We cannot guarantee that the group processes
			// the return value of the second goroutine before the first goroutine
			// exits in response to synchronizer, so we add a sleep here to make
			// this race condition vanishingly unlikely. Note that this is a race
			// in the test, not in the library.
			time.Sleep(100 * time.Millisecond)
			return 0, err1
		})
		g.Go(func() (int, error) {
			defer close(synchronizer)
			return 0, err2
		})
		res, err := g.Wait()
		require.Len(t, res, 0)
		require.ErrorIs(t, err, err2)
		require.NotErrorIs(t, err, err1)
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
		for _, maxConcurrency := range []int{1, 10, 100} {
			t.Run(strconv.Itoa(maxConcurrency), func(t *testing.T) {
				t.Parallel()
				g := NewWithResults[int]().WithErrors().WithMaxConcurrency(maxConcurrency)

				currentConcurrent := atomic.NewInt64(0)
				taskCount := maxConcurrency * 10
				for i := 0; i < taskCount; i++ {
					g.Go(func() (int, error) {
						cur := currentConcurrent.Inc()
						if cur > int64(maxConcurrency) {
							return 0, errors.Newf("expected no more than %d concurrent goroutine", maxConcurrency)
						}
						time.Sleep(time.Millisecond)
						currentConcurrent.Dec()
						return 0, nil
					})
				}
				res, err := g.Wait()
				require.Len(t, res, taskCount)
				require.NoError(t, err)
				require.Equal(t, int64(0), currentConcurrent.Load())
			})
		}
	})
}

func TestResultContextErrorGroup(t *testing.T) {
	t.Parallel()
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
		t.Parallel()
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
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		g := NewWithResults[int]().WithContext(ctx).WithMaxConcurrency(0)
		g.Go(func(context.Context) (int, error) { return 0, nil })
		res, err := g.Wait()
		require.Len(t, res, 0)
		require.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("CancelOnError", func(t *testing.T) {
		t.Parallel()
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

	t.Run("WithCollectErrored", func(t *testing.T) {
		g := NewWithResults[int]().WithContext(context.Background()).WithCollectErrored()
		g.Go(func(context.Context) (int, error) { return 0, err1 })
		res, err := g.Wait()
		require.Len(t, res, 1) // errored value is collected
		require.ErrorIs(t, err, err1)
	})

	t.Run("WithFirstError", func(t *testing.T) {
		t.Parallel()
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
		t.Parallel()
		for _, maxConcurrency := range []int{1, 10, 100} {
			t.Run(strconv.Itoa(maxConcurrency), func(t *testing.T) {
				t.Parallel()
				ctx := context.Background()
				g := NewWithResults[int]().WithContext(ctx).WithMaxConcurrency(maxConcurrency)

				currentConcurrent := atomic.NewInt64(0)
				taskCount := maxConcurrency * 10
				expected := make([]int, taskCount)
				for i := 0; i < taskCount; i++ {
					i := i
					expected[i] = i
					g.Go(func(context.Context) (int, error) {
						cur := currentConcurrent.Inc()
						if cur > int64(maxConcurrency) {
							return 0, errors.Newf("expected no more than %d concurrent goroutines", maxConcurrency)
						}
						time.Sleep(time.Millisecond)
						currentConcurrent.Dec()
						return i, nil
					})
				}
				res, err := g.Wait()
				sort.Ints(res)
				require.Equal(t, expected, res)
				require.NoError(t, err)
				require.Equal(t, int64(0), currentConcurrent.Load())
			})
		}
	})
}
