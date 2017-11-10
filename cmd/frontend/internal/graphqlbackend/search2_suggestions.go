package graphqlbackend

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"sync"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"

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

	if len(r.combinedQuery.tokens) == 0 {
		return nil, nil
	}

	var suggesters []func(ctx context.Context) ([]*searchResultResolver, error)

	showRepoSuggestions := func(ctx context.Context) ([]*searchResultResolver, error) {
		// * If user query contains only a single term, treat it as a repo field here and ignore the other repo queries.
		// * If only repo fields (except 1 term in user query), show repo suggestions.

		var effectiveRepoFieldValues []string
		if len(r.query.fieldValues[searchFieldTerm]) == 1 && len(r.query.fieldValues) == 1 {
			effectiveRepoFieldValues = append(effectiveRepoFieldValues, r.query.fieldValues[searchFieldTerm][0].Value)
		} else if len(r.combinedQuery.fieldValues[searchFieldRepo]) > 0 && ((len(r.combinedQuery.fieldValues[searchFieldRepoGroup]) > 0 && len(r.combinedQuery.fieldValues) == 2) || (len(r.combinedQuery.fieldValues[searchFieldRepoGroup]) == 0 && len(r.combinedQuery.fieldValues) == 1)) {
			effectiveRepoFieldValues = r.combinedQuery.fieldValues[searchFieldRepo].Values()
		}

		if len(effectiveRepoFieldValues) > 0 {
			_, repos, _, err := r.resolveRepositories(ctx, effectiveRepoFieldValues)
			return repos, err
		}
		return nil, nil
	}
	suggesters = append(suggesters, showRepoSuggestions)

	showFileSuggestions := func(ctx context.Context) ([]*searchResultResolver, error) {
		// If only repos/repogroups and files are specified (and at most 1 term), then show file suggestions.
		// If the user query has a file: filter AND a term, then abort; we will use showFilesWithTextMatches
		// instead.
		hasRepoOrFileFields := len(r.combinedQuery.fieldValues[searchFieldRepoGroup]) > 0 || len(r.combinedQuery.fieldValues[searchFieldRepo]) > 0 || len(r.combinedQuery.fieldValues[searchFieldFile]) > 0
		userQueryHasFileFilterAndTerm := len(r.query.fieldValues[searchFieldFile]) > 0 && len(r.query.fieldValues[""]) > 0
		if hasRepoOrFileFields && len(r.combinedQuery.fieldValues[""]) <= 1 && !userQueryHasFileFilterAndTerm {
			return r.resolveFiles(ctx)
		}
		return nil, nil
	}
	suggesters = append(suggesters, showFileSuggestions)

	showFilesWithTextMatches := func(ctx context.Context) ([]*searchResultResolver, error) {
		// If terms are specified, then show files that have text matches. Set an aggressive timeout
		// to avoid delaying repo and file suggestions for too long.
		ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		defer cancel()
		if len(r.combinedQuery.fieldValues[""]) > 0 {
			results, err := r.Results(ctx)
			if err != nil {
				if err == context.DeadlineExceeded {
					err = nil // don't log as error below
				}
				return nil, err
			}
			if len(results.results) > *args.First {
				results.results = results.results[:*args.First]
			}
			var suggestions []*searchResultResolver
			for i, res := range results.results {
				// TODO(sqs): inefficient
				u, err := url.Parse(res.uri)
				if err != nil {
					return nil, err
				}
				uri := u.Host + u.Path
				repo, err := localstore.Repos.GetByURI(ctx, uri)
				if err != nil {
					if err == context.DeadlineExceeded {
						err = nil // don't log as error below
					}
					return nil, err
				}

				path := res.JPath
				fileResolver := &fileResolver{
					path:   path,
					name:   path,
					commit: commitSpec{DefaultBranch: repo.DefaultBranch, RepoID: repo.ID},
					stat:   createFileInfo(path, false),
					repo:   &repositoryResolver{repo: repo},
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
				} else {
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
