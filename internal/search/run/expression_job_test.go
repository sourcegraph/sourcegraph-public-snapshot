package run

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

func newMockSender(waitTime time.Duration) Job {
	res := &result.RepoMatch{Name: "test", ID: 1}
	mj := NewMockJob()
	mj.RunFunc.SetDefaultHook(func(_ context.Context, _ database.DB, s streaming.Sender) (*search.Alert, error) {
		s.Send(streaming.SearchEvent{Results: []result.Match{res}})
		time.Sleep(waitTime)
		return nil, nil
	})
	return mj
}

func newMockSenders(n int, waitTime time.Duration) []Job {
	senders := make([]Job, 0, n)
	for i := 0; i < n; i++ {
		senders = append(senders, newMockSender(waitTime))
	}
	return senders
}

func TestAndJob(t *testing.T) {
	t.Run("NewAndJob", func(t *testing.T) {
		t.Run("no children is simplified", func(t *testing.T) {
			require.Equal(t, NewNoopJob(), NewAndJob())
		})
		t.Run("one child is simplified", func(t *testing.T) {
			j := NewMockJob()
			require.Equal(t, j, NewAndJob(j))
		})
	})

	t.Run("result returned from all subexpressions is streamed", func(t *testing.T) {
		for i := 2; i < 5; i++ {
			t.Run(fmt.Sprintf("%d subexpressions", i), func(t *testing.T) {
				j := NewAndJob(newMockSenders(i, 10*time.Millisecond)...)

				start := time.Now()
				var eventTime time.Time
				s := streaming.StreamFunc(func(_ streaming.SearchEvent) {
					eventTime = time.Now()
				})
				_, err := j.Run(context.Background(), nil, s)
				require.NoError(t, err)

				// event should be streamed ~immediately (definitely before the jobs exit)
				require.Less(t, eventTime.Sub(start), 5*time.Millisecond)
			})
		}
	})

	t.Run("result not returned from all subexpressions is not streamed", func(t *testing.T) {
		noSender := NewMockJob()
		noSender.RunFunc.SetDefaultReturn(nil, nil)

		for i := 2; i < 5; i++ {
			t.Run(fmt.Sprintf("%d subexpressions", i), func(t *testing.T) {
				j := NewAndJob(append([]Job{noSender}, newMockSenders(i-1, 0)...)...)

				eventSent := false
				s := streaming.StreamFunc(func(_ streaming.SearchEvent) {
					eventSent = true
				})
				_, err := j.Run(context.Background(), nil, s)
				require.NoError(t, err)
				require.False(t, eventSent)
			})
		}
	})
}

func TestOrJob(t *testing.T) {
	t.Run("NoOrJob", func(t *testing.T) {
		t.Run("no children is simplified", func(t *testing.T) {
			require.Equal(t, NewNoopJob(), NewOrJob())
		})
		t.Run("one child is simplified", func(t *testing.T) {
			j := NewMockJob()
			require.Equal(t, j, NewOrJob(j))
		})
	})

	t.Run("result returned from all subexpressions is streamed", func(t *testing.T) {
		for i := 2; i < 5; i++ {
			t.Run(fmt.Sprintf("%d subexpressions", i), func(t *testing.T) {
				j := NewOrJob(newMockSenders(i, 10*time.Millisecond)...)

				start := time.Now()
				var eventTime time.Time
				s := streaming.StreamFunc(func(_ streaming.SearchEvent) {
					eventTime = time.Now()
				})
				_, err := j.Run(context.Background(), nil, s)
				require.NoError(t, err)
				require.Less(t, eventTime.Sub(start), 5*time.Millisecond)
			})
		}
	})

	t.Run("result not streamed until all subexpression return the same result", func(t *testing.T) {
		noSender := NewMockJob()
		noSender.RunFunc.SetDefaultReturn(nil, nil)

		for i := 2; i < 5; i++ {
			t.Run(fmt.Sprintf("%d subexpressions", i), func(t *testing.T) {
				j := NewOrJob(append([]Job{noSender}, newMockSenders(i-1, 10*time.Millisecond)...)...)

				var eventTime time.Time
				s := streaming.StreamFunc(func(_ streaming.SearchEvent) {
					eventTime = time.Now()
				})
				_, err := j.Run(context.Background(), nil, s)
				require.NoError(t, err)
				// We should return results that were only matched by some subexpressions
				// right before we finish the job
				require.WithinDuration(t, time.Now(), eventTime, 5*time.Millisecond)
			})
		}
	})

}
