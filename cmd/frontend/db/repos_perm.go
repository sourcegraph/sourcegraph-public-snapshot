package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

var mockAuthzFilter func(ctx context.Context, repos []*types.Repo, p authz.Perm) ([]*types.Repo, error)

// authzFilter is the enforcement mechanism for repository permissions. It accepts a list of repositories
// and a permission type `p` and returns a subset of those repositories (no guarantee on order) for
// which the currently authenticated user has the specified permission.
func authzFilter(ctx context.Context, repos []*types.Repo, p authz.Perm) ([]*types.Repo, error) {
	if mockAuthzFilter != nil {
		return mockAuthzFilter(ctx, repos, p)
	}

	if len(repos) == 0 {
		return repos, nil
	}
	if isInternalActor(ctx) {
		return repos, nil
	}

	var currentUser *types.User
	if actor.FromContext(ctx).IsAuthenticated() {
		var err error
		currentUser, err = Users.GetByCurrentAuthUser(ctx)
		if err != nil {
			return nil, err
		}
		if currentUser.SiteAdmin {
			return repos, nil
		}
	}

	filteredRepoNames, err := getFilteredRepoNames(ctx, currentUser, authz.ToRepos(repos), p)
	if err != nil {
		return nil, err
	}

	filteredRepos := make([]*types.Repo, 0, len(filteredRepoNames))
	for _, repo := range repos {
		if _, ok := filteredRepoNames[repo.Name]; ok {
			filteredRepos = append(filteredRepos, repo)
		}
	}
	return filteredRepos, nil
}

// isInternalActor returns true if the actor represents an internal agent (i.e., non-user-bound
// request that originates from within Sourcegraph itself).
//
// 🚨 SECURITY: internal requests bypass authz provider permissions checks, so correctness is
// important here.
func isInternalActor(ctx context.Context) bool {
	return actor.FromContext(ctx).Internal
}

func getFilteredRepoNames(ctx context.Context, currentUser *types.User, repos map[authz.Repo]struct{}, p authz.Perm) (accepted map[api.RepoName]struct{}, err error) {
	var accts []*extsvc.ExternalAccount
	authzAllowByDefault, authzProviders := authz.GetProviders()
	if len(authzProviders) > 0 && currentUser != nil {
		accts, err = ExternalAccounts.List(ctx, ExternalAccountsListOptions{UserID: currentUser.ID})
		if err != nil {
			return nil, err
		}
	}

	accepted = make(map[api.RepoName]struct{})  // repositories that have been claimed and have read permissions
	unverified := make(map[authz.Repo]struct{}) // repositories that have not been claimed by any authz provider
	for repo := range repos {
		// 🚨 SECURITY: Defensively bar access to repos with no external repo spec (we don't know
		// where they came from, so can't reliably enforce permissions)
		if s := repo.ExternalRepoSpec; s.ID == "" || s.ServiceID == "" || s.ServiceType == "" {
			continue
		}
		unverified[repo] = struct{}{}
	}

	// Walk through all authz providers, checking repo permissions against each. If any own a given
	// repo, we use its permissions for that repo.
	for _, authzProvider := range authzProviders {
		if len(unverified) == 0 {
			break
		}

		// determine external account to use
		var providerAcct *extsvc.ExternalAccount
		for _, acct := range accts {
			if acct.ServiceID == authzProvider.ServiceID() && acct.ServiceType == authzProvider.ServiceType() {
				providerAcct = acct
				break
			}
		}
		if providerAcct == nil && currentUser != nil { // no existing external account for authz provider
			if pr, err := authzProvider.FetchAccount(ctx, currentUser, accts); err == nil {
				providerAcct = pr
				if providerAcct != nil {
					err := ExternalAccounts.AssociateUserAndSave(ctx, currentUser.ID, providerAcct.ExternalAccountSpec, providerAcct.ExternalAccountData)
					if err != nil {
						return nil, err
					}
				}
			} else {
				log15.Warn("Could not fetch authz provider account for user", "username", currentUser.Username, "authzProvider", authzProvider.ServiceID(), "error", err)
			}
		}

		// determine which repos "belong" to this authz provider
		myUnverified, nextUnverified := authzProvider.Repos(ctx, unverified)

		// check the perms on those repos
		perms, err := authzProvider.RepoPerms(ctx, providerAcct, myUnverified)
		if err != nil {
			return nil, err
		}
		for unverifiedRepo := range myUnverified {
			if repoPerms, ok := perms[unverifiedRepo.RepoName]; ok && repoPerms[p] {
				accepted[unverifiedRepo.RepoName] = struct{}{}
			}
		}
		// continue checking repos that didn't belong to this authz provider
		unverified = nextUnverified
	}

	if authzAllowByDefault {
		for r := range unverified {
			accepted[r.RepoName] = struct{}{}
		}
	}

	return accepted, nil
}
