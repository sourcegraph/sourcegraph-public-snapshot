package search

import (
	"context"
	"sync"

	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/own/codeowners"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewSelectOwnersJob(child job.Job, features *search.Features) job.Job {
	return &selectOwnersJob{
		child:    child,
		features: features,
	}
}

type selectOwnersJob struct {
	child job.Job

	features *search.Features
}

func (s *selectOwnersJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	if s.features == nil || !s.features.CodeOwnershipSearch {
		return nil, &featureFlagError{predicate: "select:file.owners"}
	}

	_, ctx, stream, finish := job.StartSpan(ctx, stream, s)
	defer finish(alert, err)

	var (
		mu                    sync.Mutex
		hasResultWithNoOwners bool
		errs                  error
	)

	dedup := result.NewDeduper()
	var maxAlerter search.MaxAlerter

	rules := NewRulesCache(clients.Gitserver, clients.DB)

	filteredStream := streaming.StreamFunc(func(event streaming.SearchEvent) {
		var ok bool
		event.Results, ok, err = getCodeOwnersFromMatches(ctx, &rules, event.Results)
		if err != nil {
			mu.Lock()
			errs = errors.Append(errs, err)
			mu.Unlock()
		}
		mu.Lock()
		if ok {
			hasResultWithNoOwners = true
		}
		results := event.Results[:0]
		for _, m := range event.Results {
			if !dedup.Seen(m) {
				dedup.Add(m)
				results = append(results, m)
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
	maxAlerter.Add(alert)

	if hasResultWithNoOwners {
		maxAlerter.Add(search.AlertForUnownedResult())
	}

	return maxAlerter.Alert, errs
}

func (s *selectOwnersJob) Name() string {
	return "SelectOwnersSearchJob"
}

func (s *selectOwnersJob) Fields(_ job.Verbosity) (res []otlog.Field) {
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

func getCodeOwnersFromMatches(
	ctx context.Context,
	rules *RulesCache,
	matches []result.Match,
) ([]result.Match, bool, error) {
	var (
		errs                  error
		ownerMatches          []result.Match
		hasResultWithNoOwners bool
	)

	for _, m := range matches {
		mm, ok := m.(*result.FileMatch)
		if !ok {
			continue
		}
		rs, err := rules.GetFromCacheOrFetch(ctx, mm.Repo.Name, mm.Repo.ID, mm.CommitID)
		if err != nil {
			errs = errors.Append(errs, err)
			continue
		}
		rule := rs.Match(mm.File.Path)
		// No match.
		if rule == nil || len(rule.GetOwner()) == 0 {
			hasResultWithNoOwners = true
			continue
		}

		resolvedOwners, err := rules.ownService.ResolveOwnersWithType(ctx, rule.GetOwner())
		if err != nil {
			errs = errors.Append(errs, err)
			continue
		}

		for _, o := range resolvedOwners {
			ownerMatch := &result.OwnerMatch{
				ResolvedOwner: ownerToResult(o),
				InputRev:      mm.InputRev,
				Repo:          mm.Repo,
				CommitID:      mm.CommitID,
			}
			ownerMatches = append(ownerMatches, ownerMatch)
		}
	}
	return ownerMatches, hasResultWithNoOwners, errs
}

func ownerToResult(o codeowners.ResolvedOwner) result.Owner {
	if v, ok := o.(*codeowners.Person); ok {
		return &result.OwnerPerson{
			Handle: v.Handle,
			Email:  v.GetEmail(),
			User:   v.User,
		}
	}
	if v, ok := o.(*codeowners.Team); ok {
		return &result.OwnerTeam{
			Handle: v.Handle,
			Email:  v.Email,
			Team:   v.Team,
		}
	}
	panic("unimplemented resolved owner in ownerToResult")
}
