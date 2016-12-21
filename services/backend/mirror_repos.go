package backend

import (
	"strings"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github"
)

// emptyGitCommitID is used in githttp.Event objects in the Last (or
// Commit) field to signify that a branch was created (or deleted).
const emptyGitCommitID = "0000000000000000000000000000000000000000"

var MirrorRepos = &mirrorRepos{}

type mirrorRepos struct{}

func (s *mirrorRepos) RefreshVCS(ctx context.Context, op *sourcegraph.MirrorReposRefreshVCSOp) (err error) {
	if Mocks.MirrorRepos.RefreshVCS != nil {
		return Mocks.MirrorRepos.RefreshVCS(ctx, op)
	}

	ctx, done := trace(ctx, "MirrorRepos", "RefreshVCS", op, &err)
	defer done()

	ctx = context.WithValue(ctx, github.GitHubTrackingContextKey, "RefreshVCS")
	actor := authpkg.ActorFromContext(ctx)
	asUserUID := actor.UID

	// Only admin users and the repo updater are allowed to perform this operation
	// as a different user.
	canImpersonateUser := actor.HasScope("internal:repoupdater")
	if op.AsUser != nil && canImpersonateUser {
		asUserUID = op.AsUser.UID
	}

	repo, err := localstore.Repos.Get(ctx, op.Repo)
	if err != nil {
		log15.Error("RefreshVCS: failed to get repo", "error", err, "repo", op.Repo)
		return err
	}

	// Use the auth token for asUserUID if it can be successfully looked up (it may fail if that user doesn't have one),
	// otherwise proceed without their credentials. It will work for public repos.
	remoteOpts := vcs.RemoteOpts{}
	if asUserUID != "" {
		switch {
		case strings.HasPrefix(strings.ToLower(repo.URI), "github.com/"):
			extToken, err := authpkg.FetchGitHubToken(ctx, asUserUID)
			if err != nil {
				break
			}

			// Set the auth token to be used in repo VCS operations.
			remoteOpts.HTTPS = &vcs.HTTPSConfig{
				User: "x-oauth-token", // User is unused by GitHub, but provide a non-empty value to satisfy git.
				Pass: extToken.Token,
			}

			// Set a GitHub client authed as the user in the context, to be used to make GitHub API calls.
			ctx = github.NewContextWithAuthedClient(authpkg.WithActor(ctx, &authpkg.Actor{UID: asUserUID, GitHubToken: extToken.Token}))

		case strings.HasPrefix(strings.ToLower(repo.URI), "source.developers.google.com/p/"):
			extToken, err := authpkg.FetchGoogleToken(ctx, asUserUID)
			if err != nil {
				log15.Warn("refreshing vcs with user, but problem fetching google token", "error", err)
				break
			}
			username, err := authpkg.FetchGoogleUsername(ctx, asUserUID)
			if err != nil {
				log15.Warn("refreshing vcs with user, but problem fetching google username", "error", err)
				break
			}

			// Set the auth token to be used in repo VCS operations.
			remoteOpts.HTTPS = &vcs.HTTPSConfig{
				User: username,
				Pass: extToken.Token,
			}
		}
	}

	vcsRepo, err := localstore.RepoVCS.Open(ctx, repo.ID)
	if err != nil {
		log15.Error("RefreshVCS: failed to open VCS", "error", err, "URI", repo.URI)
		return err
	}
	if err := vcsRepo.UpdateEverything(ctx, remoteOpts); err != nil {
		if !vcs.IsRepoNotExist(err) {
			log15.Error("RefreshVCS: update repo failed unexpectedly", "error", err, "repo", repo.URI)
			return err
		}
		if err.(vcs.RepoNotExistError).CloneInProgress {
			log15.Info("RefreshVCS: clone in progress, not updating", "repo", repo.URI)
			return nil
		}
		if err := s.cloneRepo(ctx, repo, remoteOpts); err != nil {
			log15.Info("RefreshVCS: cloneRepo failed", "error", err, "repo", repo.URI)
			return err
		}
	}

	{
		now := time.Now()
		ctx2 := authpkg.WithActor(ctx, &authpkg.Actor{Scope: map[string]bool{"internal:repo-internal-update": true}})
		if err := localstore.Repos.InternalUpdate(ctx2, repo.ID, localstore.InternalRepoUpdate{VCSSyncedAt: &now}); err != nil {
			log15.Info("RefreshVCS: updating repo internal VCSSyncedAt failed", "err", err, "repo", repo.URI)
			return err
		}
	}

	return nil
}

var skipCloneRepoAsyncSteps = false

func (s *mirrorRepos) cloneRepo(ctx context.Context, repo *sourcegraph.Repo, remoteOpts vcs.RemoteOpts) error {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "MirrorRepos.cloneRepo", repo.URI); err != nil {
		return err
	}

	err := localstore.RepoVCS.Clone(elevatedActor(ctx), repo.ID, &localstore.CloneInfo{
		CloneURL:   repo.HTTPCloneURL,
		RemoteOpts: remoteOpts,
	})
	if err != nil && err != vcs.ErrRepoExist {
		return err
	}

	// We've just cloned the repository, do a sanity check to ensure we
	// can resolve the DefaultBranch.
	_, err = Repos.ResolveRev(elevatedActor(ctx), &sourcegraph.ReposResolveRevOp{
		Repo: repo.ID,
		Rev:  repo.DefaultBranch,
	})
	if err != nil {
		return err
	}

	return nil
}

type MockMirrorRepos struct {
	RefreshVCS func(v0 context.Context, v1 *sourcegraph.MirrorReposRefreshVCSOp) error
}
