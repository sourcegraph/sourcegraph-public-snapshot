package codeownership

import (
	"context"
	"sync"

	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewSelectOwners(child job.Job) job.Job {
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
		mu           sync.Mutex
		errs         error
		hasSentAlert bool
	)
	dedup := result.NewDeduper()

	ownService := backend.NewOwnService(clients.Gitserver, clients.DB)
	rules := NewRulesCache(ownService)
	// owners := NewOwnerCache(ownService)

	filteredStream := streaming.StreamFunc(func(event streaming.SearchEvent) {
		event.Results, err = getCodeOwnersFromMatches(ctx, &rules, ownService, event.Results)
		if err != nil {
			mu.Lock()
			errs = errors.Append(errs, err)
			mu.Unlock()
		}
		mu.Lock()
		results := event.Results[:0]
		for _, m := range event.Results {
			if !dedup.Seen(m) {
				dedup.Add(m)
				results = append(results, m)
			}
		}
		event.Results = results

		if len(event.Results) == 0 && !hasSentAlert {
			hasSentAlert = true
			alert = &search.Alert{
				PrometheusType: "unowned_results_found",
				Title:          "Some results didn't have an owner",
				Description:    `To see which, run the same query with -file:has.owner() to see affected files.`,
			}
		}

		mu.Unlock()
		stream.Send(event)
	})

	// TODO: Don't drop.
	_, err = s.child.Run(ctx, clients, filteredStream)
	if err != nil {
		errs = errors.Append(errs, err)
	}
	return alert, errs
}

func (s *selectOwnersJob) Name() string {
	return "SelectOwnersSearchJob"
}

func (s *selectOwnersJob) Fields(v job.Verbosity) (res []otlog.Field) {
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
	ownService backend.OwnService,
	matches []result.Match,
) ([]result.Match, error) {
	var errs error
	var ownerMatches []result.Match

matchesLoop:
	for _, m := range matches {
		mm, ok := m.(*result.FileMatch)
		if !ok {
			continue
		}
		file, err := rules.GetFromCacheOrFetch(ctx, mm.Repo.Name, mm.CommitID)
		if err != nil {
			errs = errors.Append(errs, err)
			continue matchesLoop
		}

		resolvedOwners, err := ownService.ResolveOwnersWithType(ctx, file.FindOwners(mm.File.Path), backend.OwnerResolutionContext{
			RepoName: mm.Repo.Name,
			RepoID:   mm.Repo.ID,
		})
		if err != nil {
			errs = errors.Append(errs, err)
			continue matchesLoop
		}

		for _, o := range resolvedOwners {
			ownerMatch := &result.OwnerMatch{
				ResolvedOwner: o,
				InputRev:      mm.InputRev,
				Repo:          mm.Repo,
				CommitID:      mm.CommitID,
			}
			ownerMatches = append(ownerMatches, ownerMatch)
		}
	}
	return ownerMatches, errs
}
