package codycontext

import (
	"context"

	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewSearchJob(plan query.Plan, newJob func(query.Basic) (job.Job, error)) (job.Job, error) {
	if len(plan) > 1 {
		return nil, errors.New("The 'codycontext' patterntype does not support multiple clauses")
	}

	basicQuery := plan[0].ToParseTree()
	q, err := queryStringToKeywordQuery(query.StringHuman(basicQuery))
	if err != nil || q == nil {
		return nil, err
	}

	child, err := newJob(q.query)
	if err != nil {
		return nil, err
	}
	return &searchJob{child: child, patterns: q.patterns}, nil
}

type searchJob struct {
	child    job.Job
	patterns []string
}

func (j *searchJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, stream, finish := job.StartSpan(ctx, stream, j)
	defer func() { finish(alert, err) }()

	return j.child.Run(ctx, clients, stream)
}

func (j *searchJob) Name() string {
	return "KeywordSearchJob"
}

func (j *searchJob) Attributes(v job.Verbosity) (res []attribute.KeyValue) {
	switch v {
	case job.VerbosityMax:
		fallthrough
	case job.VerbosityBasic:
		res = append(res,
			attribute.StringSlice("patterns", j.patterns),
		)
	}
	return res
}

func (j *searchJob) Children() []job.Describer {
	return []job.Describer{j.child}
}

func (j *searchJob) MapChildren(fn job.MapFunc) job.Job {
	cp := *j
	cp.child = job.Map(j.child, fn)
	return &cp
}
