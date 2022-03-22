package job

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/search/job/jobutil"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

// NewSelectJob creates a job that transforms streamed results with
// the given filter.SelectPath.
func NewSelectJob(path filter.SelectPath, child Job) Job {
	return &selectJob{path: path, child: child}
}

type selectJob struct {
	path  filter.SelectPath
	child Job
}

func (j *selectJob) Run(ctx context.Context, db database.DB, stream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, stream, finish := jobutil.StartSpan(ctx, stream, j)
	defer func() { finish(alert, err) }()

	selectingStream := streaming.WithSelect(stream, j.path)
	return j.child.Run(ctx, db, selectingStream)
}

func (j *selectJob) Name() string {
	return "SelectJob"
}
