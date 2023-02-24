package codeownership

import (
	"context"
	"strings"
	"sync"

	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/own/codeowners"
	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/v1"
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
	defer finish(alert, err)

	var (
		mu   sync.Mutex
		errs error
	)

	ownService := backend.NewOwnService(clients.Gitserver, clients.DB)
	rules := NewRulesCache(ownService)
	// owners := NewOwnerCache(ownService)

	// Resolve input strings to ResolvedOwners so we can match them.
	var (
		includeOwners = make(codeowners.ResolvedOwners, len(s.includeOwners))
		excludeOwners = make(codeowners.ResolvedOwners, len(s.excludeOwners))
	)
	if len(s.includeOwners) > 0 {
		pbOwners := make([]*codeownerspb.Owner, 0, len(s.includeOwners))
		for _, o := range s.includeOwners {
			if o == "" {
				includeOwners.Add(&codeowners.Any{})
				continue
			}
			pbOwners = append(pbOwners, codeowners.ParseOwner(strings.ToLower(o)))
		}
		owners, err := ownService.ResolveOwnersWithType(ctx, pbOwners, backend.OwnerResolutionContext{
			// No context, only resolve Sourcegraph users for the input.
		})
		if err != nil {
			return nil, err
		}
		for _, o := range owners {
			includeOwners.Add(o)
		}
	}
	if len(s.excludeOwners) > 0 {
		pbOwners := make([]*codeownerspb.Owner, 0, len(s.excludeOwners))
		for _, o := range s.excludeOwners {
			if o == "" {
				excludeOwners.Add(&codeowners.Any{})
				continue
			}
			pbOwners = append(pbOwners, codeowners.ParseOwner(strings.ToLower(o)))
		}
		owners, err := ownService.ResolveOwnersWithType(ctx, pbOwners, backend.OwnerResolutionContext{
			// No context, only resolve Sourcegraph users for the input.
		})
		if err != nil {
			return nil, err
		}
		for _, o := range owners {
			excludeOwners.Add(o)
		}
	}

	// fmt.Printf("Resolved Owners input: %#+v\n", includeOwners)

	filteredStream := streaming.StreamFunc(func(event streaming.SearchEvent) {
		var err error
		event.Results, err = applyCodeOwnershipFiltering(ctx, &rules, ownService, includeOwners, excludeOwners, event.Results)
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
	rules *RulesCache,
	// ownersCache *OwnerCache,
	ownService backend.OwnService,
	includeOwners,
	excludeOwners codeowners.ResolvedOwners,
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

		// Load ownership data for the file in question.
		file, err := rules.GetFromCacheOrFetch(ctx, mm.Repo.Name, mm.CommitID)
		if err != nil {
			errs = errors.Append(errs, err)
			continue matchesLoop
		}

		// Find the owners for the file in question and resolve the owners to
		// ResolvedOwners.
		resolvedOwners, err := ownService.ResolveOwnersWithType(
			ctx,
			file.FindOwners(mm.File.Path),
			backend.OwnerResolutionContext{
				RepoID:   mm.Repo.ID,
				RepoName: mm.Repo.Name,
			},
		)
		if err != nil {
			errs = errors.Append(errs, err)
			continue matchesLoop
		}

		// fmt.Printf("Resolved Owners output: %#+v\n", resolvedOwners)

		// Matching time!
		for _, owner := range includeOwners {
			// TODO: This doesn't work anymore since I added teams.
			// if a team doesn't match, this returns, because includeOwners is
			// AND. We want it to be OR though for [user, ...userTeams].
			if !containsOwner(resolvedOwners, owner) {
				continue matchesLoop
			}

			// Even more todo: this changes to OR now from AND
			// if containsOwner(resolvedOwners, owner) {
			// 	filtered = append(filtered, m)
			// 	continue matchesLoop
			// }
		}
		for _, notOwner := range excludeOwners {
			// Even more todo: this changes to OR now from AND
			if containsOwner(resolvedOwners, notOwner) {
				continue matchesLoop
			}
			// if !containsOwner(resolvedOwners, notOwner) {
			// 	filtered = append(filtered, m)
			// 	continue matchesLoop
			// }
		}
		// Even more todo: this changes to OR now from AND
		// filtered = append(filtered, m)
		filtered = append(filtered, m)
	}

	return filtered, errs
}

// containsOwner searches within emails and handles in a case-insensitive
// manner. Empty string passed as search term means any, so the predicate
// returns true if there is at least one owner, and false otherwise.
func containsOwner(owners codeowners.ResolvedOwners, owner codeowners.ResolvedOwner) bool {
	if len(owners) == 0 {
		_, ok := owner.(*codeowners.Any)
		return ok
	}
	for _, want := range owners {
		if want.Equals(owner) {
			return true
		}
	}
	return false
}
