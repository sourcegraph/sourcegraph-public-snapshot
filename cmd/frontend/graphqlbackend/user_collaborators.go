package graphqlbackend

import (
	"context"

	"github.com/inconshreveable/log15"

	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (r *UserResolver) InvitableCollaborators(ctx context.Context) ([]*invitableCollaboratorResolver, error) {
	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return nil, errors.New("no current user")
	}

	// We'll search for collaborators in 25 of the user's most-starred repositories.
	const maxReposToScan = 25
	pickedRepos, err := backend.NewRepos(r.db.Repos()).List(ctx, database.ReposListOptions{
		// SECURITY: This must be the authenticated user's ID.
		UserID:                 a.UID,
		IncludeUserPublicRepos: true,
		OrderBy: database.RepoListOrderBy{{
			Field:      "stars",
			Descending: true,
		}},
		LimitOffset: &database.LimitOffset{Limit: maxReposToScan},
	})
	if err != nil {
		return nil, errors.Wrap(err, "Repos.List")
	}

	// In parallel collect all recent committers info for the few repos we're going to scan.
	recentCommitters, err := gitserverParallelRecentCommitters(ctx, pickedRepos, git.Commits)
	if err != nil {
		return nil, err
	}

	authUserEmails, err := database.UserEmails(r.db).ListByUser(ctx, database.UserEmailsListOptions{
		UserID: a.UID,
	})
	if err != nil {
		return nil, err
	}

	userExistsByUsername := func(username string) bool {
		_, err := r.db.Users().GetByUsername(ctx, username)
		return err == nil
	}
	userExistsByEmail := func(email string) bool {
		_, err := r.db.Users().GetByVerifiedEmail(ctx, email)
		return err == nil
	}
	return filterInvitableCollaborators(recentCommitters, authUserEmails, userExistsByUsername, userExistsByEmail), nil
}

type GitCommitsFunc func(context.Context, api.RepoName, git.CommitsOptions, authz.SubRepoPermissionChecker) ([]*gitdomain.Commit, error)

func gitserverParallelRecentCommitters(ctx context.Context, repos []*types.Repo, gitCommits GitCommitsFunc) (allRecentCommitters []*invitableCollaboratorResolver, err error) {
	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)
	for _, repo := range repos {
		repo := repo
		wg.Add(1)
		goroutine.Go(func() {
			defer wg.Done()

			recentCommits, err := gitCommits(ctx, repo.Name, git.CommitsOptions{
				N:                200,
				NoEnsureRevision: true, // Don't try to fetch missing commits.
				NameOnly:         true, // Don't fetch detailed info like commit diffs.
			}, authz.DefaultSubRepoPermsChecker)
			if err != nil {
				log15.Error("InvitableCollaborators: failed to get recent committers", "err", err)
				return
			}
			mu.Lock()
			defer mu.Unlock()

			for _, commit := range recentCommits {
				for _, collaborator := range []*gitdomain.Signature{&commit.Author, commit.Committer} {
					if collaborator == nil {
						continue
					}
					allRecentCommitters = append(allRecentCommitters, &invitableCollaboratorResolver{
						likelySourcegraphUsername: "",
						email:                     collaborator.Email,
						name:                      collaborator.Name,
						avatarURL:                 "",
						date:                      commit.Author.Date,
					})
				}
			}
		})
	}
	wg.Wait()
	return
}
