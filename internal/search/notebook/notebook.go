package notebook

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

type SearchJob struct {
}

func (s *SearchJob) Run(ctx context.Context, db database.DB, stream streaming.Sender) (alert *search.Alert, err error) {
	// TODO:
	//  ~1. New result.NotebookMatch~
	//  2. Search database for input pattern
	//  3. Return NotebookMatches to frontend.
	stream.Send(streaming.SearchEvent{
		Results: result.Matches{
			&result.NotebookMatch{
				Name:          "FOOBAR",
				NamespaceName: "sourcegraph",
				ID:            1,
				Stars:         64,
				Private:       false,
			},
			&result.NotebookMatch{
				Name:          "BAZ",
				NamespaceName: "robert",
				ID:            2,
				Stars:         0,
				Private:       true,
			},
		},
	})

	return nil, nil
}

func (*SearchJob) Name() string {
	return "NotebookSearch"
}
