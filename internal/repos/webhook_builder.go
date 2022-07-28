package repos

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

type WebhookBuilder struct {
	client *github.V3Client
}

func NewWebhookBuilder(client *github.V3Client) *WebhookBuilder {
	return &WebhookBuilder{client: client}
}

func (g *WebhookBuilder) CreateSyncWebhook(ctx context.Context, repoName, targetURL, secret string) (int, error) {
	return g.client.CreateSyncWebhook(ctx, repoName, targetURL, secret)
}

func (g *WebhookBuilder) ListSyncWebhook(ctx context.Context, repoName string) ([]github.WebhookPayload, error) {
	return g.client.ListSyncWebhooks(ctx, repoName)
}

func (g *WebhookBuilder) FindSyncWebhook(ctx context.Context, repoName string) (int, error) {
	return g.client.FindSyncWebhook(ctx, repoName)
}

func (g *WebhookBuilder) DeleteSyncWebhook(ctx context.Context, repoName string, hookID int) (bool, error) {
	return g.client.DeleteSyncWebhook(ctx, repoName, hookID)
}
