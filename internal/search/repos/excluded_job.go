package repos

import (
	"context"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type ComputeExcludedReposJob struct {
	RepoOpts search.RepoOptions
}

func (c *ComputeExcludedReposJob) Run(ctx context.Context, clients job.RuntimeClients, s streaming.Sender) (alert *search.Alert, err error) {
	tr, ctx, s, finish := job.StartSpan(ctx, s, c)
	defer func() { finish(alert, err) }()
	tr.TagFields(trace.LazyFields(c.Tags))

	excluded, err := computeExcludedRepos(ctx, clients.DB, c.RepoOpts)
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

func (c *ComputeExcludedReposJob) Name() string {
	return "ComputeExcludedReposJob"
}

func (c *ComputeExcludedReposJob) Tags() []log.Field {
	return []log.Field{
		trace.Stringer("repoOpts", &c.RepoOpts),
	}
}
