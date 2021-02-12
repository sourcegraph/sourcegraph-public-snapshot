package webhookhandlers

import "github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"

func Init(w *webhooks.GitHubWebhook) {
	w.Register(handleGitHubRepoAuthzEvent, "public")
	w.Register(handleGitHubRepoAuthzEvent, "repository")
	w.Register(handleGitHubRepoAuthzEvent, "member") // member has both users and repos
	w.Register(handleGitHubRepoAuthzEvent, "team_add")

	w.Register(handleGitHubUserAuthzEvent, "organisation")
	w.Register(handleGitHubUserAuthzEvent, "member") // member has both users and repos
	w.Register(handleGitHubUserAuthzEvent, "membership")
}
