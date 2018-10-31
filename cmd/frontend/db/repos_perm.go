package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/authz"
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

	filteredURIs, acceptAll, err := getFilteredRepoURIs(ctx, authz.ToRepos(repos), p)
	if err != nil {
		return nil, err
	}
	if acceptAll {
		return repos, nil
	}

	filteredRepos := make([]*types.Repo, 0, len(filteredURIs))
	for _, repo := range repos {
		if _, ok := filteredURIs[repo.URI]; ok {
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

func getFilteredRepoURIs(ctx context.Context, repos map[authz.Repo]struct{}, p authz.Perm) (
	accepted map[api.RepoURI]struct{}, acceptAll bool, err error,
) {
	var (
		currentUser *types.User
		accts       []*extsvc.ExternalAccount
	)
	if authz.NumAuthzProviders() > 0 && actor.FromContext(ctx).IsAuthenticated() {
		var err error
		currentUser, err = Users.GetByCurrentAuthUser(ctx)
		if err != nil {
			return nil, false, err
		}
		accts, err = ExternalAccounts.List(ctx, ExternalAccountsListOptions{
			UserID: currentUser.ID,
		})
		if err != nil {
			return nil, false, err
		}
	}

	accepted = make(map[api.RepoURI]struct{})   // repositories tha thave been claimed and have read permissions
	unverified := make(map[authz.Repo]struct{}) // repositories that have not been claimed by any authz provider
	for repo := range repos {
		unverified[repo] = struct{}{}
	}

	// Walk through all authz providers, checking repo permissions against each. If any own a given
	// repo, we use its permissions for that repo.
	err = authz.DoWithAuthzProviders(func(authzProviders []authz.AuthzProvider) error {
		for _, authzProvider := range authzProviders {
			if len(unverified) == 0 {
				break
			}

			// Determine external account to use
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
							return err
						}
					}
				} else {
					log15.Warn("Could not fetch authz provider account for user", "username", currentUser.Username, "authzProvider", authzProvider.ServiceID(), "error", err)
				}
			}

			perms, err := authzProvider.RepoPerms(ctx, providerAcct, unverified)
			if err != nil {
				return err
			}

			newUnverified := make(map[authz.Repo]struct{})
			for unverifiedRepo := range unverified {
				repoPerms, ok := perms[unverifiedRepo.URI]
				if !ok {
					newUnverified[unverifiedRepo] = struct{}{}
					continue
				}
				if repoPerms[p] {
					accepted[unverifiedRepo.URI] = struct{}{}
				}
			}
			unverified = newUnverified
		}
		return nil
	})
	if err != nil {
		return nil, false, err
	}

	if authz.AllowByDefault() {
		for r := range unverified {
			accepted[r.URI] = struct{}{}
		}
	}

	return accepted, false, nil
}
