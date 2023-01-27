package codeownership

import (
	"context"
	"fmt"

	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewSelectOwnersSearch(child job.Job, owning string) job.Job {
	return &selectOwnersJob{
		child:  child,
		owning: owning,
	}
}

type selectOwnersJob struct {
	child job.Job

	owning string
}

func (s *selectOwnersJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, stream, finish := job.StartSpan(ctx, stream, s)
	defer finish(alert, err)

	var (
		errs error
	)

	_ = NewRulesCache()

	filteredStream := streaming.StreamFunc(func(event streaming.SearchEvent) {
		fmt.Println("event is sent")
		stream.Send(event)
	})

	alert, err = s.child.Run(ctx, clients, filteredStream)
	if err != nil {
		errs = errors.Append(errs, err)
	}
	return alert, errs
}

func (s *selectOwnersJob) Name() string {
	return "SelectOwnersSearchJob"
}

func (s *selectOwnersJob) Fields(v job.Verbosity) (res []otlog.Field) {
	switch v {
	case job.VerbosityMax:
		fallthrough
	case job.VerbosityBasic:
		res = append(res,
			trace.Strings("owning", []string{s.owning}),
		)
	}
	return res
}

func (s *selectOwnersJob) Children() []job.Describer {
	return []job.Describer{s.child}
}

func (s *selectOwnersJob) MapChildren(fn job.MapFunc) job.Job {
	cp := *s
	cp.child = job.Map(s.child, fn)
	return &cp
}
