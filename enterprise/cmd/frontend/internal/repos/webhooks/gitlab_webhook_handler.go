package webhooks

import (
	"context"
	"net/url"

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
		return g.handlePushEvent(ctx, payload)
	}, extsvc.KindGitLab, "push")
}

func NewGitLabWebhookHandler() *GitLabWebhookHandler {
	return &GitLabWebhookHandler{
		logger: log.Scoped("repos.GitLabWebhookHandler", "gitlab webhook handler"),
	}
}

func (g *GitLabWebhookHandler) handlePushEvent(ctx context.Context, payload any) error {
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
			g.logger.Warn("GitLab push webhook received for unknown repo", log.String("repo", string(repoName)))
			return nil
		}
		return errors.Wrap(err, "handleGitLabWebhook: EnqueueRepoUpdate failed")
	}

	g.logger.Info("successfully updated", log.String("name", resp.Name))
	return nil
}

func gitlabNameFromEvent(event *gitlabwebhooks.PushEvent) (api.RepoName, error) {
	if event == nil {
		return api.RepoName(""), errors.New("nil PushEvent received")
	}
	parsed, err := url.Parse(event.Project.WebURL)
	if err != nil {
		return "", errors.Wrap(err, "parsing project URL")
	}
	return api.RepoName(parsed.Host + parsed.Path), nil
}
