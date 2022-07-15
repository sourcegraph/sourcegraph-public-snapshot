package repos

import (
	"context"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type ComputeExcludedJob struct {
	RepoOpts search.RepoOptions
}

func (c *ComputeExcludedJob) Run(ctx context.Context, clients job.RuntimeClients, s streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, s, finish := job.StartSpan(ctx, s, c)
	defer func() { finish(alert, err) }()

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

func (c *ComputeExcludedJob) Name() string {
	return "ReposComputeExcludedJob"
}

func (c *ComputeExcludedJob) Fields(v job.Verbosity) (res []log.Field) {
	switch v {
	case job.VerbosityMax:
		fallthrough
	case job.VerbosityBasic:
		res = append(res,
			trace.Scoped("repoOpts", c.RepoOpts.Tags()...),
		)
	}
	return res
}

func (c *ComputeExcludedJob) Children() []job.Describer       { return nil }
func (c *ComputeExcludedJob) MapChildren(job.MapFunc) job.Job { return c }
