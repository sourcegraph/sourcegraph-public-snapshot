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
		res := &result.RepoMatch{Name: "test", ID: 1}
		mj := NewMockJob()
		mj.RunFunc.SetDefaultHook(func(_ context.Context, _ database.DB, s streaming.Sender) (*search.Alert, error) {
			s.Send(streaming.SearchEvent{Results: []result.Match{res}})
			time.Sleep(10 * time.Millisecond)
			return nil, nil
		})

		for i := 2; i < 5; i++ {
			t.Run(fmt.Sprintf("%d subexpressions", i), func(t *testing.T) {
				var subexpressions []Job
				for j := 0; j < i; j++ {
					subexpressions = append(subexpressions, mj)
				}

				j := NewAndJob(subexpressions...)

				start := time.Now()
				var eventTime time.Time
				s := streaming.StreamFunc(func(e streaming.SearchEvent) {
					eventTime = time.Now()
				})
				_, err := j.Run(context.Background(), nil, s)
				require.NoError(t, err)

				// we should wait ~10ms for all parallel subexpressions to complete
				require.WithinDuration(t, start.Add(10*time.Millisecond), time.Now(), 3*time.Millisecond)

				// event should be streamed ~immediately
				require.Less(t, eventTime.Sub(start), time.Millisecond)
			})
		}
	})

	t.Run("result not returned from all subexpressions is not streamed", func(t *testing.T) {
		res := &result.RepoMatch{Name: "test", ID: 1}
		sender, noSender := NewMockJob(), NewMockJob()
		sender.RunFunc.SetDefaultHook(func(_ context.Context, _ database.DB, s streaming.Sender) (*search.Alert, error) {
			s.Send(streaming.SearchEvent{Results: []result.Match{res}})
			return nil, nil
		})
		noSender.RunFunc.SetDefaultReturn(nil, nil)

		for i := 2; i < 5; i++ {
			t.Run(fmt.Sprintf("%d subexpressions", i), func(t *testing.T) {
				subexpressions := []Job{noSender}
				for j := 1; j < i; j++ {
					subexpressions = append(subexpressions, sender)
				}

				j := NewAndJob(subexpressions...)

				eventSent := false
				s := streaming.StreamFunc(func(e streaming.SearchEvent) {
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
		res := &result.RepoMatch{Name: "test", ID: 1}
		mj := NewMockJob()
		mj.RunFunc.SetDefaultHook(func(_ context.Context, _ database.DB, s streaming.Sender) (*search.Alert, error) {
			s.Send(streaming.SearchEvent{Results: []result.Match{res}})
			time.Sleep(10 * time.Millisecond)
			return nil, nil
		})

		for i := 2; i < 5; i++ {
			t.Run(fmt.Sprintf("%d subexpressions", i), func(t *testing.T) {
				var subexpressions []Job
				for j := 0; j < i; j++ {
					subexpressions = append(subexpressions, mj)
				}

				j := NewOrJob(subexpressions...)

				start := time.Now()
				var eventTime time.Time
				s := streaming.StreamFunc(func(e streaming.SearchEvent) {
					eventTime = time.Now()
				})
				_, err := j.Run(context.Background(), nil, s)
				require.NoError(t, err)
				require.Less(t, eventTime.Sub(start), time.Millisecond)
			})
		}
	})

	t.Run("result not returned from all subexpressions is not streamed until all subexpressions return", func(t *testing.T) {
		res := &result.RepoMatch{Name: "test", ID: 1}
		sender, noSender := NewMockJob(), NewMockJob()
		sender.RunFunc.SetDefaultHook(func(_ context.Context, _ database.DB, s streaming.Sender) (*search.Alert, error) {
			s.Send(streaming.SearchEvent{Results: []result.Match{res}})
			time.Sleep(10 * time.Millisecond)
			return nil, nil
		})
		noSender.RunFunc.SetDefaultReturn(nil, nil)

		for i := 2; i < 5; i++ {
			t.Run(fmt.Sprintf("%d subexpressions", i), func(t *testing.T) {
				subexpressions := []Job{noSender}
				for j := 1; j < i; j++ {
					subexpressions = append(subexpressions, sender)
				}

				j := NewOrJob(subexpressions...)

				var eventTime time.Time
				s := streaming.StreamFunc(func(e streaming.SearchEvent) {
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
