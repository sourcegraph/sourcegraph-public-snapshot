package webhooks

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	gh "github.com/google/go-github/v43/github"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
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
	// TODO: This MUST be removed once permissions syncing jobs are database backed!
	// If we react too quickly to a webhook, the changes may not yet have properly
	// propagated on GitHub's system, and we'll get old results, making the
	// webhook useless.
	// We have to wait some amount of time to process the webhook to ensure
	// that we are getting fresh results.
	go func() {
		time.Sleep(sleepTime)
		eventContext, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()

		switch e := payload.(type) {
		case *gh.RepositoryEvent:
			h.handleRepositoryEvent(eventContext, db, e)
		case *gh.MemberEvent:
			h.handleMemberEvent(eventContext, db, e, codeHostURN)
		case *gh.OrganizationEvent:
			h.handleOrganizationEvent(eventContext, db, e, codeHostURN)
		case *gh.MembershipEvent:
			h.handleMembershipEvent(eventContext, db, e, codeHostURN)
		case *gh.TeamEvent:
			h.handleTeamEvent(eventContext, e, db)
		}
	}()
	return nil
}

func (h *GitHubWebhook) handleRepositoryEvent(ctx context.Context, db database.DB, e *gh.RepositoryEvent) error {
	// On repository events, we only care if a public repository is made private, in which case a permissions sync should happen
	if e.GetAction() != "privatized" {
		return nil
	}

	return h.getRepoAndSyncPerms(ctx, db, e)
}

func (h *GitHubWebhook) handleMemberEvent(ctx context.Context, db database.DB, e *gh.MemberEvent, codeHostURN extsvc.CodeHostBaseURL) error {
	if e.GetAction() != "added" && e.GetAction() != "removed" {
		return nil
	}
	user := e.GetMember()

	return h.getUserAndSyncPerms(ctx, db, user, codeHostURN)
}

func (h *GitHubWebhook) handleOrganizationEvent(ctx context.Context, db database.DB, e *gh.OrganizationEvent, codeHostURN extsvc.CodeHostBaseURL) error {
	if e.GetAction() != "member_added" && e.GetAction() != "member_removed" {
		return nil
	}

	user := e.GetMembership().GetUser()

	return h.getUserAndSyncPerms(ctx, db, user, codeHostURN)
}

func (h *GitHubWebhook) handleMembershipEvent(ctx context.Context, db database.DB, e *gh.MembershipEvent, codeHostURN extsvc.CodeHostBaseURL) error {
	if e.GetAction() != "added" && e.GetAction() != "removed" {
		return nil
	}
	user := e.GetMember()

	return h.getUserAndSyncPerms(ctx, db, user, codeHostURN)
}

func (h *GitHubWebhook) handleTeamEvent(ctx context.Context, e *gh.TeamEvent, db database.DB) error {
	if e.GetAction() != "added_to_repository" && e.GetAction() != "removed_from_repository" {
		return nil
	}

	return h.getRepoAndSyncPerms(ctx, db, e)
}

func (h *GitHubWebhook) getUserAndSyncPerms(ctx context.Context, db database.DB, user *gh.User, codeHostURN extsvc.CodeHostBaseURL) error {
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

	permsJobsStore := database.PermissionSyncJobsWith(h.logger.Scoped("PermissionSyncJobsStore", ""), db)
	err = permsJobsStore.CreateUserSyncJob(ctx, externalAccounts[0].UserID, database.PermissionSyncJobOpts{
		HighPriority: true,
	})
	if err != nil {
		h.logger.Error("could not schedule permissions sync for user", log.Error(err), log.Int("user ID", int(externalAccounts[0].UserID)))
	}

	return err
}

func (h *GitHubWebhook) getRepoAndSyncPerms(ctx context.Context, db database.DB, e interface{ GetRepo() *gh.Repository }) error {
	ghRepo := e.GetRepo()

	repo, err := db.Repos().GetFirstRepoByCloneURL(ctx, strings.TrimSuffix(ghRepo.GetCloneURL(), ".git"))
	if err != nil {
		return err
	}

	permsJobsStore := database.PermissionSyncJobsWith(h.logger.Scoped("PermissionSyncJobsStore", ""), db)
	err = permsJobsStore.CreateRepoSyncJob(ctx, int32(repo.ID), database.PermissionSyncJobOpts{
		HighPriority: true,
	})
	if err != nil {
		h.logger.Error("could not schedule permissions sync for repo", log.Error(err), log.Int("repo ID", int(repo.ID)))
	}

	return err
}
