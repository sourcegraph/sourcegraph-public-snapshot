package repos

import (
	"context"
	"fmt"

	gh "github.com/google/go-github/v43/github"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitHubWebhookHandler struct {
	client *github.V3Client
}

func newGitHubWebhookHandler(client *github.V3Client) *GitHubWebhookHandler {
	return &GitHubWebhookHandler{client: client}
}

func handleGitHubWebhook(ctx context.Context, extSvc *types.ExternalService, payload any) error {
	event, ok := payload.(*gh.PushEvent)
	if !ok {
		return errors.Newf("expected GitHub.PushEvent, got %T", payload)
	}

	notify = func(ch chan struct{}) {
		select {
		case ch <- struct{}{}:
		default:
		}
	}

	fullName := *event.Repo.URL
	repoName := api.RepoName(fullName[8:])

	resp, err := repoupdater.DefaultClient.EnqueueRepoUpdate(ctx, repoName)
	if err != nil {
		return err
	}

	log.Scoped("GitHub handler", fmt.Sprintf("Successfully updated: %s", resp.Name))
	// TODO
	return nil
}

func (g *GitHubWebhookHandler) CreateSyncWebhook(ctx context.Context, repoName, targetURL, secret string) (int, error) {
	return g.client.CreateSyncWebhook(ctx, repoName, targetURL, secret)
}

func (g *GitHubWebhookHandler) ListSyncWebhook(ctx context.Context, repoName string) ([]github.WebhookPayload, error) {
	return g.client.ListSyncWebhooks(ctx, repoName)
}

func (g *GitHubWebhookHandler) FindSyncWebhook(ctx context.Context, repoName string) (int, error) {
	return g.client.FindSyncWebhook(ctx, repoName)
}

func (g *GitHubWebhookHandler) DeleteSyncWebhook(ctx context.Context, repoName string, hookID int) (bool, error) {
	return g.client.DeleteSyncWebhook(ctx, repoName, hookID)
}
