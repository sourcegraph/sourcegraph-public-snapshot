package webhooks

import (
	"context"
	"net/url"

	gh "github.com/google/go-github/v43/github"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitHubHandler struct {
	logger log.Logger
}

func (g *GitHubHandler) Register(router *webhooks.Router) {
	router.Register(func(ctx context.Context, _ database.DB, _ extsvc.CodeHostBaseURL, payload any) error {
		return g.handlePushEvent(ctx, payload)
	}, extsvc.KindGitHub, "push")
}

func NewGitHubHandler() *GitHubHandler {
	return &GitHubHandler{
		logger: log.Scoped("webhooks.GitHubHandler", "github webhook handler"),
	}
}

func (g *GitHubHandler) handlePushEvent(ctx context.Context, payload any) error {
	event, ok := payload.(*gh.PushEvent)
	if !ok {
		return errors.Newf("expected GitHub.PushEvent, got %T", payload)
	}

	repoName, err := githubNameFromEvent(event)
	if err != nil {
		return errors.Wrap(err, "handlePushEvent: get name failed")
	}

	resp, err := repoupdater.DefaultClient.EnqueueRepoUpdate(ctx, repoName)
	if err != nil {
		// Repo not existing on Sourcegraph is fine
		if errcode.IsNotFound(err) {
			g.logger.Warn("GitHub push webhook received for unknown repo", log.String("repo", string(repoName)))
			return nil
		}
		return errors.Wrap(err, "handlePushEvent: EnqueueRepoUpdate failed")
	}

	g.logger.Info("successfully updated", log.String("name", resp.Name))
	return nil
}

func githubNameFromEvent(event *gh.PushEvent) (api.RepoName, error) {
	if event == nil || event.Repo == nil || event.Repo.URL == nil {
		return "", errors.New("URL for repository not found")
	}
	parsed, err := url.Parse(*event.Repo.URL)
	if err != nil {
		return "", errors.Wrap(err, "unable to parse repository URL")
	}
	return api.RepoName(parsed.Host + parsed.Path), nil
}
