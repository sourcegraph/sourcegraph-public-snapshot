package codeownership

import (
	"context"
	"fmt"
	"strings"
	"sync"

	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/database"
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
	owners := NewOwnerCache(ownService)

	// Resolve input strings to ResolvedOwners so we can match them.
	var (
		includeOwners []codeowners.ResolvedOwner
		excludeOwners []codeowners.ResolvedOwner
	)
	for _, o := range s.includeOwners {
		if o == "" {
			includeOwners = append(includeOwners, &codeowners.Person{Handle: matchesAllOwner})
			continue
		}
		owners, err := ownService.ResolveOwnersWithType(ctx, []*codeownerspb.Owner{codeowners.ParseOwner(strings.ToLower(o))})
		if err != nil {
			return nil, err
		}
		if len(owners) == 1 {
			// Append teams of a user, if any.
			if person, ok := owners[0].(*codeowners.Person); ok {
				if person.User != nil {
					teams, _, err := clients.DB.Teams().ListTeams(ctx, database.ListTeamsOpts{ForUserMember: person.User.ID})
					if err != nil {
						return nil, err
					}
					for _, team := range teams {
						includeOwners = append(includeOwners, &codeowners.Team{Handle: team.Name, Team: team})
					}
				}
			}
			includeOwners = append(includeOwners, owners[0])
		}
	}
	for _, o := range s.excludeOwners {
		if o == "" {
			excludeOwners = append(includeOwners, &codeowners.Person{Handle: matchesAllOwner})
			continue
		}
		owners, err := ownService.ResolveOwnersWithType(ctx, []*codeownerspb.Owner{codeowners.ParseOwner(strings.ToLower(o))})
		if err != nil {
			return nil, err
		}
		if len(owners) == 1 {
			// Append teams of a user, if any.
			if person, ok := owners[0].(*codeowners.Person); ok {
				if person.User != nil {
					teams, _, err := clients.DB.Teams().ListTeams(ctx, database.ListTeamsOpts{ForUserMember: person.User.ID})
					if err != nil {
						return nil, err
					}
					for _, team := range teams {
						includeOwners = append(includeOwners, &codeowners.Team{Handle: team.Name, Team: team})
					}
				}
			}
			excludeOwners = append(excludeOwners, owners[0])
		}
	}

	for _, o := range includeOwners {
		fmt.Printf("Include Owners #%+v\n", o)
	}

	filteredStream := streaming.StreamFunc(func(event streaming.SearchEvent) {
		var err error
		event.Results, err = applyCodeOwnershipFiltering(ctx, &rules, &owners, includeOwners, excludeOwners, event.Results)
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
	ownersCache *OwnerCache,
	includeOwners,
	excludeOwners []codeowners.ResolvedOwner,
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

		file, err := rules.GetFromCacheOrFetch(ctx, mm.Repo.Name, mm.CommitID)
		if err != nil {
			errs = errors.Append(errs, err)
			continue matchesLoop
		}
		owners := file.FindOwners(mm.File.Path)
		resolvedOwners := make([]codeowners.ResolvedOwner, 0, len(owners))
		for _, owner := range owners {
			resolvedOwner, err := ownersCache.GetFromCacheOrFetch(ctx, mm.Repo.Name, owner)
			if err != nil {
				errs = errors.Append(errs, err)
				continue matchesLoop
			}
			resolvedOwners = append(resolvedOwners, resolvedOwner)
		}
		for _, owner := range includeOwners {
			// TODO: This doesn't work anymore since I added teams.
			// if a team doesn't match, this returns, because includeOwners is
			// AND. We want it to be OR though for [user, ...userTeams].
			if !containsOwner(resolvedOwners, owner) {
				continue matchesLoop
			}
		}
		for _, notOwner := range excludeOwners {
			if containsOwner(resolvedOwners, notOwner) {
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
func containsOwner(owners []codeowners.ResolvedOwner, owner codeowners.ResolvedOwner) bool {
	fmt.Printf("containsOwner %d\n", len(owners))
	if owner.Identifier() == matchesAllOwner {
		return len(owners) > 0
	}
	for _, o := range owners {
		if equalOwners(o, owner) {
			return true
		}
	}
	return false
}

const matchesAllOwner = "owner-superlongrandomstringthatnooneshoulduse"

func equalOwners(a, b codeowners.ResolvedOwner) (ok bool) {
	defer func() {
		fmt.Printf("Comparing owners %#+v %#+v: %t\n", a, b, ok)
	}()
	if a.Identifier() == matchesAllOwner || b.Identifier() == matchesAllOwner {
		return true
	}
	return a.Identifier() == b.Identifier()
}
