package webhookhandlers

import "github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"

func Init(w *webhooks.GithubWebhook) {
	w.Register(handleGithubRepoAuthzEvent, "public")
	w.Register(handleGithubRepoAuthzEvent, "repository")
	w.Register(handleGithubRepoAuthzEvent, "member") // member has both users and repos

	w.Register(handleGithubUserAuthzEvent, "team_add")
	w.Register(handleGithubUserAuthzEvent, "organisation")
	w.Register(handleGithubUserAuthzEvent, "member") // member has both users and repos
	w.Register(handleGithubUserAuthzEvent, "membership")
}
