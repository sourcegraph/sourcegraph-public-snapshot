package jobutil

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
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

	selectingStream := newSelectingStream(stream, j.path)
	return j.child.Run(ctx, clients, selectingStream)
}

func (j *selectJob) Name() string {
	return "SelectJob"
}
func (j *selectJob) Attributes(v job.Verbosity) (res []attribute.KeyValue) {
	switch v {
	case job.VerbosityMax:
		fallthrough
	case job.VerbosityBasic:
		res = append(res,
			attribute.StringSlice("select", j.path),
		)
	}
	return res
}

func (j *selectJob) Children() []job.Describer {
	return []job.Describer{j.child}
}

func (j *selectJob) MapChildren(fn job.MapFunc) job.Job {
	cp := *j
	cp.child = job.Map(j.child, fn)
	return &cp
}

// newSelectingStream returns a child Stream of parent that runs the select operation
// on each event, deduplicating where possible.
func newSelectingStream(parent streaming.Sender, s filter.SelectPath) streaming.Sender {
	var mux sync.Mutex
	dedup := result.NewDeduper()

	return streaming.StreamFunc(func(e streaming.SearchEvent) {
		mux.Lock()

		selected := e.Results[:0]
		for _, match := range e.Results {
			current := match.Select(s)
			if current == nil {
				continue
			}

			// If the selected file is a file match send it unconditionally
			// to ensure we get all line matches for a file. One exception:
			// if we are only interested in the path (via `select:file`),
			// we only send the result once.
			seen := dedup.Seen(current)
			fm, isFileMatch := current.(*result.FileMatch)
			if seen && !isFileMatch {
				continue
			}
			if seen && isFileMatch && fm.IsPathMatch() {
				continue
			}

			dedup.Add(current)
			selected = append(selected, current)
		}
		e.Results = selected

		mux.Unlock()
		parent.Send(e)
	})
}
