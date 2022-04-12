package jobutil

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

// NewSelectJob creates a job that transforms streamed results with
// the given filter.SelectPath.
func NewSelectJob(path filter.SelectPath, child job.Job) job.Job {
	return &selectJob{path: path, child: child}
}

type selectJob struct {
	path  filter.SelectPath
	child job.Job
}

func (j *selectJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, stream, finish := job.StartSpan(ctx, stream, j)
	defer func() { finish(alert, err) }()

	selectingStream := streaming.WithSelect(stream, j.path)
	return j.child.Run(ctx, clients, selectingStream)
}

func (j *selectJob) Name() string {
	return "SelectJob"
}
