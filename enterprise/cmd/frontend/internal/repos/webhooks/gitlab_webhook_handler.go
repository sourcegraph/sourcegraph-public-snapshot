package webhooks

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	gitlabwebhooks "github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitLabWebhookHandler struct {
	logger log.Logger
}

func (g *GitLabWebhookHandler) Register(router *webhooks.WebhookRouter) {
	router.Register(func(ctx context.Context, _ database.DB, _ extsvc.CodeHostBaseURL, payload any) error {
		return g.handle(ctx, payload)
	}, extsvc.KindGitLab, "push")
}

func NewGitLabWebhookHandler() *GitLabWebhookHandler {
	return &GitLabWebhookHandler{
		logger: log.Scoped("repos.GitLabWebhookHandler", "gitlab webhook handler"),
	}
}

func (g *GitLabWebhookHandler) handle(ctx context.Context, payload any) error {
	event, ok := payload.(*gitlabwebhooks.PushEvent)
	if !ok {
		return errors.Newf("expected GitLab.PushEvent, got %T", payload)
	}

	repoName, err := gitlabNameFromEvent(event)
	if err != nil {
		return errors.Wrap(err, "handleGitLabWebhook: get name failed")
	}

	resp, err := repoupdater.DefaultClient.EnqueueRepoUpdate(ctx, repoName)
	if err != nil {
		// Repo not existing on Sourcegraph is fine
		if errcode.IsNotFound(err) {
			return nil
		}
		return errors.Wrap(err, "handleGitLabWebhook: EnqueueRepoUpdate failed")
	}

	g.logger.Info("successfully updated", log.String("name", resp.Name))
	return nil
}

func gitlabNameFromEvent(event *gitlabwebhooks.PushEvent) (api.RepoName, error) {
	url := event.Project.WebURL
	if len(url) <= 8 {
		return "", errors.Newf("expected URL length > 8, got %v", len(url))
	}
	repoName := url[8:] // [ https:// ] accounts for 8 chars
	return api.RepoName(repoName), nil
}
