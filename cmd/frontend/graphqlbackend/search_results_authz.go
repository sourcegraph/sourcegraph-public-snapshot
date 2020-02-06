package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

// 🚨 SECURITY: This filterer enforces the integrity of authz checks against search results when
// post-filtering is enabled.
type authzSearchResultFilterer struct {
	// Inputs

	Ctx     context.Context
	Results []SearchResultResolver

	// Outputs

	// FilteredResults contains the list of results after filtering.
	FilteredResults []SearchResultResolver
	// FilteredRepoNames is the set of repos in the filtered set.
	FilteredRepoNames map[string]struct{}
}

func newAuthzSearchResultFilterer(ctx context.Context, results []SearchResultResolver) *authzSearchResultFilterer {
	return &authzSearchResultFilterer{
		Ctx:               ctx,
		Results:           results,
		FilteredRepoNames: map[string]struct{}{},
	}
}

func (f *authzSearchResultFilterer) Len() int { return len(f.Results) }

func (f *authzSearchResultFilterer) RepoNameForElem(i int) string {
	repoName, _ := f.Results[i].searchResultURIs()
	return repoName
}

func (f *authzSearchResultFilterer) Select(i int) {
	f.FilteredResults = append(f.FilteredResults, f.Results[i])
	repoName, _ := f.Results[i].searchResultURIs()
	f.FilteredRepoNames[repoName] = struct{}{}
}

func (f *authzSearchResultFilterer) FilterRepoNames(repoNames map[string]struct{}) map[string]struct{} {
	ctx := authz.WithBypassPermissionsCheck(f.Ctx, false)
	filteredRepoNames := make(map[string]struct{})
	for repoName := range repoNames {
		if repoName == "" {
			continue
		}
		if _, err := backend.Repos.GetByName(ctx, api.RepoName(repoName)); err == nil {
			filteredRepoNames[repoName] = struct{}{}
		}
	}
	return filteredRepoNames
}

type authzRepoSetFilterer struct {
	KeepRepoNames map[string]struct{}
	Repos         []*types.Repo

	FilteredRepos []*types.Repo
}

func (f *authzRepoSetFilterer) Len() int { return len(f.Repos) }
func (f *authzRepoSetFilterer) RepoNameForElem(i int) string {
	return string(f.Repos[i].Name)
}
func (f *authzRepoSetFilterer) Select(i int) {
	f.FilteredRepos = append(f.FilteredRepos, f.Repos[i])
}
func (f *authzRepoSetFilterer) FilterRepoNames(repoNames map[string]struct{}) map[string]struct{} {
	return f.KeepRepoNames
}

func filterRepos(repos []*types.Repo, keepRepoNames map[string]struct{}) []*types.Repo {
	repoFilterer := authzRepoSetFilterer{
		KeepRepoNames: keepRepoNames,
		Repos:         repos,
	}
	authz.Filter(&repoFilterer)
	return repoFilterer.FilteredRepos
}
