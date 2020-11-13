package webhookhandlers

import (
	"context"
	"fmt"
	gh "github.com/google/go-github/v28/github"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
)

func handleGitHubUserAuthzEvent(ctx context.Context, extSvc *repos.ExternalService, payload interface{}) error {
	if !conf.Get().ExperimentalFeatures.EnablePermissionsWebhooks {
		return nil
	}

	log15.Debug(fmt.Sprintf("handleGitHubUserAuthzEvent: Got github event %T", payload))

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
		return fmt.Errorf("could not extract GitHub user from %T GitHub event", payload)
	}

	return scheduleUserUpdate(ctx, user)
}

type memberGetter interface {
	GetMember() *gh.User
}

type membershipGetter interface {
	GetMembership() *gh.Membership
}

func scheduleUserUpdate(ctx context.Context, githubUser *gh.User) error {
	if githubUser == nil {
		return nil
	}
	accs, err := db.ExternalAccounts.List(ctx, db.ExternalAccountsListOptions{
		AccountID: githubUser.GetID(),
	})
	if err != nil {
		return err
	}
	if len(accs) == 0 {
		// this user is not a sourcegraph user (yet...)
		return nil
	}
	if len(accs) > 1 {
		return fmt.Errorf("could not map github user to external account, %d external accounts returned, 1 expected", len(accs))
	}

	log15.Debug(fmt.Sprintf("scheduleUserUpdate: Dispatching permissions update for user %d", accs[0].UserID))

	c := repoupdater.DefaultClient
	return c.SchedulePermsSync(ctx, protocol.PermsSyncRequest{
		UserIDs: []int32{accs[0].UserID},
	})
}
