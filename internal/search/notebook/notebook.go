package notebook

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

type SearchJob struct {
	Query query.Basic
}

func (s *SearchJob) Run(ctx context.Context, db database.DB, stream streaming.Sender) (alert *search.Alert, err error) {
	store := Search(db)
	// TODO:
	// - search over everything by default
	// - search only "full name" on 'notebook:' filter
	// - account for search pattern types
	// - actually filter blocks (we return all right now)
	notebooks, err := store.SearchNotebooks(ctx, s.Query.PatternString(), true)
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
