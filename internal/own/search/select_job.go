package search

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/own"
	"github.com/sourcegraph/sourcegraph/internal/own/codeowners"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewSelectOwnersJob(child job.Job) job.Job {
	return &selectOwnersJob{
		child: child,
	}
}

type selectOwnersJob struct {
	child job.Job
}

func (s *selectOwnersJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, stream, finish := job.StartSpan(ctx, stream, s)
	defer finish(alert, err)

	var (
		mu                    sync.Mutex
		hasResultWithNoOwners bool
		maxAlerter            search.MaxAlerter
		bagMu                 sync.Mutex // TODO(#52553): Make bag thread-safe
	)

	dedup := result.NewDeduper()

	rules := NewRulesCache(clients.Gitserver, clients.DB)
	bag := own.EmptyBag()

	filteredStream := streaming.StreamFunc(func(event streaming.SearchEvent) {
		matches, ok, err := getCodeOwnersFromMatches(ctx, &rules, event.Results)
		if err != nil {
			maxAlerter.Add(search.AlertForOwnershipSearchError())
		}
		mu.Lock()
		if ok {
			hasResultWithNoOwners = true
		}
		func() {
			bagMu.Lock()
			defer bagMu.Unlock()
			for _, m := range matches {
				for _, r := range m.references {
					bag.Add(r)
				}
			}
			bag.Resolve(ctx, clients.DB)
		}()
		var results result.Matches
		for _, m := range matches {
		nextReference:
			for _, r := range m.references {
				ro, found := bag.FindResolved(r)
				if !found {
					guess := r.ResolutionGuess()
					// No text references found to make a guess, something is wrong.
					if guess == nil {
						maxAlerter.Add(search.AlertForOwnershipSearchError())
						continue nextReference
					}
					ro = guess
				}
				if ro != nil {
					om := &result.OwnerMatch{
						ResolvedOwner: ownerToResult(ro),
						InputRev:      m.fileMatch.InputRev,
						Repo:          m.fileMatch.Repo,
						CommitID:      m.fileMatch.CommitID,
					}
					if !dedup.Seen(om) {
						dedup.Add(om)
						results = append(results, om)
					}
				}
			}
		}
		event.Results = results
		mu.Unlock()
		stream.Send(event)
	})

	alert, err = s.child.Run(ctx, clients, filteredStream)
	maxAlerter.Add(alert)

	if hasResultWithNoOwners {
		maxAlerter.Add(search.AlertForUnownedResult())
	}

	return maxAlerter.Alert, err
}

func (s *selectOwnersJob) Name() string {
	return "SelectOwnersSearchJob"
}

func (s *selectOwnersJob) Attributes(_ job.Verbosity) []attribute.KeyValue { return nil }

func (s *selectOwnersJob) Children() []job.Describer {
	return []job.Describer{s.child}
}

func (s *selectOwnersJob) MapChildren(fn job.MapFunc) job.Job {
	cp := *s
	cp.child = job.Map(s.child, fn)
	return &cp
}

type ownerFileMatch struct {
	fileMatch  *result.FileMatch
	references []own.Reference
}

func getCodeOwnersFromMatches(
	ctx context.Context,
	rules *RulesCache,
	matches []result.Match,
) ([]ownerFileMatch, bool, error) {
	var (
		errs                  error
		ownerMatches          []ownerFileMatch
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
		if rule.Empty() {
			hasResultWithNoOwners = true
			continue
		}
		refs := rule.References()
		for i := range refs {
			refs[i].RepoContext = &own.RepoContext{
				Name:         mm.Repo.Name,
				CodeHostKind: rs.codeowners.GetCodeHostType(),
			}
		}

		ownerMatches = append(ownerMatches, ownerFileMatch{
			fileMatch:  mm,
			references: refs,
		})
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
