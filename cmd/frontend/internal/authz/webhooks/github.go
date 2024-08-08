package webhooks

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	gh "github.com/google/go-github/v55/github"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz/permssync"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var githubEvents = []string{
	"repository",
	"member",
	"organization",
	"membership",
	"team",
	"team_add",
}

type GitHubWebhook struct {
	logger log.Logger
}

func NewGitHubWebhook(logger log.Logger) *GitHubWebhook {
	return &GitHubWebhook{logger: logger}
}

func (h *GitHubWebhook) Register(router *webhooks.Router) {
	router.Register(
		h.handleGitHubWebhook,
		extsvc.KindGitHub,
		githubEvents...,
	)
}

// This should be set to zero for testing
var sleepTime = 10 * time.Second

func TestSetGitHubHandlerSleepTime(t *testing.T, val time.Duration) {
	old := sleepTime
	t.Cleanup(func() { sleepTime = old })
	sleepTime = val
}

func (h *GitHubWebhook) handleGitHubWebhook(ctx context.Context, db database.DB, codeHostURN extsvc.CodeHostBaseURL, payload any) error {
	// TODO: Should we return errors here? We could receive a webhook for a repo
	// that's not on Sourcegraph, so probably not?
	switch e := payload.(type) {
	case *gh.RepositoryEvent:
		_ = h.handleRepositoryEvent(ctx, db, e)
	case *gh.MemberEvent:
		_ = h.handleMemberEvent(ctx, db, e, codeHostURN)
	case *gh.OrganizationEvent:
		_ = h.handleOrganizationEvent(ctx, db, e, codeHostURN)
	case *gh.MembershipEvent:
		_ = h.handleMembershipEvent(ctx, db, e, codeHostURN)
	case *gh.TeamEvent:
		_ = h.handleTeamEvent(ctx, e, db)
	}
	return nil
}

func (h *GitHubWebhook) handleRepositoryEvent(ctx context.Context, db database.DB, e *gh.RepositoryEvent) error {
	// On repository events, we only care if a public repository is made private, in which case a permissions sync should happen
	if e.GetAction() != "privatized" {
		return nil
	}

	return h.getRepoAndSyncPerms(ctx, db, e, database.ReasonGitHubRepoMadePrivateEvent)
}

func (h *GitHubWebhook) handleMemberEvent(ctx context.Context, db database.DB, e *gh.MemberEvent, codeHostURN extsvc.CodeHostBaseURL) error {
	action := e.GetAction()
	var reason database.PermissionsSyncJobReason
	if action == "added" {
		reason = database.ReasonGitHubUserAddedEvent
	} else if action == "removed" {
		reason = database.ReasonGitHubUserRemovedEvent
	} else {
		// unknown event type
		return nil
	}
	user := e.GetMember()

	return h.getUserAndSyncPerms(ctx, db, user, codeHostURN, reason)
}

func (h *GitHubWebhook) handleOrganizationEvent(ctx context.Context, db database.DB, e *gh.OrganizationEvent, codeHostURN extsvc.CodeHostBaseURL) error {
	action := e.GetAction()
	var reason database.PermissionsSyncJobReason
	if action == "member_added" {
		reason = database.ReasonGitHubOrgMemberAddedEvent
	} else if action == "member_removed" {
		reason = database.ReasonGitHubOrgMemberRemovedEvent
	} else {
		return nil
	}

	user := e.GetMembership().GetUser()

	return h.getUserAndSyncPerms(ctx, db, user, codeHostURN, reason)
}

func (h *GitHubWebhook) handleMembershipEvent(ctx context.Context, db database.DB, e *gh.MembershipEvent, codeHostURN extsvc.CodeHostBaseURL) error {
	action := e.GetAction()
	var reason database.PermissionsSyncJobReason
	if action == "added" {
		reason = database.ReasonGitHubUserMembershipAddedEvent
	} else if action == "removed" {
		reason = database.ReasonGitHubUserMembershipRemovedEvent
	} else {
		return nil
	}
	user := e.GetMember()

	return h.getUserAndSyncPerms(ctx, db, user, codeHostURN, reason)
}

func (h *GitHubWebhook) handleTeamEvent(ctx context.Context, e *gh.TeamEvent, db database.DB) error {
	action := e.GetAction()
	var reason database.PermissionsSyncJobReason
	if action == "added_to_repository" {
		reason = database.ReasonGitHubTeamAddedToRepoEvent
	} else if action == "removed_from_repository" {
		reason = database.ReasonGitHubTeamRemovedFromRepoEvent
	} else {
		return nil
	}

	return h.getRepoAndSyncPerms(ctx, db, e, reason)
}

func (h *GitHubWebhook) getUserAndSyncPerms(ctx context.Context, db database.DB, user *gh.User, codeHostURN extsvc.CodeHostBaseURL, reason database.PermissionsSyncJobReason) error {
	externalAccounts, err := db.UserExternalAccounts().List(ctx, database.ExternalAccountsListOptions{
		ServiceID:      codeHostURN.String(),
		AccountID:      strconv.Itoa(int(user.GetID())),
		ExcludeExpired: true,
	})
	if err != nil {
		return err
	}

	if len(externalAccounts) == 0 {
		return errors.Newf("no github external accounts found with account id %d", user.GetID())
	}

	permssync.SchedulePermsSync(ctx, h.logger, db, permssync.ScheduleSyncOpts{
		UserIDs:      []int32{externalAccounts[0].UserID},
		Reason:       reason,
		ProcessAfter: time.Now().Add(sleepTime),
	})

	return err
}

func (h *GitHubWebhook) getRepoAndSyncPerms(ctx context.Context, db database.DB, e interface{ GetRepo() *gh.Repository }, reason database.PermissionsSyncJobReason) error {
	ghRepo := e.GetRepo()

	repo, err := db.Repos().GetFirstRepoByCloneURL(ctx, strings.TrimSuffix(ghRepo.GetCloneURL(), ".git"))
	if err != nil {
		return err
	}

	permssync.SchedulePermsSync(ctx, h.logger, db, permssync.ScheduleSyncOpts{
		RepoIDs:      []api.RepoID{repo.ID},
		Reason:       reason,
		ProcessAfter: time.Now().Add(sleepTime),
	})

	return nil
}
