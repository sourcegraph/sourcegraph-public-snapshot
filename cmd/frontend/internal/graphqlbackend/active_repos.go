package graphqlbackend

import (
	"context"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

var (
	inactiveRepos    = env.Get("INACTIVE_REPOS", "", "comma-separated list of repos to consider 'inactive' (e.g. while searching)")
	inactiveReposMap map[string]struct{}
)

func init() {
	// Build the map of inactive repos.
	inactiveSplit := strings.Split(inactiveRepos, ",")
	inactiveReposMap = make(map[string]struct{}, len(inactiveSplit))
	for _, r := range inactiveSplit {
		r = strings.TrimSpace(r)
		if r != "" {
			inactiveReposMap[r] = struct{}{}
		}
	}
}

type activeRepoResults struct {
	active, inactive []string
}

func (a *activeRepoResults) Active() []string {
	if a.active == nil {
		return []string{}
	}
	return a.active
}

func (a *activeRepoResults) Inactive() []string {
	if a.inactive == nil {
		return []string{}
	}
	return a.inactive
}

// ActiveRepos returns a list of active and inactive repository URIs for the
// given user.
//
// In the case of on-prem, active repos is defined as all repositories known by
// Sourcegraph minus inactive repositories (specified via $INACTIVE_REPOS).
//
// In the case of Sourcegraph.com, active repos is defined as all remote repos
// for the authenticated user minus inactive repositories (again, specified via
// $INACTIVE_REPOS).
//
func (*rootResolver) ActiveRepos(ctx context.Context) (*activeRepoResults, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Find the list of all repos (this is the union of active + inactive
	// repos, see description of this function above).
	var (
		all *sourcegraph.RepoList
		err error
	)
	if envvar.DeploymentOnPrem() {
		all, err = backend.Repos.List(ctx, &sourcegraph.RepoListOptions{
			ListOptions: sourcegraph.ListOptions{
				PerPage: 1000, // we want every repo.
			},
		})
	} else {
		all, err = backend.Repos.List(ctx, &sourcegraph.RepoListOptions{
			RemoteOnly: true, // user's repo list
			ListOptions: sourcegraph.ListOptions{
				// we want every repo for the user here too, but we use one
				// GitHub API request (which is bad because of rate limit) per
				// repo here. So we limit to 100 (this response from this
				// endpoint is cached on the browser client for 30m).
				PerPage: 100,
			},
		})
	}
	if err != nil {
		return nil, err
	}

	// Create result lists (split all.Repos into active and inactive groups).
	res := &activeRepoResults{}
	for _, r := range all.Repos {
		if _, ok := inactiveReposMap[r.URI]; ok {
			// repo is inactive
			res.inactive = append(res.inactive, r.URI)
			continue
		}
		// repo is active
		res.active = append(res.active, r.URI)
	}
	return res, nil
}
