package webhookhandlers

import (
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/authz"
)

func Init(w *webhooks.Router) {
	logger := log.Scoped("webhookhandlers")

	// Refer to https://docs.github.com/en/developers/webhooks-and-events/webhooks/webhook-events-and-payloads
	// for event types

	// Repository events
	w.Register(handleGitHubRepoAuthzEvent(logger, authz.FetchPermsOptions{}), "public")
	w.Register(handleGitHubRepoAuthzEvent(logger, authz.FetchPermsOptions{}), "repository")

	// Member refers to repository collaborators, and has both users and repos
	w.Register(handleGitHubRepoAuthzEvent(logger, authz.FetchPermsOptions{}), "member")
	w.Register(handleGitHubUserAuthzEvent(logger, authz.FetchPermsOptions{}), "member")

	// Events that touch cached permissions in authz/github.Provider implementation
	w.Register(handleGitHubRepoAuthzEvent(logger, authz.FetchPermsOptions{InvalidateCaches: true}), "team_add")
	w.Register(handleGitHubUserAuthzEvent(logger, authz.FetchPermsOptions{InvalidateCaches: true}), "organisation")
	w.Register(handleGitHubUserAuthzEvent(logger, authz.FetchPermsOptions{InvalidateCaches: true}), "membership")
}
