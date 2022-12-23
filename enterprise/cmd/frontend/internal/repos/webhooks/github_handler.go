package webhooks

import (
	"context"

	gh "github.com/google/go-github/v43/github"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/cloneurls"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitHubHandler struct {
	db     database.DB
	logger log.Logger
}

func (g *GitHubHandler) Register(router *webhooks.Router) {
	router.Register(func(ctx context.Context, _ database.DB, _ extsvc.CodeHostBaseURL, payload any) error {
		return g.handlePushEvent(ctx, payload)
	}, extsvc.KindGitHub, "push")
}

func NewGitHubHandler(db database.DB) *GitHubHandler {
	return &GitHubHandler{
		db:     db,
		logger: log.Scoped("webhooks.GitHubHandler", "github webhook handler"),
	}
}

func (g *GitHubHandler) handlePushEvent(ctx context.Context, payload any) error {
	event, ok := payload.(*gh.PushEvent)
	if !ok {
		return errors.Newf("expected GitHub.PushEvent, got %T", payload)
	}

	cloneURL, err := gitHubCloneURLFromEvent(event)
	if err != nil {
		return errors.Wrap(err, "getting clone URL from event")
	}
	repoName, err := cloneurls.RepoSourceCloneURLToRepoName(ctx, g.db, cloneURL)
	if err != nil {
		return errors.Wrap(err, "getting repo name from clone URL")
	}
	if repoName == "" {
		return errors.New("could not determine repo from CloneURL")
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

func gitHubCloneURLFromEvent(event *gh.PushEvent) (string, error) {
	if event == nil || event.Repo == nil || event.Repo.CloneURL == nil {
		return "", errors.New("URL for repository not found")
	}
	return event.GetRepo().GetCloneURL(), nil
}
