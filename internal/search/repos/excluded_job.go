package repos

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

type ComputeExcludedRepos struct {
	DB      database.DB
	Options search.RepoOptions
}

func (c *ComputeExcludedRepos) Run(ctx context.Context, s streaming.Sender, _ Pager) (err error) {
	repositoryResolver := Resolver{DB: c.DB}
	excluded, err := repositoryResolver.Excluded(ctx, c.Options)
	if err != nil {
		return err
	}

	s.Send(streaming.SearchEvent{
		Stats: streaming.Stats{
			ExcludedArchived: excluded.Archived,
			ExcludedForks:    excluded.Forks,
		},
	})

	return nil
}

func (c *ComputeExcludedRepos) Name() string {
	return "ComputeExcludedRepos"
}
