package jobutil

import (
	"context"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

func NewLogJob(child job.Job) job.Job {
	return &LogJob{
		child: child,
	}
}

type LogJob struct {
	child job.Job
}

func (l *LogJob) Run(ctx context.Context, clients job.RuntimeClients, s streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, s, finish := job.StartSpan(ctx, s, l)
	defer func() { finish(alert, err) }()

	return l.child.Run(ctx, clients, s)
}

func (l *LogJob) Name() string {
	return "LogJob"
}

func (l *LogJob) Fields(v job.Verbosity) (res []log.Field) { return nil }

func (l *LogJob) Children() []job.Describer {
	return []job.Describer{l.child}
}

func (l *LogJob) MapChildren(fn job.MapFunc) job.Job {
	cp := *l
	cp.child = job.Map(l.child, fn)
	return &cp
}
