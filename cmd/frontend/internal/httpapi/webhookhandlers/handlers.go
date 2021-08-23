package webhookhandlers

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

func Init(db dbutil.DB, w *webhooks.GitHubWebhook) {
	// Refer to https://docs.github.com/en/developers/webhooks-and-events/webhooks/webhook-events-and-payloads
	// for event types

	w.Register(handleGitHubRepoAuthzEvent, "public")
	w.Register(handleGitHubRepoAuthzEvent, "repository")
	w.Register(handleGitHubRepoAuthzEvent, "member") // member has both users and repos
	w.Register(handleGitHubRepoAuthzEvent, "team_add")

	w.Register(handleGitHubUserAuthzEvent(db, authz.FetchPermsOptions{}), "member") // member has both users and repos

	// Events that touch cached permissions in authz/github.Provider implementation
	w.Register(handleGitHubUserAuthzEvent(db, authz.FetchPermsOptions{InvalidateCaches: true}), "organisation")
	w.Register(handleGitHubUserAuthzEvent(db, authz.FetchPermsOptions{InvalidateCaches: true}), "membership")
}
