package webhooks

import (
	"context"
	"strconv"
	"strings"

	gh "github.com/google/go-github/v43/github"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
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
	return &GitHubWebhook{logger}
}

func (h *GitHubWebhook) Register(router *webhooks.WebhookRouter) {
	router.Register(
		h.handleGitHubWebhook,
		extsvc.KindGitHub,
		githubEvents...,
	)
}

func (h *GitHubWebhook) handleGitHubWebhook(ctx context.Context, db database.DB, codeHostURN extsvc.CodeHostBaseURL, payload any) error {
	switch e := payload.(type) {
	case *gh.RepositoryEvent:
		return h.handleRepositoryEvent(ctx, db, e)
	case *gh.MemberEvent:
		return h.handleMemberEvent(ctx, db, e, codeHostURN)
	case *gh.OrganizationEvent:
		return h.handleOrganizationEvent(ctx, db, e, codeHostURN)
	case *gh.MembershipEvent:
		return h.handleMembershipEvent(ctx, db, e, codeHostURN)
	case *gh.TeamEvent:
		return h.handleTeamEvent(ctx, e, db)
	}
	return nil
}

func (h *GitHubWebhook) handleRepositoryEvent(ctx context.Context, db database.DB, e *gh.RepositoryEvent) error {
	// On repository events, we only care if a public repository is made private, in which case a permissions sync should happen
	if e.GetAction() != "privatized" {
		return nil
	}

	ghRepo := e.GetRepo()
	if ghRepo == nil {
		return errors.New("no repo found in webhook event")
	}

	repoName, err := db.Repos().GetFirstRepoNameByCloneURL(ctx, strings.TrimSuffix(ghRepo.GetCloneURL(), ".git"))
	if err != nil {
		return err
	}

	repo, err := db.Repos().GetByName(ctx, repoName)
	if err != nil {
		return err
	}

	err = repoupdater.DefaultClient.SchedulePermsSync(ctx, protocol.PermsSyncRequest{
		RepoIDs: []api.RepoID{repo.ID},
	})
	if err != nil {
		h.logger.Error("could not schedule permissions sync for repo", log.Error(err), log.Int("repo ID", int(repo.ID)))
	}

	return err
}

func (h *GitHubWebhook) handleMemberEvent(ctx context.Context, db database.DB, e *gh.MemberEvent, codeHostURN extsvc.CodeHostBaseURL) error {
	if e.GetAction() != "added" && e.GetAction() != "removed" {
		return nil
	}

	user := e.GetMember()
	if user == nil {
		return errors.New("no user found in webhook event")
	}

	externalAccounts, err := db.UserExternalAccounts().List(ctx, database.ExternalAccountsListOptions{
		ServiceID:      codeHostURN.String(),
		AccountID:      strconv.Itoa(int(user.GetID())),
		ExcludeExpired: true,
	})
	if err != nil {
		return err
	}

	if len(externalAccounts) == 0 {
		return errors.Newf("no external accounts found for user with external account id %d", user.GetID())
	}

	err = repoupdater.DefaultClient.SchedulePermsSync(ctx, protocol.PermsSyncRequest{
		UserIDs: []int32{externalAccounts[0].UserID},
	})
	if err != nil {
		h.logger.Error("could not schedule permissions sync for user", log.Error(err), log.Int("user ID", int(externalAccounts[0].UserID)))
	}

	return err
}

func (h *GitHubWebhook) handleOrganizationEvent(ctx context.Context, db database.DB, e *gh.OrganizationEvent, codeHostURN extsvc.CodeHostBaseURL) error {
	if e.GetAction() != "member_added" && e.GetAction() != "member_removed" {
		return nil
	}

	user := e.GetMembership().GetUser()
	if user == nil {
		return errors.New("could not get user from webhook event")
	}

	externalAccounts, err := db.UserExternalAccounts().List(ctx, database.ExternalAccountsListOptions{
		ServiceID:      codeHostURN.String(),
		AccountID:      strconv.Itoa(int(user.GetID())),
		ExcludeExpired: true,
	})
	if err != nil {
		return err
	}

	if len(externalAccounts) == 0 {
		return errors.Newf("no external accounts found for user with external account id %d", user.GetID())
	}

	err = repoupdater.DefaultClient.SchedulePermsSync(ctx, protocol.PermsSyncRequest{
		UserIDs: []int32{externalAccounts[0].UserID},
	})
	if err != nil {
		h.logger.Error("could not schedule permissions sync for user", log.Error(err), log.Int("user ID", int(externalAccounts[0].UserID)))
	}

	return err
}

func (h *GitHubWebhook) handleMembershipEvent(ctx context.Context, db database.DB, e *gh.MembershipEvent, codeHostURN extsvc.CodeHostBaseURL) error {
	if e.GetAction() != "added" && e.GetAction() != "removed" {
		return nil
	}

	user := e.GetMember()
	if user == nil {
		return errors.New("no user found in webhook event")
	}

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

	err = repoupdater.DefaultClient.SchedulePermsSync(ctx, protocol.PermsSyncRequest{
		UserIDs: []int32{externalAccounts[0].UserID},
	})
	if err != nil {
		h.logger.Error("could not schedule permissions sync for user", log.Error(err), log.Int("user ID", int(externalAccounts[0].UserID)))
	}

	return err
}

func (h *GitHubWebhook) handleTeamEvent(ctx context.Context, e *gh.TeamEvent, db database.DB) error {
	if e.GetAction() != "added_to_repository" && e.GetAction() != "removed_from_repository" {
		return nil
	}

	ghRepo := e.GetRepo()
	if ghRepo == nil {
		return errors.New("no repo found in webhook event")
	}

	repoName, err := db.Repos().GetFirstRepoNameByCloneURL(ctx, strings.TrimSuffix(ghRepo.GetCloneURL(), ".git"))
	if err != nil {
		return err
	}

	repo, err := db.Repos().GetByName(ctx, repoName)
	if err != nil {
		return err
	}

	err = repoupdater.DefaultClient.SchedulePermsSync(ctx, protocol.PermsSyncRequest{
		RepoIDs: []api.RepoID{repo.ID},
	})
	if err != nil {
		h.logger.Error("could not schedule permissions sync for repo", log.Error(err), log.Int("repo ID", int(repo.ID)))
	}

	return err
}
