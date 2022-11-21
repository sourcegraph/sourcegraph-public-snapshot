package webhooks

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

var (
    githubEvents = []string{
        "repository",
        "member",
        "organization",
        "membership",
        "team",
        "team_add",
    }
)

type GitHubWebhook struct {
}

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
    return nil
}
