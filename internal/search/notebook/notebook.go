package notebook

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

type SearchJob struct {
}

func (s *SearchJob) Run(ctx context.Context, db database.DB, stream streaming.Sender) (alert *search.Alert, err error) {
	// TODO:
	//  1. New result.NotebookMatch
	//  2. Search database for input pattern
	//  3. Return NotebookMatches to frontend.
	stream.Send(streaming.SearchEvent{
		Results: result.Matches{
			&result.RepoMatch{
				Name: api.RepoName("FOOBAR"),
				ID:   1,
			},
		},
	})

	return nil, nil
}

func (*SearchJob) Name() string {
	return "NotebookSearch"
}
