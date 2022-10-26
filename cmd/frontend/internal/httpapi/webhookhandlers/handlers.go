package webhookhandlers

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func Init(db database.DB, w *webhooks.GitHubWebhook) {
	// Refer to https://docs.github.com/en/developers/webhooks-and-events/webhooks/webhook-events-and-payloads
	// for event types

	// Repository events
	w.Register(handleGitHubRepoAuthzEvent(db, authz.FetchPermsOptions{}), "public")
	w.Register(handleGitHubRepoAuthzEvent(db, authz.FetchPermsOptions{}), "repository")

	// Member refers to repository collaborators, and has both users and repos
	w.Register(handleGitHubRepoAuthzEvent(db, authz.FetchPermsOptions{}), "member")
	w.Register(handleGitHubUserAuthzEvent(db, authz.FetchPermsOptions{}), "member")

	// Events that touch cached permissions in authz/github.Provider implementation
	w.Register(handleGitHubRepoAuthzEvent(db, authz.FetchPermsOptions{InvalidateCaches: true}), "team_add")
	w.Register(handleGitHubUserAuthzEvent(db, authz.FetchPermsOptions{InvalidateCaches: true}), "organisation")
	w.Register(handleGitHubUserAuthzEvent(db, authz.FetchPermsOptions{InvalidateCaches: true}), "membership")
}
