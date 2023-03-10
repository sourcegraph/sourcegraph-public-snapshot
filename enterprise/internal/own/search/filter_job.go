package search

import (
	"context"
	"fmt"
	"strings"
	"sync"

	otlog "github.com/opentracing/opentracing-go/log"

	codeownerspb "github.com/sourcegraph/sourcegraph/enterprise/internal/own/codeowners/v1"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type featureFlagError struct {
	predicate string
}

func (e *featureFlagError) Error() string {
	return fmt.Sprintf("`%s` searches are not enabled on this instance. <a href=\"/help/own\">Learn more about Own.</a>", e.predicate)
}

func NewFileHasOwnersJob(child job.Job, features *search.Features, includeOwners, excludeOwners []string) job.Job {
	return &fileHasOwnersJob{
		child:         child,
		features:      features,
		includeOwners: includeOwners,
		excludeOwners: excludeOwners,
	}
}

type fileHasOwnersJob struct {
	child    job.Job
	features *search.Features

	includeOwners []string
	excludeOwners []string
}

func (s *fileHasOwnersJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	if s.features == nil || !s.features.CodeOwnershipSearch {
		return nil, &featureFlagError{predicate: "file:has.owner()"}
	}
	_, ctx, stream, finish := job.StartSpan(ctx, stream, s)
	defer finish(alert, err)

	var (
		mu   sync.Mutex
		errs error
	)

	rules := NewRulesCache(clients.Gitserver, clients.DB)

	filteredStream := streaming.StreamFunc(func(event streaming.SearchEvent) {
		var err error
		event.Results, err = applyCodeOwnershipFiltering(ctx, &rules, s.includeOwners, s.excludeOwners, event.Results)
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

func (s *fileHasOwnersJob) Name() string {
	return "FileHasOwnersFilterJob"
}

func (s *fileHasOwnersJob) Fields(v job.Verbosity) (res []otlog.Field) {
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

func (s *fileHasOwnersJob) Children() []job.Describer {
	return []job.Describer{s.child}
}

func (s *fileHasOwnersJob) MapChildren(fn job.MapFunc) job.Job {
	cp := *s
	cp.child = job.Map(s.child, fn)
	return &cp
}

func applyCodeOwnershipFiltering(
	ctx context.Context,
	rules *RulesCache,
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

		file, err := rules.GetFromCacheOrFetch(ctx, mm.Repo.Name, mm.Repo.ID, mm.CommitID)
		if err != nil {
			errs = errors.Append(errs, err)
			continue matchesLoop
		}
		rule := file.Match(mm.File.Path)
		var owners []*codeownerspb.Owner
		// If match.
		if rule != nil {
			owners = rule.GetOwner()
		}
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

		filtered = append(filtered, m)
	}

	return filtered, errs
}

// containsOwner searches within emails and handles in a case-insensitive
// manner. Empty string passed as search term means any, so the predicate
// returns true if there is at least one owner, and false otherwise.
func containsOwner(owners []*codeownerspb.Owner, owner string) bool {
	if owner == "" {
		return len(owners) > 0
	}
	isHandle := strings.HasPrefix(owner, "@")
	owner = strings.ToLower(strings.TrimPrefix(owner, "@"))
	for _, o := range owners {
		if strings.ToLower(o.Handle) == owner {
			return true
		}
		// Prefixing the search term with `@` indicates intent to match a handle,
		// so we do not match email in that case.
		if !isHandle && (strings.ToLower(o.Email) == owner) {
			return true
		}
	}
	return false
}
