package repos

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type ComputeExcludedRepos struct {
	Options search.RepoOptions
}

func (c *ComputeExcludedRepos) Run(ctx context.Context, db database.DB, s streaming.Sender) (err error) {
	tr, ctx := trace.New(ctx, "ComputeExcludedRepos", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	repositoryResolver := Resolver{DB: db}
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
