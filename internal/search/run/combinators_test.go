package run

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
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
}
