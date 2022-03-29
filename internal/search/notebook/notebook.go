package notebook

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

type SearchJob struct {
	Query string
}

func (s *SearchJob) Run(ctx context.Context, db database.DB, stream streaming.Sender) (alert *search.Alert, err error) {
	store := Search(db)
	notebooks, err := store.SearchNotebooks(ctx, s.Query)
	if err != nil {
		return nil, err
	}
	matches := make([]result.Match, len(notebooks))
	for i, n := range notebooks {
		matches[i] = n
	}
	stream.Send(streaming.SearchEvent{
		Results: matches,
	})
	return nil, nil
}

func (*SearchJob) Name() string {
	return "NotebookSearch"
}
