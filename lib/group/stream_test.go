package group

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestStreamGroup(t *testing.T) {
	t.Parallel()
	sleepReturn := func(d time.Duration, r int) func() int {
		return func() int {
			time.Sleep(d)
			return r
		}
	}

	t.Run("in order", func(t *testing.T) {
		var results []int
		g := NewWithStreaming[int]().WithMaxConcurrency(10)
		cb := func(i int) { results = append(results, i) }
		g.Go(sleepReturn(5*time.Millisecond, 1), cb)
		g.Go(sleepReturn(10*time.Millisecond, 2), cb)
		g.Go(sleepReturn(15*time.Millisecond, 3), cb)
		g.Wait()
		require.Equal(t, []int{1, 2, 3}, results)
	})

	t.Run("out of order", func(t *testing.T) {
		var results []int
		cb := func(i int) { results = append(results, i) }
		g := NewWithStreaming[int]().WithMaxConcurrency(10)
		g.Go(sleepReturn(15*time.Millisecond, 1), cb)
		g.Go(sleepReturn(10*time.Millisecond, 2), cb)
		g.Go(sleepReturn(5*time.Millisecond, 3), cb)
		g.Wait()
		require.Equal(t, []int{1, 2, 3}, results)
	})

	t.Run("no parallel", func(t *testing.T) {
		var results []int
		cb := func(i int) { results = append(results, i) }
		g := NewWithStreaming[int]().WithMaxConcurrency(1)
		g.Go(sleepReturn(15*time.Millisecond, 1), cb)
		g.Go(sleepReturn(10*time.Millisecond, 2), cb)
		g.Go(sleepReturn(5*time.Millisecond, 3), cb)
		g.Wait()
		require.Equal(t, []int{1, 2, 3}, results)
	})

	t.Run("very parallel", func(t *testing.T) {
		var results []int
		cb := func(i int) { results = append(results, i) }
		g := NewWithStreaming[int]().WithMaxConcurrency(20)
		expected := make([]int, 100)
		for i := 0; i < 100; i++ {
			g.Go(sleepReturn(10*time.Millisecond, i), cb)
			expected[i] = i
		}
		g.Wait()
		require.Equal(t, expected, results)
	})

	t.Run("limit", func(t *testing.T) {
		t.Parallel()
		for _, maxConcurrent := range []int{1, 10, 100} {
			t.Run(strconv.Itoa(maxConcurrent), func(t *testing.T) {
				currentConcurrent := atomic.NewInt64(0)
				errCount := atomic.NewInt64(0)
				taskCount := maxConcurrent * 10

				expected := make([]int, taskCount)
				got := make([]int, 0, taskCount)

				g := NewWithStreaming[int]().WithMaxConcurrency(maxConcurrent)

				cb := func(i int) { got = append(got, i) }
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
					}, cb)
				}
				g.Wait()
				require.Equal(t, int64(0), errCount.Load())
				require.Equal(t, int64(0), currentConcurrent.Load())
				require.Equal(t, expected, got)
			})
		}
	})
}

func TestErrorStreamGroup(t *testing.T) {
	t.Parallel()
	err1 := errors.New("error1")
	err2 := errors.New("error2")

	sleepReturn := func(d time.Duration, r int, err error) func() (int, error) {
		return func() (int, error) {
			time.Sleep(d)
			return r, err
		}
	}

	t.Run("same behavior as StreamGroup", func(t *testing.T) {
		t.Run("in order", func(t *testing.T) {
			var results []int
			g := NewWithStreaming[int]().WithErrors().WithMaxConcurrency(10)
			cb := func(i int, err error) { results = append(results, i) }
			g.Go(sleepReturn(5*time.Millisecond, 1, nil), cb)
			g.Go(sleepReturn(10*time.Millisecond, 2, nil), cb)
			g.Go(sleepReturn(15*time.Millisecond, 3, nil), cb)
			g.Wait()
			require.Equal(t, []int{1, 2, 3}, results)
		})

		t.Run("out of order", func(t *testing.T) {
			var results []int
			cb := func(i int, err error) { results = append(results, i) }
			g := NewWithStreaming[int]().WithErrors().WithMaxConcurrency(10)
			g.Go(sleepReturn(15*time.Millisecond, 1, nil), cb)
			g.Go(sleepReturn(10*time.Millisecond, 2, nil), cb)
			g.Go(sleepReturn(5*time.Millisecond, 3, nil), cb)
			g.Wait()
			require.Equal(t, []int{1, 2, 3}, results)
		})

		t.Run("no parallel", func(t *testing.T) {
			var results []int
			cb := func(i int, err error) { results = append(results, i) }
			g := NewWithStreaming[int]().WithErrors().WithMaxConcurrency(1)
			g.Go(sleepReturn(15*time.Millisecond, 1, nil), cb)
			g.Go(sleepReturn(10*time.Millisecond, 2, nil), cb)
			g.Go(sleepReturn(5*time.Millisecond, 3, nil), cb)
			g.Wait()
			require.Equal(t, []int{1, 2, 3}, results)
		})

		t.Run("very parallel", func(t *testing.T) {
			var results []int
			cb := func(i int, err error) { results = append(results, i) }
			g := NewWithStreaming[int]().WithErrors().WithMaxConcurrency(20)
			expected := make([]int, 100)
			for i := 0; i < 100; i++ {
				g.Go(sleepReturn(10*time.Millisecond, i, nil), cb)
				expected[i] = i
			}
			g.Wait()
			require.Equal(t, expected, results)
		})
	})

	t.Run("errors are passed", func(t *testing.T) {
		var errs error
		cb := func(_ int, err error) { errs = errors.Append(errs, err) }
		g := NewWithStreaming[int]().WithErrors().WithMaxConcurrency(1)
		g.Go(sleepReturn(15*time.Millisecond, 1, err1), cb)
		g.Go(sleepReturn(10*time.Millisecond, 2, nil), cb)
		g.Go(sleepReturn(5*time.Millisecond, 3, err2), cb)
		g.Wait()
		require.ErrorIs(t, errs, err1)
		require.ErrorIs(t, errs, err2)
	})

	t.Run("limit", func(t *testing.T) {
		t.Parallel()
		for _, maxConcurrent := range []int{1, 10, 100} {
			t.Run(strconv.Itoa(maxConcurrent), func(t *testing.T) {
				currentConcurrent := atomic.NewInt64(0)
				taskCount := maxConcurrent * 10

				expected := make([]int, taskCount)
				got := make([]int, 0, taskCount)
				var errs error

				g := NewWithStreaming[int]().WithErrors().WithMaxConcurrency(maxConcurrent)

				cb := func(i int, err error) {
					got = append(got, i)
					errs = errors.Append(errs, err)
				}
				for i := 0; i < taskCount; i++ {
					i := i
					expected[i] = i
					g.Go(func() (int, error) {
						cur := currentConcurrent.Inc()
						if cur > int64(maxConcurrent) {
							return 0, errors.Newf("expected at most %d concurrent calls", maxConcurrent)
						}
						time.Sleep(time.Millisecond)
						currentConcurrent.Dec()
						return i, nil
					}, cb)
				}
				g.Wait()
				require.NoError(t, errs)
				require.Equal(t, int64(0), currentConcurrent.Load())
				require.Equal(t, expected, got)
			})
		}
	})
}

func TestContextErrorStreamGroup(t *testing.T) {
	t.Parallel()
	err1 := errors.New("error1")
	err2 := errors.New("error2")
	bgctx := context.Background()

	sleepReturn := func(d time.Duration, r int, err error) func(context.Context) (int, error) {
		return func(context.Context) (int, error) {
			time.Sleep(d)
			return r, err
		}
	}

	t.Run("same behavior as ErrorStreamGroup", func(t *testing.T) {
		t.Run("in order", func(t *testing.T) {
			var results []int
			g := NewWithStreaming[int]().WithContext(bgctx).WithMaxConcurrency(10)
			cb := func(_ context.Context, i int, err error) { results = append(results, i) }
			g.Go(sleepReturn(5*time.Millisecond, 1, nil), cb)
			g.Go(sleepReturn(10*time.Millisecond, 2, nil), cb)
			g.Go(sleepReturn(15*time.Millisecond, 3, nil), cb)
			g.Wait()
			require.Equal(t, []int{1, 2, 3}, results)
		})

		t.Run("out of order", func(t *testing.T) {
			var results []int
			cb := func(_ context.Context, i int, err error) { results = append(results, i) }
			g := NewWithStreaming[int]().WithErrors().WithContext(bgctx).WithMaxConcurrency(10)
			g.Go(sleepReturn(15*time.Millisecond, 1, nil), cb)
			g.Go(sleepReturn(10*time.Millisecond, 2, nil), cb)
			g.Go(sleepReturn(5*time.Millisecond, 3, nil), cb)
			g.Wait()
			require.Equal(t, []int{1, 2, 3}, results)
		})

		t.Run("no parallel", func(t *testing.T) {
			var results []int
			cb := func(_ context.Context, i int, err error) { results = append(results, i) }
			g := NewWithStreaming[int]().WithContext(bgctx).WithMaxConcurrency(1)
			g.Go(sleepReturn(15*time.Millisecond, 1, nil), cb)
			g.Go(sleepReturn(10*time.Millisecond, 2, nil), cb)
			g.Go(sleepReturn(5*time.Millisecond, 3, nil), cb)
			g.Wait()
			require.Equal(t, []int{1, 2, 3}, results)
		})

		t.Run("very parallel", func(t *testing.T) {
			var results []int
			cb := func(_ context.Context, i int, err error) { results = append(results, i) }
			g := NewWithStreaming[int]().WithContext(bgctx).WithMaxConcurrency(20)
			expected := make([]int, 100)
			for i := 0; i < 100; i++ {
				g.Go(sleepReturn(10*time.Millisecond, i, nil), cb)
				expected[i] = i
			}
			g.Wait()
			require.Equal(t, expected, results)
		})

		t.Run("errors are passed", func(t *testing.T) {
			var errs error
			cb := func(_ context.Context, _ int, err error) { errs = errors.Append(errs, err) }
			g := NewWithStreaming[int]().WithContext(bgctx).WithMaxConcurrency(1)
			g.Go(sleepReturn(15*time.Millisecond, 1, err1), cb)
			g.Go(sleepReturn(10*time.Millisecond, 2, nil), cb)
			g.Go(sleepReturn(5*time.Millisecond, 3, err2), cb)
			g.Wait()
			require.ErrorIs(t, errs, err1)
			require.ErrorIs(t, errs, err2)
		})
	})

	t.Run("context cancels limiter wait", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(bgctx, time.Millisecond)
		defer cancel()
		emptyChan := make(chan struct{})

		var results []int
		var errs error
		cb := func(_ context.Context, i int, err error) {
			results = append(results, i)
			errs = errors.Append(errs, err)
		}
		g := NewWithStreaming[int]().WithContext(ctx).WithMaxConcurrency(0)
		g.Go(func(ctx context.Context) (int, error) {
			<-emptyChan // will never unblock
			return 1, nil
		}, cb)
		g.Wait()

		// should return the zero value for limit errors
		require.Equal(t, results, []int{0})
		// should call callback with context error
		require.ErrorIs(t, errs, context.DeadlineExceeded)
	})

	t.Run("limit", func(t *testing.T) {
		t.Parallel()
		for _, maxConcurrent := range []int{1, 10, 100} {
			t.Run(strconv.Itoa(maxConcurrent), func(t *testing.T) {
				ctx := context.Background()
				currentConcurrent := atomic.NewInt64(0)
				taskCount := maxConcurrent * 10

				expected := make([]int, taskCount)
				got := make([]int, 0, taskCount)
				var errs error

				g := NewWithStreaming[int]().WithContext(ctx).WithMaxConcurrency(maxConcurrent)

				cb := func(_ context.Context, i int, err error) {
					got = append(got, i)
					errs = errors.Append(errs, err)
				}
				for i := 0; i < taskCount; i++ {
					i := i
					expected[i] = i
					g.Go(func(context.Context) (int, error) {
						cur := currentConcurrent.Inc()
						if cur > int64(maxConcurrent) {
							return 0, errors.Newf("expected at most %d concurrent calls", maxConcurrent)
						}
						time.Sleep(time.Millisecond)
						currentConcurrent.Dec()
						return i, nil
					}, cb)
				}
				g.Wait()
				require.NoError(t, errs)
				require.Equal(t, int64(0), currentConcurrent.Load())
				require.Equal(t, expected, got)
			})
		}
	})
}

func BenchmarkStreamGroup(b *testing.B) {
	for _, size := range []int{1, 100, 10000} {
		for _, parallelism := range []int{1, 4, 8, 16, 32} {
			b.Run(fmt.Sprintf("%d_tasks_%d_workers", size, parallelism), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					g := NewWithStreaming[int]().WithMaxConcurrency(parallelism)
					for j := 0; j < size; j++ {
						g.Go(func() int { return 1 }, func(int) {})
					}
					g.Wait()
				}
			})
		}
	}
}
