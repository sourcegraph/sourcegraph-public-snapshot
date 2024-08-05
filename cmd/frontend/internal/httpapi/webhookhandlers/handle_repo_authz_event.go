package webhookhandlers

import (
	"context"
	"fmt"

	gh "github.com/google/go-github/v55/github"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/authz/permssync"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// handleGithubRepoAuthzEvent handles any github event containing a repository
// field, and enqueues the contained repo for permissions synchronisation.
func handleGitHubRepoAuthzEvent(logger log.Logger, opts authz.FetchPermsOptions) webhooks.Handler {
	return func(ctx context.Context, db database.DB, urn extsvc.CodeHostBaseURL, payload any) error {
		logger.Debug("handleGitHubRepoAuthzEvent: Got github event", log.String("type", fmt.Sprintf("%T", payload)))

		e, ok := payload.(repoGetter)
		if !ok {
			return errors.Errorf("incorrect event type sent to github event handler: %T", payload)
		}
		return scheduleRepoUpdate(ctx, logger, db, e.GetRepo(), opts)
	}
}

type repoGetter interface {
	GetRepo() *gh.Repository
}

// scheduleRepoUpdate finds an internal repo from a github repo, and posts it to
// repo-updater to schedule a permissions update
//
// 🚨 SECURITY: we want to be able to find any private repo here, so the DB call
// uses internal actor
func scheduleRepoUpdate(ctx context.Context, logger log.Logger, db database.DB, repo *gh.Repository, opts authz.FetchPermsOptions) error {
	if repo == nil {
		return nil
	}

	// 🚨 SECURITY: we want to be able to find any private repo here, so set internal actor
	ctx = actor.WithInternalActor(ctx)
	r, err := db.Repos().GetByName(ctx, api.RepoName("github.com/"+repo.GetFullName()))
	if err != nil {
		return err
	}

	logger.Debug("scheduleRepoUpdate: Dispatching permissions update", log.String("repo", repo.GetFullName()))

	permssync.SchedulePermsSync(ctx, logger, db, permssync.ScheduleSyncOpts{
		RepoIDs: []api.RepoID{r.ID},
		Options: opts,
		Reason:  database.ReasonGitHubRepoEvent,
	})

	return nil
}
