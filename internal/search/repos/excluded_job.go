package repos

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

type ComputeExcludedRepos struct {
	Options search.RepoOptions
}

func (c *ComputeExcludedRepos) Run(ctx context.Context, clients job.RuntimeClients, s streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, s, finish := job.StartSpan(ctx, s, c)
	defer func() { finish(alert, err) }()

	excluded, err := computeExcludedRepos(ctx, clients.DB, c.Options)
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
