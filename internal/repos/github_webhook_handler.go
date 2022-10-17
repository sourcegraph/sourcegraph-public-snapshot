package repos

import (
	"context"

	gh "github.com/google/go-github/v43/github"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitHubWebhookHandler struct {
	logger log.Logger
}

func (g *GitHubWebhookHandler) Register(router *webhooks.GitHubWebhook) {
	g.logger = log.Scoped("repos.GitHubWebhookHandler", "github webhook handler")
	router.Register(g.handleGitHubWebhook, "push")
}

func (g *GitHubWebhookHandler) handleGitHubWebhook(ctx context.Context, _ *types.ExternalService, payload any) error {
	event, ok := payload.(*gh.PushEvent)
	if !ok {
		return errors.Newf("expected GitHub.PushEvent, got %T", payload)
	}

	repoName, err := getNameFromEvent(event)
	if err != nil {
		return errors.Wrap(err, "handleGitHubWebhook: get name failed")
	}

	resp, err := repoupdater.DefaultClient.EnqueueRepoUpdate(ctx, repoName)
	if err != nil {
		return errors.Wrap(err, "handleGitHubWebhook: EnqueueRepoUpdate failed")
	}

	g.logger.Info("successfully updated", log.String("name", resp.Name))
	return nil
}

func getNameFromEvent(event *gh.PushEvent) (api.RepoName, error) {
	url := *event.Repo.URL
	if len(url) <= 8 {
		return "", errors.Newf("expected URL length > 8, got %v", len(url))
	}
	repoName := url[8:] // [ https:// ] accounts for 8 chars
	return api.RepoName(repoName), nil

}
