package job

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestLimitJob(t *testing.T) {
	t.Run("only send limit", func(t *testing.T) {
		mockJob := NewMockJob()
		mockJob.RunFunc.SetDefaultHook(func(ctx context.Context, db database.DB, s streaming.Sender) (*search.Alert, error) {
			for i := 0; i < 10; i++ {
				s.Send(streaming.SearchEvent{
					Results: []result.Match{&result.FileMatch{}},
				})
			}
			return nil, nil
		})

		var sent []result.Match
		stream := streaming.StreamFunc(func(e streaming.SearchEvent) {
			sent = append(sent, e.Results...)
		})

		limitJob := NewLimitJob(5, mockJob)
		limitJob.Run(context.Background(), database.NewMockDB(), stream)

		// The number sent is one more than the limit because
		// the stream limiter only cancels after the limit is exceeded,
		// but doesn't stop the results from going through.
		require.Equal(t, 5, len(sent))
	})

	t.Run("send partial event", func(t *testing.T) {
		mockJob := NewMockJob()
		mockJob.RunFunc.SetDefaultHook(func(ctx context.Context, db database.DB, s streaming.Sender) (*search.Alert, error) {
			for i := 0; i < 10; i++ {
				s.Send(streaming.SearchEvent{
					Results: []result.Match{
						&result.FileMatch{},
						&result.FileMatch{},
					},
				})
			}
			return nil, nil
		})

		var sent []result.Match
		stream := streaming.StreamFunc(func(e streaming.SearchEvent) {
			sent = append(sent, e.Results...)
		})

		limitJob := NewLimitJob(5, mockJob)
		limitJob.Run(context.Background(), database.NewMockDB(), stream)

		// The number sent is one more than the limit because
		// the stream limiter only cancels after the limit is exceeded,
		// but doesn't stop the results from going through.
		require.Equal(t, 5, len(sent))
	})

	t.Run("cancel after limit", func(t *testing.T) {
		mockJob := NewMockJob()
		mockJob.RunFunc.SetDefaultHook(func(ctx context.Context, db database.DB, s streaming.Sender) (*search.Alert, error) {
			for i := 0; i < 10; i++ {
				select {
				case <-ctx.Done():
					return nil, nil
				default:
				}
				s.Send(streaming.SearchEvent{
					Results: []result.Match{&result.FileMatch{}},
				})
			}
			return nil, nil
		})

		var sent []result.Match
		stream := streaming.StreamFunc(func(e streaming.SearchEvent) {
			sent = append(sent, e.Results...)
		})

		limitJob := NewLimitJob(5, mockJob)
		limitJob.Run(context.Background(), database.NewMockDB(), stream)

		// The number sent is one more than the limit because
		// the stream limiter only cancels after the limit is exceeded,
		// but doesn't stop the results from going through.
		require.Equal(t, 5, len(sent))
	})

	t.Run("NewLimitJob propagates noop", func(t *testing.T) {
		job := NewLimitJob(10, NewNoopJob())
		require.Equal(t, NewNoopJob(), job)
	})
}

func TestTimeoutJob(t *testing.T) {
	t.Run("timeout works", func(t *testing.T) {
		timeoutWaiter := NewMockJob()
		timeoutWaiter.RunFunc.SetDefaultHook(func(ctx context.Context, _ database.DB, _ streaming.Sender) (*search.Alert, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		})
		timeoutJob := NewTimeoutJob(10*time.Millisecond, timeoutWaiter)
		_, err := timeoutJob.Run(context.Background(), nil, nil)
		require.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("early return returns early", func(t *testing.T) {
		timeoutWaiter := NewMockJob()
		timeoutWaiter.RunFunc.SetDefaultHook(func(ctx context.Context, _ database.DB, _ streaming.Sender) (*search.Alert, error) {
			select {
			case <-time.After(10 * time.Millisecond):
				return nil, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		})
		timeoutJob := NewTimeoutJob(time.Second, timeoutWaiter)
		start := time.Now()
		_, err := timeoutJob.Run(context.Background(), nil, nil)
		require.NoError(t, err)
		require.WithinDuration(t, time.Now(), start.Add(10*time.Millisecond), 5*time.Millisecond)
	})

	t.Run("NewTimeoutJob propagates noop", func(t *testing.T) {
		job := NewTimeoutJob(10*time.Second, NewNoopJob())
		require.Equal(t, NewNoopJob(), job)
	})
}

func TestParallelJob(t *testing.T) {
	t.Run("jobs run in parallel", func(t *testing.T) {
		waiter := NewMockJob()
		waiter.RunFunc.SetDefaultHook(func(ctx context.Context, _ database.DB, _ streaming.Sender) (*search.Alert, error) {
			time.Sleep(10 * time.Millisecond)
			return nil, nil
		})
		parallelJob := NewParallelJob(waiter, waiter, waiter)
		start := time.Now()
		_, err := parallelJob.Run(context.Background(), nil, nil)
		require.NoError(t, err)
		require.WithinDuration(t, time.Now(), start.Add(20*time.Millisecond), 10*time.Millisecond)
	})

	t.Run("errors are aggregated", func(t *testing.T) {
		e1 := errors.New("error 1")
		e2 := errors.New("error 2")
		j1, j2 := NewMockJob(), NewMockJob()
		j1.RunFunc.SetDefaultReturn(nil, e1)
		j2.RunFunc.SetDefaultReturn(nil, e2)

		parallelJob := NewParallelJob(j1, j2)
		_, err := parallelJob.Run(context.Background(), nil, nil)
		require.ErrorIs(t, err, e1)
		require.ErrorIs(t, err, e2)
	})

	t.Run("alerts are aggregated", func(t *testing.T) {
		a1 := &search.Alert{Priority: 1}
		a2 := &search.Alert{Priority: 2}
		j1, j2 := NewMockJob(), NewMockJob()
		j1.RunFunc.SetDefaultReturn(a1, nil)
		j2.RunFunc.SetDefaultReturn(a2, nil)

		parallelJob := NewParallelJob(j1, j2)
		alert, err := parallelJob.Run(context.Background(), nil, nil)
		require.NoError(t, err)
		require.Equal(t, a2, alert)
	})

	t.Run("NewParallelJob", func(t *testing.T) {
		t.Run("no children is simplified", func(t *testing.T) {
			require.Equal(t, NewNoopJob(), NewParallelJob())
		})

		t.Run("one child is simplified", func(t *testing.T) {
			m := NewMockJob()
			require.Equal(t, m, NewParallelJob(m))
		})
	})
}

func TestPriorityJob(t *testing.T) {
	t.Run("optional job is canceled after required finishes", func(t *testing.T) {
		required, optional := NewMockJob(), NewMockJob()
		required.RunFunc.SetDefaultReturn(nil, nil)
		optional.RunFunc.SetDefaultHook(func(ctx context.Context, _ database.DB, _ streaming.Sender) (*search.Alert, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(200 * time.Millisecond):
				return nil, nil
			}
		})

		start := time.Now()
		job := NewPriorityJob(required, optional)
		_, err := job.Run(context.Background(), nil, nil)
		require.ErrorIs(t, err, context.Canceled)
		require.WithinDuration(t, time.Now(), start.Add(100*time.Millisecond), 40*time.Millisecond)
	})

	t.Run("optional job has some time to complete", func(t *testing.T) {
		required, optional := NewMockJob(), NewMockJob()
		required.RunFunc.SetDefaultReturn(nil, nil)
		optional.RunFunc.SetDefaultHook(func(ctx context.Context, _ database.DB, _ streaming.Sender) (*search.Alert, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(50 * time.Millisecond):
				return nil, nil
			}
		})

		start := time.Now()
		job := NewPriorityJob(required, optional)
		_, err := job.Run(context.Background(), nil, nil)
		require.NoError(t, err, context.Canceled)
		require.WithinDuration(t, time.Now(), start.Add(50*time.Millisecond), 30*time.Millisecond)
	})

	t.Run("NewPriorityJob", func(t *testing.T) {
		t.Run("noop optional is simplified", func(t *testing.T) {
			j := NewMockJob()
			require.Equal(t, j, NewPriorityJob(j, NewNoopJob()))
		})
	})
}
