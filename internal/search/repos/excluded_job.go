package repos

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job/jobutil"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

type ComputeExcludedRepos struct {
	Options search.RepoOptions
}

func (c *ComputeExcludedRepos) Run(ctx context.Context, db database.DB, s streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, s, finish := jobutil.StartSpan(ctx, s, c)
	defer func() { finish(alert, err) }()

	repositoryResolver := Resolver{DB: db}
	excluded, err := repositoryResolver.Excluded(ctx, c.Options)
	if err != nil {
		return nil, err
	}

	s.Send(streaming.SearchEvent{
		Stats: streaming.Stats{
			ExcludedArchived: excluded.Archived,
			ExcludedForks:    excluded.Forks,
		},
	})

	return nil, nil
}

func (c *ComputeExcludedRepos) Name() string {
	return "ComputeExcludedRepos"
}
