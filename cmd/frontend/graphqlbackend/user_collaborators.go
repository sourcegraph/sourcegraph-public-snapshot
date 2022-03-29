package graphqlbackend

import (
	"context"
	"crypto/md5"
	"fmt"
	"sync"

	"github.com/inconshreveable/log15"

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
	db := r.db
	pickedRepos, err := backend.NewRepos(db).List(ctx, database.ReposListOptions{
		// SECURITY: This must be the authenticated user's ID.
		UserID:                 a.UID,
		IncludeUserPublicRepos: false,
		NoForks:                true,
		NoArchived:             true,
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
	recentCommitters := gitserverParallelRecentCommitters(ctx, db, pickedRepos, git.Commits)

	authUserEmails, err := database.UserEmails(db).ListByUser(ctx, database.UserEmailsListOptions{
		UserID: a.UID,
	})
	if err != nil {
		return nil, err
	}

	userExistsByUsername := func(username string) bool {
		// We do not actually have usernames, gitserverParallelRecentCommitters does not produce
		// them and so we always have an empty string here. However, we leave this function
		// implemented for the future where we may.
		if username == "" {
			return false
		}
		_, err := db.Users().GetByUsername(ctx, username)
		return err == nil
	}
	userExistsByEmail := func(email string) bool {
		_, err := db.Users().GetByVerifiedEmail(ctx, email)
		return err == nil
	}
	return filterInvitableCollaborators(recentCommitters, authUserEmails, userExistsByUsername, userExistsByEmail), nil
}

type GitCommitsFunc func(context.Context, database.DB, api.RepoName, git.CommitsOptions, authz.SubRepoPermissionChecker) ([]*gitdomain.Commit, error)

func gitserverParallelRecentCommitters(ctx context.Context, db database.DB, repos []*types.Repo, gitCommits GitCommitsFunc) (allRecentCommitters []*invitableCollaboratorResolver) {
	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)
	for _, repo := range repos {
		repo := repo
		wg.Add(1)
		goroutine.Go(func() {
			defer wg.Done()

			recentCommits, err := gitCommits(ctx, db, repo.Name, git.CommitsOptions{
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

					// We cannot do anything better than a Gravatar profile picture for the
					// collaborator. GitHub does not provide an API that allows us to lookup a user
					// by email effectively: only their older search API can do so, and it is rate
					// limited *heavily* to just 30 req/min per API token. For an enterprise instance
					// that token is shared between all Sourcegraph users, and so is a non-viable
					// approach.
					gravatarURL := fmt.Sprintf("https://www.gravatar.com/avatar/%x?d=mp", md5.Sum([]byte(collaborator.Email)))

					allRecentCommitters = append(allRecentCommitters, &invitableCollaboratorResolver{
						likelySourcegraphUsername: "",
						email:                     collaborator.Email,
						name:                      collaborator.Name,
						avatarURL:                 gravatarURL,
						date:                      commit.Author.Date,
					})
				}
			}
		})
	}
	wg.Wait()
	return
}
