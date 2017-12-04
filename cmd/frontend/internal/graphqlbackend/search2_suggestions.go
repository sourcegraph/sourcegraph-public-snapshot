package graphqlbackend

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"sync"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/searchquery"

	"github.com/neelance/parallel"
)

const (
	maxSearchSuggestions = 100
)

type searchSuggestionsArgs struct {
	First *int
}

func (a *searchSuggestionsArgs) applyDefaultsAndConstraints() {
	if a.First == nil || *a.First < 0 || *a.First > maxSearchSuggestions {
		n := maxSearchSuggestions
		a.First = &n
	}
}

func (r *searchResolver2) Suggestions(ctx context.Context, args *searchSuggestionsArgs) ([]*searchResultResolver, error) {
	args.applyDefaultsAndConstraints()

	if len(r.combinedQuery.Syntax.Expr) == 0 {
		return nil, nil
	}

	// Only suggest for type:file.
	typeValues, _ := r.combinedQuery.StringValues(searchquery.FieldType)
	for _, resultType := range typeValues {
		if resultType != "file" {
			return nil, nil
		}
	}

	var suggesters []func(ctx context.Context) ([]*searchResultResolver, error)

	showRepoSuggestions := func(ctx context.Context) ([]*searchResultResolver, error) {
		// * If user query contains only a single term, treat it as a repo field here and ignore the other repo queries.
		// * If only repo fields (except 1 term in user query), show repo suggestions.

		var effectiveRepoFieldValues []string
		if len(r.query.Values(searchquery.FieldDefault)) == 1 && len(r.query.Fields) == 1 {
			effectiveRepoFieldValues = append(effectiveRepoFieldValues, asString(r.query.Values(searchquery.FieldDefault)[0]))
		} else if len(r.combinedQuery.Values(searchquery.FieldRepo)) > 0 && ((len(r.combinedQuery.Values(searchquery.FieldRepoGroup)) > 0 && len(r.combinedQuery.Fields) == 2) || (len(r.combinedQuery.Values(searchquery.FieldRepoGroup)) == 0 && len(r.combinedQuery.Fields) == 1)) {
			effectiveRepoFieldValues, _ = r.combinedQuery.RegexpPatterns(searchquery.FieldRepo)
		}

		// If we have a query which is not valid, just ignore it since this is for a suggestion.
		i := 0
		for _, v := range effectiveRepoFieldValues {
			if _, err := regexp.Compile(v); err == nil {
				effectiveRepoFieldValues[i] = v
				i++
			}
		}
		effectiveRepoFieldValues = effectiveRepoFieldValues[:i]

		if len(effectiveRepoFieldValues) > 0 {
			_, _, repos, _, err := r.resolveRepositories(ctx, effectiveRepoFieldValues)
			return repos, err
		}
		return nil, nil
	}
	suggesters = append(suggesters, showRepoSuggestions)

	showFileSuggestions := func(ctx context.Context) ([]*searchResultResolver, error) {
		// If only repos/repogroups and files are specified (and at most 1 term), then show file suggestions.
		// If the user query has a file: filter AND a term, then abort; we will use showFilesWithTextMatches
		// instead.
		hasRepoOrFileFields := len(r.combinedQuery.Values(searchquery.FieldRepoGroup)) > 0 || len(r.combinedQuery.Values(searchquery.FieldRepo)) > 0 || len(r.combinedQuery.Values(searchquery.FieldFile)) > 0
		userQueryHasFileFilterAndTerm := len(r.query.Values(searchquery.FieldFile)) > 0 && len(r.query.Values(searchquery.FieldDefault)) > 0
		if hasRepoOrFileFields && len(r.combinedQuery.Values(searchquery.FieldDefault)) <= 1 && !userQueryHasFileFilterAndTerm {
			return r.resolveFiles(ctx, maxSearchSuggestions)
		}
		return nil, nil
	}
	suggesters = append(suggesters, showFileSuggestions)

	showFilesWithTextMatches := func(ctx context.Context) ([]*searchResultResolver, error) {
		cache := map[string]commitSpec{}
		getCommitSpec := func(ctx context.Context, fm *fileMatch) (commitSpec, error) {
			u, err := url.Parse(fm.uri)
			if err != nil {
				return commitSpec{}, err
			}
			key := u.Host + u.Path + "?" + u.RawQuery
			if spec, ok := cache[key]; ok {
				return spec, nil
			}
			repo, err := localstore.Repos.GetByURI(ctx, u.Host+u.Path)
			if err != nil {
				return commitSpec{}, err
			}
			rev, err := backend.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{
				Repo: repo.ID,
				Rev:  u.RawQuery,
			})
			if err != nil {
				return commitSpec{}, err
			}
			spec := commitSpec{RepoID: repo.ID, CommitID: rev.CommitID}
			cache[key] = spec
			return spec, nil
		}

		// If terms are specified, then show files that have text matches. Set an aggressive timeout
		// to avoid delaying repo and file suggestions for too long.
		ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		defer cancel()
		if len(r.combinedQuery.Values(searchquery.FieldDefault)) > 0 {
			results, err := r.doResults(ctx, "file") // only "file" result type
			if err != nil {
				if err == context.DeadlineExceeded {
					return nil, nil // don't log as error below
				}
				return nil, err
			}
			if len(results.results) > *args.First {
				results.results = results.results[:*args.First]
			}
			var suggestions []*searchResultResolver
			for i, res := range results.results {
				// TODO(sqs): should parallelize, or reuse data fetched elsewhere
				commit, err := getCommitSpec(ctx, res.fileMatch)
				if err != nil {
					if err == context.DeadlineExceeded {
						err = nil // don't log as error below
					}
					return nil, err
				}

				path := res.fileMatch.JPath
				fileResolver := &fileResolver{
					path:   path,
					name:   path,
					commit: commit,
					stat:   createFileInfo(path, false),
				}
				suggestions = append(suggestions, newSearchResultResolver(fileResolver, len(results.results)-i))
			}
			return suggestions, nil
		}
		return nil, nil
	}
	suggesters = append(suggesters, showFilesWithTextMatches)

	// Run suggesters.
	var (
		allSuggestions []*searchResultResolver
		mu             sync.Mutex
		par            = parallel.NewRun(len(suggesters))
	)
	for _, suggester := range suggesters {
		par.Acquire()
		go func(suggester func(ctx context.Context) ([]*searchResultResolver, error)) {
			defer par.Release()
			ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()
			suggestions, err := suggester(ctx)
			if err == nil {
				mu.Lock()
				allSuggestions = append(allSuggestions, suggestions...)
				mu.Unlock()
			} else {
				if err == context.DeadlineExceeded || err == context.Canceled {
					log15.Warn("search suggestions exceeded deadline (skipping)", "query", r.args.Query, "scopeQuery", r.args.ScopeQuery)
				} else if !isBadRequest(err) {
					// We exclude bad user input. Note that this means that we
					// may have some tokens in the input that are valid, but
					// typing something "bad" results in no suggestions from the
					// this suggester. In future we should just ignore the bad
					// token.
					par.Error(err)
				}
			}
		}(suggester)
	}
	if err := par.Wait(); err != nil {
		return nil, err
	}

	// Eliminate duplicates.
	type key struct {
		repoURI string
		repoID  int32
		repoRev string
		file    string
	}
	seen := make(map[key]struct{}, len(allSuggestions))
	uniqueSuggestions := allSuggestions[:0]
	for _, s := range allSuggestions {
		var k key
		switch s := s.result.(type) {
		case *repositoryResolver:
			k.repoURI = s.URI()
		case *fileResolver:
			k.repoID = s.commit.RepoID
			k.repoRev = s.commit.CommitID
			k.file = s.path
		default:
			panic(fmt.Sprintf("unhandled: %#v", s))
		}

		if _, dup := seen[k]; !dup {
			uniqueSuggestions = append(uniqueSuggestions, s)
			seen[k] = struct{}{}
		}
	}
	allSuggestions = uniqueSuggestions

	sort.Sort(searchResultSorter(allSuggestions))
	if len(allSuggestions) > *args.First {
		allSuggestions = allSuggestions[:*args.First]
	}

	return allSuggestions, nil
}
