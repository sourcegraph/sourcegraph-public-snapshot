package repos

import (
	"context"
	"reflect"

	gh "github.com/google/go-github/v43/github"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitHubWebhookAPI struct {
	Client *github.V3Client
}

func NewGitHubWebhookAPI(client *github.V3Client) *GitHubWebhookAPI {
	return &GitHubWebhookAPI{Client: client}
}

func (g *GitHubWebhookAPI) Register(router *webhooks.GitHubWebhook) {
	router.Register(g.handleGitHubWebhook, "push")
}

func (g *GitHubWebhookAPI) handleGitHubWebhook(ctx context.Context, extSvc *types.ExternalService, payload any) error {
	event, ok := payload.(*gh.PushEvent)
	if !ok {
		return errors.Errorf("expected GitHub.PushEvent, got %s", reflect.TypeOf(event))
	}

	repoName := *event.Repo.URL
	name := api.RepoName(repoName[8:])
	repoupdater.DefaultClient.EnqueueRepoUpdate(ctx, name)

	return nil
}
