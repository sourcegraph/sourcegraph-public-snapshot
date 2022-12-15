package webhookhandlers

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/authz"
)

func Init(w *webhooks.WebhookRouter) {
	// Refer to https://docs.github.com/en/developers/webhooks-and-events/webhooks/webhook-events-and-payloads
	// for event types

	// Repository events
	w.Register(handleGitHubRepoAuthzEvent(authz.FetchPermsOptions{}), "public")
	w.Register(handleGitHubRepoAuthzEvent(authz.FetchPermsOptions{}), "repository")

	// Member refers to repository collaborators, and has both users and repos
	w.Register(handleGitHubRepoAuthzEvent(authz.FetchPermsOptions{}), "member")
	w.Register(handleGitHubUserAuthzEvent(authz.FetchPermsOptions{}), "member")

	// Events that touch cached permissions in authz/github.Provider implementation
	w.Register(handleGitHubRepoAuthzEvent(authz.FetchPermsOptions{InvalidateCaches: true}), "team_add")
	w.Register(handleGitHubUserAuthzEvent(authz.FetchPermsOptions{InvalidateCaches: true}), "organisation")
	w.Register(handleGitHubUserAuthzEvent(authz.FetchPermsOptions{InvalidateCaches: true}), "membership")
}
