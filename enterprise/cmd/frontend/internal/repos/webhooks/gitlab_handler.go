package webhooks

import (
	"context"
	"fmt"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/cloneurls"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	gitlabwebhooks "github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitLabHandler struct {
	db     database.DB
	logger log.Logger
}

func (g *GitLabHandler) Register(router *webhooks.Router) {
	router.Register(func(ctx context.Context, _ database.DB, _ extsvc.CodeHostBaseURL, payload any) error {
		return g.handlePushEvent(ctx, payload)
	}, extsvc.KindGitLab, "push")
}

func NewGitLabHandler(db database.DB) *GitLabHandler {
	return &GitLabHandler{
		db:     db,
		logger: log.Scoped("webhooks.GitLabHandler", "gitlab webhook handler"),
	}
}

func (g *GitLabHandler) handlePushEvent(ctx context.Context, payload any) error {
	event, ok := payload.(*gitlabwebhooks.PushEvent)
	if !ok {
		return errors.Newf("expected GitLab.PushEvent, got %T", payload)
	}

	cloneURL, err := gitLabCloneURLFromEvent(event)
	if err != nil {
		return errors.Wrap(err, "getting clone URL from event")
	}
	fmt.Println("cloneURL:", cloneURL)
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
			g.logger.Warn("GitLab push webhook received for unknown repo", log.String("repo", string(repoName)))
			return nil
		}
		return errors.Wrap(err, "handlePushEvent: EnqueueRepoUpdate failed")
	}

	g.logger.Info("successfully updated", log.String("name", resp.Name))
	return nil
}

func gitLabCloneURLFromEvent(event *gitlabwebhooks.PushEvent) (string, error) {
	if event == nil {
		return "", errors.New("nil PushEvent received")
	}
	return event.Repository.GitSSHURL, nil
}
