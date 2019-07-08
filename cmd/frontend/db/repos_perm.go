package db

import (
	"context"

	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
	"github.com/sourcegraph/sourcegraph/pkg/trace"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

var mockAuthzFilter func(ctx context.Context, repos []*types.Repo, p authz.Perm) ([]*types.Repo, error)

// authzFilter is the enforcement mechanism for repository permissions. It is the root
// repository-permission-enforcing function (i.e., all other code that wants to check/enforce
// permissions and is not itself part of the permission-checking code should call this function).
//
// It accepts a list of repositories and a permission type `p` and returns a subset of those
// repositories (no guarantee on order) for which the currently authenticated user has the specified
// permission.
//
// The enforcement policy:
//
// - If there are no authz providers and `authzAllowByDefault` is true, then the repository is
//   accessible to everyone.
//
// - Otherwise, each repository must have an external repo spec. If a repo doesn't have one, we
//   cannot definitively associate the repository with an authz provider, and therefore we
//   *never* return the repository.
//
// - Scan through the list of authz providers until we find one that matches the repository. Return
//   whether or not the repository accessible according to that authz provider.
//
// - If no authz providers match the repository, consult `authzAllowByDefault`. If true, then return
//   the repository; otherwise, do not.
func authzFilter(ctx context.Context, repos []*types.Repo, p authz.Perm) (rs []*types.Repo, err error) {
	tr, ctx := trace.New(ctx, "authzFilter", "")
	defer func() {
		if err != nil {
			tr.SetError(err)
		}
		tr.Finish()
	}()

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
	tr, ctx := trace.New(ctx, "getFilteredRepoNames", "")
	defer func() {
		if err != nil {
			tr.SetError(err)
		}

		fields := []otlog.Field{
			otlog.String("permission", string(p)),
			otlog.Int("repos.count", len(repos)),
			otlog.Int("authorized.count", len(accepted)),
		}

		if currentUser != nil {
			fields = append(fields, otlog.Object("user", currentUser))
		}

		tr.LogFields(fields...)

		tr.Finish()
	}()

	var accts []*extsvc.ExternalAccount
	authzAllowByDefault, authzProviders := authz.GetProviders()
	if len(authzProviders) > 0 && currentUser != nil {
		accts, err = ExternalAccounts.List(ctx, ExternalAccountsListOptions{UserID: currentUser.ID})
		if err != nil {
			return nil, err
		}
	}

	accepted = make(map[api.RepoName]struct{}) // repositories that have been claimed and have read permissions
	toverify := make(map[authz.Repo]struct{})  // repositories that have not been claimed by any authz provider
	for repo := range repos {
		// 🚨 SECURITY: Defensively bar access to repos with no external repo spec (we don't know
		// where they came from, so can't reliably enforce permissions). If external repo spec is
		// NOT set, then we exclude the repo (unless there are no authz providers and
		// `authzAllowByDefault` is true).
		if repo.ExternalRepoSpec.IsSet() {
			toverify[repo] = struct{}{}
		} else if authzAllowByDefault && len(authzProviders) == 0 {
			accepted[repo.RepoName] = struct{}{}
		}
	}

	// Walk through all authz providers, checking repo permissions against each. If any own a given
	// repo, we use its permissions for that repo.
	for _, authzProvider := range authzProviders {
		if len(toverify) == 0 {
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
		myToVerify, nextToVerify := authzProvider.Repos(ctx, toverify)

		// check the perms on those repos
		perms, err := authzProvider.RepoPerms(ctx, providerAcct, myToVerify)
		if err != nil {
			return nil, err
		}
		for repoToVerify := range myToVerify {
			if repoPerms, ok := perms[repoToVerify.RepoName]; ok && repoPerms[p] {
				accepted[repoToVerify.RepoName] = struct{}{}
			}
		}
		// continue checking repos that didn't belong to this authz provider
		toverify = nextToVerify
	}

	if authzAllowByDefault {
		for r := range toverify {
			accepted[r.RepoName] = struct{}{}
		}
	}

	return accepted, nil
}
