package codeownership

import (
	"context"
	"fmt"
	"strings"
	"sync"

	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/proto"
)

func New(child job.Job, onlyOwned, onlyUnowned bool, includeOwners, excludeOwners []string) job.Job {
	return &codeownershipJob{
		child:         child,
		onlyOwned:     onlyOwned,
		onlyUnowned:   onlyUnowned,
		includeOwners: includeOwners,
		excludeOwners: excludeOwners,
	}
}

type codeownershipJob struct {
	child job.Job

	includeOwners []string
	excludeOwners []string
	onlyOwned     bool
	onlyUnowned   bool
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
		event.Results, err = applyCodeOwnershipFiltering(ctx, clients.Gitserver, &rules, s.onlyOwned, s.onlyUnowned, s.includeOwners, s.excludeOwners, event.Results)
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
			trace.Strings("onlyOwned", []string{fmt.Sprintf("%t", s.onlyOwned)}),
			trace.Strings("onlyUnowned", []string{fmt.Sprintf("%t", s.onlyUnowned)}),
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
	onlyOwned,
	onlyUnowned bool,
	includeOwners,
	excludeOwners []string,
	matches []result.Match,
) ([]result.Match, error) {
	var errs error

	filtered := matches[:0]

matchesLoop:
	for _, m := range matches {
		// Code ownership is currently only implemented for files.
		mm, ok := m.(*result.FileMatch)
		if !ok {
			continue
		}

		file, err := rules.GetFromCacheOrFetch(ctx, gitserver, mm.Repo.Name, mm.CommitID)
		if err != nil {
			errs = errors.Append(errs, err)
		}
		owners := file.FindOwners(mm.File.Path)
		for _, owner := range includeOwners {
			if !containsOwner(owners, owner) {
				continue matchesLoop
			}
		}
		for _, notOwner := range excludeOwners {
			if containsOwner(owners, notOwner) {
				continue matchesLoop
			}
		}
		if onlyOwned && len(owners) == 0 {
			continue matchesLoop
		}
		if onlyUnowned && len(owners) > 0 {
			continue matchesLoop
		}

		filtered = append(filtered, m)
	}

	return filtered, errs
}

func containsOwner(owners []*codeownerspb.Owner, owner string) bool {
	ownerHasAtPrefix := strings.HasPrefix(owner, "@")
	var ownerWithoutAtPrefix string
	if ownerHasAtPrefix {
		ownerWithoutAtPrefix = strings.TrimPrefix(owner, "@")
	}
	for _, o := range owners {
		// todo: should we match case-insensitive?
		if ownerHasAtPrefix && o.Handle == ownerWithoutAtPrefix {
			return true
		}
		if o.Email == owner {
			return true
		}
	}
	return false
}
