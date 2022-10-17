package codeownership

import (
	"context"
	"sync"

	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func New(child job.Job, includeOwners, excludeOwners []string) job.Job {
	return &codeownershipJob{
		child:         child,
		includeOwners: includeOwners,
		excludeOwners: excludeOwners,
	}
}

type codeownershipJob struct {
	child job.Job

	includeOwners []string
	excludeOwners []string
}

func (s *codeownershipJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, stream, finish := job.StartSpan(ctx, stream, s)
	defer func() { finish(alert, err) }()

	var (
		mu   sync.Mutex
		errs error
	)

	rules := NewRulesCache()

	filteredStream := streaming.StreamFunc(func(event streaming.SearchEvent) {
		var err error
		event.Results, err = applyCodeOwnershipFiltering(ctx, clients.Gitserver, &rules, s.includeOwners, s.excludeOwners, event.Results)
		if err != nil {
			mu.Lock()
			errs = errors.Append(errs, err)
			mu.Unlock()
		}
		stream.Send(event)
	})

	alert, err = s.child.Run(ctx, clients, filteredStream)
	if err != nil {
		errs = errors.Append(errs, err)
	}
	return alert, errs
}

func (s *codeownershipJob) Name() string {
	return "CodeOwnershipFilterJob"
}

func (s *codeownershipJob) Fields(v job.Verbosity) (res []otlog.Field) {
	switch v {
	case job.VerbosityMax:
		fallthrough
	case job.VerbosityBasic:
		res = append(res,
			trace.Strings("includeOwners", s.includeOwners),
			trace.Strings("excludeOwners", s.excludeOwners),
		)
	}
	return res
}

func (s *codeownershipJob) Children() []job.Describer {
	return []job.Describer{s.child}
}

func (s *codeownershipJob) MapChildren(fn job.MapFunc) job.Job {
	cp := *s
	cp.child = job.Map(s.child, fn)
	return &cp
}

func applyCodeOwnershipFiltering(
	ctx context.Context,
	gitserver gitserver.Client,
	rules *RulesCache,
	includeOwners,
	excludeOwners []string,
	matches []result.Match) ([]result.Match, error) {
	var errs error

	filtered := matches[:0]

matchesLoop:
	for _, m := range matches {
		// Code ownership is currently only implemented for files.
		mm, ok := m.(*result.FileMatch)
		if !ok {
			continue
		}

		ruleset, err := rules.GetFromCacheOrFetch(ctx, gitserver, mm.Repo.Name, mm.CommitID)
		if err != nil {
			errs = errors.Append(errs, err)
		}

		var owners Owners
		owners, err = ruleset.Match(mm.File.Path)
		if err != nil {
			errs = errors.Append(errs, err)
		}

		for _, owner := range includeOwners {
			if !containsOwner(owners, owner) {
				continue matchesLoop
			}
		}

		filtered = append(filtered, m)
	}

	return filtered, errs
}

func containsOwner(owners Owners, owner string) bool {
	for _, o := range owners {
		if o.String() == owner {
			return true
		}
	}
	return false
}
