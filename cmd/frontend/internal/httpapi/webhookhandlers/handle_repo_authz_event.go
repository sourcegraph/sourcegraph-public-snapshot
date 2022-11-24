package webhookhandlers

import (
	"context"
	"fmt"

	gh "github.com/google/go-github/v43/github"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// handleGithubRepoAuthzEvent handles any github event containing a repository field, and enqueues the contained
// repo for permissions synchronisation.
func handleGitHubRepoAuthzEvent(opts authz.FetchPermsOptions) webhooks.Handler {
	return webhooks.Handler(func(ctx context.Context, db database.DB, urn extsvc.CodeHostBaseURL, payload any) error {
		if !conf.ExperimentalFeatures().EnablePermissionsWebhooks {
			return nil
		}
		if globals.PermissionsUserMapping().Enabled {
			return nil
		}

		log15.Debug("handleGitHubRepoAuthzEvent: Got github event", "type", fmt.Sprintf("%T", payload))

		e, ok := payload.(repoGetter)
		if !ok {
			return errors.Errorf("incorrect event type sent to github event handler: %T", payload)
		}
		return scheduleRepoUpdate(ctx, db, e.GetRepo(), opts)
	})
}

type repoGetter interface {
	GetRepo() *gh.Repository
}

// scheduleRepoUpdate finds an internal repo from a github repo, and posts it to repo-updater to
// schedule a permissions update
// 🚨 SECURITY: we want to be able to find any private repo here, so the DB call uses internal actor
func scheduleRepoUpdate(ctx context.Context, db database.DB, repo *gh.Repository, opts authz.FetchPermsOptions) error {
	if repo == nil {
		return nil
	}

	// 🚨 SECURITY: we want to be able to find any private repo here, so set internal actor
	ctx = actor.WithInternalActor(ctx)
	r, err := db.Repos().GetByName(ctx, api.RepoName("github.com/"+repo.GetFullName()))
	if err != nil {
		return err
	}

	log15.Debug("scheduleRepoUpdate: Dispatching permissions update", "repos", repo.GetFullName())

	logger := log.Scoped("TODO", "horsegraph")
	return database.PermissionSyncJobsWith(logger, db).CreateRepoSyncJob(ctx, int32(r.ID), true)
}
