package search

import (
	"context"

	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/own"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/types"
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

func (s *fileHasOwnersJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, stream, finish := job.StartSpan(ctx, stream, s)
	defer finish(alert, err)

	var maxAlerter search.MaxAlerter

	rules := NewRulesCache(clients.Gitserver, clients.DB)

	// Semantics of multiple values in includeOwners and excludeOwners is that all
	// need to match for ownership. Therefore we create a single bag per entry.
	var includeBags []own.Bag
	for _, o := range s.includeOwners {
		b := own.ByTextReference(ctx, clients.DB, o)
		includeBags = append(includeBags, b)
	}
	var excludeBags []own.Bag
	for _, o := range s.excludeOwners {
		b := own.ByTextReference(ctx, clients.DB, o)
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
		var (
			filePaths []string
			commitID  api.CommitID
			repo      types.MinimalRepo
		)
		switch mm := m.(type) {
		case *result.FileMatch:
			filePaths = []string{mm.File.Path}
			commitID = mm.CommitID
			repo = mm.Repo
		case *result.CommitMatch:
			filePaths = mm.ModifiedFiles
			commitID = mm.Commit.ID
			repo = mm.Repo
		}
		if len(filePaths) == 0 {
			continue matchesLoop
		}
		file, err := rules.GetFromCacheOrFetch(ctx, repo.Name, repo.ID, commitID)
		if err != nil {
			errs = errors.Append(errs, err)
			continue matchesLoop
		}
		// For multiple files considered for ownership in single result (CommitMatch case) we:
		// * exclude a result if none of the files is owned by all included owners,
		// * exclude a result if any of the files is owned by all excluded owners.
		var fileMatchesIncludeTerms bool
		for _, path := range filePaths {
			fileOwners := file.Match(path)
			if len(includeTerms) > 0 && ownersFilters(fileOwners, includeTerms, includeBags, false) {
				fileMatchesIncludeTerms = true
			}
			if len(excludeTerms) > 0 && !ownersFilters(fileOwners, excludeTerms, excludeBags, true) {
				continue matchesLoop
			}
		}
		if len(includeTerms) > 0 && !fileMatchesIncludeTerms {
			continue matchesLoop
		}

		filtered = append(filtered, m)
	}

	return filtered, errs
}

// ownersFilters searches within emails to determine if ownership passes filtering by searchTerms and allBags.
//   - Multiple bags have AND semantics, so ownership data needs to pass filtering criteria of each Bag.
//   - If exclude is true then we expect ownership to not be within a bag (i.e. IsWithin() is false)
//   - Empty string passed as search term means any, so the ownership is a match if there is at least one owner,
//     and false otherwise.
//   - Filtering is handled in a case-insensitive manner.
func ownersFilters(ownership fileOwnershipData, searchTerms []string, allBags []own.Bag, exclude bool) bool {
	// Empty search terms means any owner matches.
	if len(searchTerms) == 1 && searchTerms[0] == "" {
		return ownership.NonEmpty() == !exclude
	}
	for _, bag := range allBags {
		if ownership.IsWithin(bag) == exclude {
			return false
		}
	}
	return true
}
