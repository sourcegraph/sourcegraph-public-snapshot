package job

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
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

func (j *selectJob) Run(ctx context.Context, stream streaming.Sender) (*search.Alert, error) {
	selectingStream := streaming.WithSelect(stream, j.path)
	return j.child.Run(ctx, selectingStream)
}

func (j *selectJob) Name() string {
	return "Select"
}
