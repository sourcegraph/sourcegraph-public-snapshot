package search

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/own"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
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

	var maxAlerter search.MaxAlerter

	rules := NewRulesCache(clients.Gitserver, clients.DB)

	// Semantics of multiple values in includeOwners and excludeOwners is that all
	// need to match for ownership. Therefore we create a single bag per entry.
	var includeBags []own.Bag
	for _, o := range s.includeOwners {
		b := own.ByTextReference(ctx, database.NewEnterpriseDB(clients.DB), o)
		includeBags = append(includeBags, b)
	}
	var excludeBags []own.Bag
	for _, o := range s.excludeOwners {
		b := own.ByTextReference(ctx, database.NewEnterpriseDB(clients.DB), o)
		excludeBags = append(excludeBags, b)
	}

	filteredStream := streaming.StreamFunc(func(event streaming.SearchEvent) {
		var err error
		event.Results, err = applyCodeOwnershipFiltering(ctx, &rules, includeBags, s.includeOwners, excludeBags, s.excludeOwners, event.Results)
		if err != nil {
			maxAlerter.Add(search.AlertForOwnershipSearchError())
		}
		stream.Send(event)
	})

	alert, err = s.child.Run(ctx, clients, filteredStream)
	// Add is nil-safe, we can just add an alert even if its pointer is nil.
	maxAlerter.Add(alert)
	return maxAlerter.Alert, err
}

func (s *fileHasOwnersJob) Name() string {
	return "FileHasOwnersFilterJob"
}

func (s *fileHasOwnersJob) Attributes(v job.Verbosity) (res []attribute.KeyValue) {
	switch v {
	case job.VerbosityMax:
		fallthrough
	case job.VerbosityBasic:
		res = append(res,
			attribute.StringSlice("includeOwners", s.includeOwners),
			attribute.StringSlice("excludeOwners", s.excludeOwners),
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
	includeBags []own.Bag,
	includeTerms []string,
	excludeBags []own.Bag,
	excludeTerms []string,
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
		fileOwners := file.Match(mm.File.Path)
		if len(includeTerms) > 0 && !containsOwner(fileOwners, includeTerms, includeBags) {
			continue matchesLoop
		}
		if len(excludeTerms) > 0 && containsOwner(fileOwners, excludeTerms, excludeBags) {
			continue matchesLoop
		}

		filtered = append(filtered, m)
	}

	return filtered, errs
}

// containsOwner searches within emails and handles in a case-insensitive
// manner.
//
//   - Empty string passed as search term means any, so the predicate
//     returns true if there is at least one owner, and false otherwise.
//   - Multiple bags have AND semantics, so ownership data needs to be within
//     all of the search term bags.
func containsOwner(ownership fileOwnershipData, searchTerms []string, allBags []own.Bag) bool {
	// Empty search terms means any owner matches.
	if len(searchTerms) == 1 && searchTerms[0] == "" {
		return ownership.NonEmpty()
	}
	for _, bag := range allBags {
		if !ownership.IsWithin(bag) {
			return false
		}
	}
	return true
}
