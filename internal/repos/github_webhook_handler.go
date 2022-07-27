package repos

import (
	"context"

	gh "github.com/google/go-github/v43/github"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitHubWebhookHandler struct {
	client *github.V3Client
}

func newGitHubWebhookHandler(client *github.V3Client) *GitHubWebhookHandler {
	return &GitHubWebhookHandler{client: client}
}

func (g *GitHubWebhookHandler) handleGitHubWebhook(ctx context.Context, extSvc *types.ExternalService, payload any) error {
	event, ok := payload.(*gh.PushEvent)
	if !ok {
		return errors.Newf("expected GitHub.PushEvent, got %T", event)
	}

	// TODO
	return nil
}
