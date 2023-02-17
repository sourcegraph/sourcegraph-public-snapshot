package codeownership

import (
	"context"
	"sync"

	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewDedupJob(child job.Job) job.Job {
	return &dedupOwnersJob{
		child: child,
	}
}

// dedupOwnersJob is a separate job that should be called after a SelectOwnersJob.
// The reasoning for a separate job is that we need to keep all owner matches in the first run in case we might need
// to apply sub-repo permissions filtering, but we do not want duplicate owner matches for all the filepaths they own.
type dedupOwnersJob struct {
	child job.Job
}

func (s *dedupOwnersJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, stream, finish := job.StartSpan(ctx, stream, s)
	defer finish(alert, err)

	var (
		mu   sync.Mutex
		errs error
	)
	dedup := result.NewDeduper()

	filteredStream := streaming.StreamFunc(func(event streaming.SearchEvent) {
		if err != nil {
			mu.Lock()
			errs = errors.Append(errs, err)
			mu.Unlock()
		}
		mu.Lock()
		results := event.Results[:0]
		for _, m := range event.Results {
			switch m.(type) {
			case *result.OwnerMatch:
				if !dedup.Seen(m) {
					dedup.Add(m)
					results = append(results, m)
				}
			}
		}
		event.Results = results
		mu.Unlock()
		stream.Send(event)
	})

	alert, err = s.child.Run(ctx, clients, filteredStream)
	if err != nil {
		errs = errors.Append(errs, err)
	}
	return alert, errs
}

func (s *dedupOwnersJob) Name() string {
	return "DeduplicateOwnersJob"
}

func (s *dedupOwnersJob) Fields(v job.Verbosity) (res []otlog.Field) {
	return res
}

func (s *dedupOwnersJob) Children() []job.Describer {
	return []job.Describer{s.child}
}

func (s *dedupOwnersJob) MapChildren(fn job.MapFunc) job.Job {
	cp := *s
	cp.child = job.Map(s.child, fn)
	return &cp
}
