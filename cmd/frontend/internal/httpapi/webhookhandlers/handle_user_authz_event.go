package webhookhandlers

import (
	"context"
	"fmt"
	"strconv"

	gh "github.com/google/go-github/v43/github"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// handleGitHubUserAuthzEvent handles a github webhook for the events described in webhookhandlers/handlers.go
// extracting a user from the github event and scheduling it for a perms update in repo-updater
func handleGitHubUserAuthzEvent(opts authz.FetchPermsOptions) webhooks.Handler {
	return webhooks.Handler(func(ctx context.Context, db database.DB, _ extsvc.CodeHostBaseURL, payload any) error {
		if !conf.ExperimentalFeatures().EnablePermissionsWebhooks {
			return nil
		}
		if globals.PermissionsUserMapping().Enabled {
			return nil
		}

		log15.Debug("handleGitHubUserAuthzEvent: Got github event", "type", fmt.Sprintf("%T", payload))

		var user *gh.User

		// github events contain a user object at a few different levels, so try and find the first that matches
		// and extract the user
		switch e := payload.(type) {
		case memberGetter:
			user = e.GetMember()
		case membershipGetter:
			user = e.GetMembership().GetUser()
		}
		if user == nil {
			return errors.Errorf("could not extract GitHub user from %T GitHub event", payload)
		}

		return scheduleUserUpdate(ctx, db, user, opts)
	})
}

type memberGetter interface {
	GetMember() *gh.User
}

type membershipGetter interface {
	GetMembership() *gh.Membership
}

func scheduleUserUpdate(ctx context.Context, db database.DB, githubUser *gh.User, opts authz.FetchPermsOptions) error {
	if githubUser == nil {
		return nil
	}
	accs, err := db.UserExternalAccounts().List(ctx, database.ExternalAccountsListOptions{
		ServiceType: "github",
		AccountID:   strconv.FormatInt(githubUser.GetID(), 10),
	})
	if err != nil {
		return err
	}
	if len(accs) == 0 {
		// this user is not a sourcegraph user (yet...)
		return nil
	}

	store := db.PermissionSyncJobs()
	jobOpts := database.PermissionSyncJobOpts{
		HighPriority:     true,
		InvalidateCaches: opts.InvalidateCaches,
	}
	for _, acc := range accs {
		if err := store.CreateUserSyncJob(ctx, acc.UserID, jobOpts); err != nil {
			return err
		}
	}

	return nil
}
