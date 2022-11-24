package webhooks

import (
	"context"
	"strconv"
	"strings"

	gh "github.com/google/go-github/v43/github"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
)

var githubEvents = []string{
	"repository",
	"member",
	"organization",
	"membership",
	"team",
	"team_add",
}

type GitHubWebhook struct{}

func NewGitHubWebhook() *GitHubWebhook {
	return &GitHubWebhook{}
}

func (h *GitHubWebhook) Register(router *webhooks.WebhookRouter) {
	router.Register(
		h.handleGitHubWebhook,
		extsvc.KindGitHub,
		githubEvents...,
	)
}

func (h *GitHubWebhook) handleGitHubWebhook(ctx context.Context, db database.DB, codeHostURN extsvc.CodeHostBaseURL, payload any) error {
	h.convertEvent(ctx, db, codeHostURN, payload)
	return nil
}

func (h *GitHubWebhook) convertEvent(ctx context.Context, db database.DB, codeHostURN extsvc.CodeHostBaseURL, theirs any) {
	repoupdaterClient := repoupdater.DefaultClient
	switch e := theirs.(type) {
	case *gh.RepositoryEvent:
		repo := e.GetRepo()
		if repo == nil {
			return
		}

		if e.Action == nil {
			return
		}

		switch e.GetAction() {
		case "privatized":
			// Get a list of collaborators for the repository
			// For each collaborator:
			//    Use the collaborator's access token to confirm that they can access the repository
			repoName, err := db.Repos().GetFirstRepoNameByCloneURL(ctx, strings.TrimSuffix(repo.GetCloneURL(), ".git"))
			if err != nil {
				return
			}

			repo, err := db.Repos().GetByName(ctx, repoName)
			if err != nil {
				return
			}

			repoupdaterClient.SchedulePermsSync(ctx, protocol.PermsSyncRequest{
				RepoIDs: []api.RepoID{repo.ID},
			})
		}

	case *gh.MemberEvent:
		user := e.GetMember()
		if user == nil {
			return
		}

		externalAccounts, err := db.UserExternalAccounts().ListBySQL(ctx, sqlf.Sprintf("WHERE service_type=%s AND account_id=%s", extsvc.TypeGitHub, strconv.FormatInt(user.GetID(), 10)))
		if err != nil {
			return
		}

		if len(externalAccounts) <= 0 {
			return
		}

		switch e.GetAction() {
		case "added":
			repoupdaterClient.SchedulePermsSync(ctx, protocol.PermsSyncRequest{
				UserIDs: []int32{externalAccounts[0].UserID},
			})
		case "removed":
			// User's token is used to confirm access to the repository.
			// User loses access if not confirmed.
			repoupdaterClient.SchedulePermsSync(ctx, protocol.PermsSyncRequest{
				UserIDs: []int32{externalAccounts[0].UserID},
			})
		}

	case *gh.OrganizationEvent:
		user := e.GetMembership().GetUser()
		if user == nil {
			return
		}

		if e.Action == nil {
			return
		}

		externalAccounts, err := db.UserExternalAccounts().ListBySQL(ctx, sqlf.Sprintf("WHERE service_type=%s AND account_id=%s", extsvc.TypeGitHub, strconv.FormatInt(user.GetID(), 10)))
		if err != nil {
			return
		}

		if len(externalAccounts) <= 0 {
			return
		}

		switch e.GetAction() {
		case "member_added":
			// Use the user's token to fetch a list of repositories that belong to the organization, grant access to each
			err := repoupdaterClient.SchedulePermsSync(ctx, protocol.PermsSyncRequest{
				UserIDs: []int32{externalAccounts[0].UserID},
			})
			if err != nil {
				return
			}
		case "member_removed":
			// Use the user's token to fetch the list of repositories that belong to the organization.
			// Access to organization repositories not on the list is removed
			err := repoupdaterClient.SchedulePermsSync(ctx, protocol.PermsSyncRequest{
				UserIDs: []int32{externalAccounts[0].UserID},
			})
			if err != nil {
				return
			}
		}

	case *gh.MembershipEvent:
		user := e.GetMember()
		if user == nil {
			return
		}

		if e.Action == nil {
			return
		}
		externalAccounts, err := db.UserExternalAccounts().ListBySQL(ctx, sqlf.Sprintf("WHERE service_type=%s AND account_id=%s", extsvc.TypeGitHub, strconv.FormatInt(user.GetID(), 10)))
		if err != nil {
			return
		}

		if len(externalAccounts) <= 0 {
			return
		}

		switch e.GetAction() {
		case "added":
			// Fetch team repositories using user token
			repoupdaterClient.SchedulePermsSync(ctx, protocol.PermsSyncRequest{
				UserIDs: []int32{externalAccounts[0].UserID},
			})
		case "removed":
			// Use code host connection token to list the team's repositories
			// Use user's access token to check access to each of the repositories
			// User may still have access to the repos by other means (explicit collaborator)
			repoupdaterClient.SchedulePermsSync(ctx, protocol.PermsSyncRequest{
				UserIDs: []int32{externalAccounts[0].UserID},
			})
		}

	case *gh.TeamEvent:
		repo := e.GetRepo()
		if repo == nil {
			return
		}

		if e.Action == nil {
			return
		}
		repoName, err := db.Repos().GetFirstRepoNameByCloneURL(ctx, strings.TrimSuffix(repo.GetCloneURL(), ".git"))
		if err != nil {
			return
		}

		ghRepo, err := db.Repos().GetByName(ctx, repoName)
		if err != nil {
			return
		}

		switch e.GetAction() {
		case "added_to_repository":
			// Fetch list of members on the team
			// Use each member's user token to determine access to the repository
			repoupdaterClient.SchedulePermsSync(ctx, protocol.PermsSyncRequest{
				RepoIDs: []api.RepoID{ghRepo.ID},
			})
		case "removed_from_repository":
			// Fetch list of members on the team
			// Use each member's user token to determine access to the repository
			// Team members can still have access to a repo by other means
			repoupdaterClient.SchedulePermsSync(ctx, protocol.PermsSyncRequest{
				RepoIDs: []api.RepoID{ghRepo.ID},
			})
		}
	}
}
