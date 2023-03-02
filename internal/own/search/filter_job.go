package search

import (
	"context"
	"strings"
	"sync"

	otlog "github.com/opentracing/opentracing-go/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/own"
	"github.com/sourcegraph/sourcegraph/internal/own/codeowners"
	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/v1"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewFileHasOwnersJob(child job.Job, includeOwners, excludeOwners []string) job.Job {
	return &fileHasOwnersJob{
		child:         child,
		includeOwners: includeOwners,
		excludeOwners: excludeOwners,
	}
}

type fileHasOwnersJob struct {
	child job.Job

	includeOwners []string
	excludeOwners []string
}

var (
	ownService  own.Service
	cache       *Cache
	serviceOnce sync.Once
)

func (s *fileHasOwnersJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	tr, ctx, stream, finish := job.StartSpan(ctx, stream, s)
	defer finish(alert, err)

	var (
		mu   sync.Mutex
		errs error
	)

	// TODO: Very dirty hack to keep a warm cache between searches.
	serviceOnce.Do(func() {
		ownService = own.NewService(clients.Gitserver, clients.DB)
		cache = NewCache(ownService)
	})

	tr.AddEvent("StartEvaluateInput")
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
		owners, err := ownService.ResolveOwnersWithType(ctx, pbOwners, own.OwnerResolutionContext{
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
		owners, err := ownService.ResolveOwnersWithType(ctx, pbOwners, own.OwnerResolutionContext{
			// No context, only resolve Sourcegraph users for the input.
		})
		if err != nil {
			return nil, err
		}
		for _, o := range owners {
			excludeOwners.Add(o)
		}
	}
	tr.AddEvent("EndEvaluateInput")

	filteredStream := streaming.StreamFunc(func(event streaming.SearchEvent) {
		var err error
		event.Results, err = applyCodeOwnershipFiltering(ctx, cache, includeOwners.Slice(), excludeOwners.Slice(), event.Results)
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
	cache *Cache,
	includeOwners,
	excludeOwners []codeowners.ResolvedOwner,
	matches []result.Match,
) (_ []result.Match, errs error) {
	tr, ctx := trace.New(ctx, "applyCodeOwnershipFiltering", "")
	defer func() {
		tr.SetError(errs)
		tr.Finish()
	}()

	filtered := matches[:0]

	candidates := make([]codeowners.ResolvedOwner, 0, len(includeOwners)+len(excludeOwners))
	candidates = append(candidates, includeOwners...)
	candidates = append(candidates, excludeOwners...)

matchesLoop:
	for _, m := range matches {
		// Code ownership is currently only implemented for files.
		mm, ok := m.(*result.FileMatch)
		if !ok {
			continue
		}

		tr.AddEvent("Match", attribute.String("repo", string(mm.Repo.Name)))

		// Load ownership data for the file in question.
		file, err := cache.GetFromCacheOrFetch(ctx, mm.Repo.ID, mm.Repo.Name, mm.CommitID)
		if err != nil {
			errs = errors.Append(errs, err)
			continue matchesLoop
		}

		// Find the owners for the file in question and resolve the owners to
		// ResolvedOwners.
		resolvedOwners := file.FindOwnersFiltered(mm.File.Path, candidates)

		// Matching time!
		for _, owner := range includeOwners {
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
	for _, want := range owners {
		if want.Equals(owner) {
			return true
		}
	}
	return false
}
