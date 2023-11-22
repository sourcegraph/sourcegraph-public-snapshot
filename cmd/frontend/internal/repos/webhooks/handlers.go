package webhooks

import (
	"context"

	gh "github.com/google/go-github/v55/github"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/repos/webhooks/resolvers"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/cloneurls"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	gitlabwebhooks "github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Init initializes the given enterpriseServices with the webhook handlers for
// handling push events.
func Init(
	_ context.Context,
	_ *observation.Context,
	db database.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
) error {
	enterpriseServices.ReposGithubWebhook = NewGitHubHandler()
	enterpriseServices.ReposGitLabWebhook = NewGitLabHandler()
	enterpriseServices.ReposBitbucketServerWebhook = NewBitbucketServerHandler()
	enterpriseServices.ReposBitbucketCloudWebhook = NewBitbucketCloudHandler()

	enterpriseServices.WebhooksResolver = resolvers.NewWebhooksResolver(db)
	return nil
}

type GitHubHandler struct {
	logger log.Logger
}

func NewGitHubHandler() *GitHubHandler {
	return &GitHubHandler{
		logger: log.Scoped("webhooks.GitHubHandler"),
	}
}

func (g *GitHubHandler) Register(router *webhooks.Router) {
	router.Register(func(ctx context.Context, db database.DB, _ extsvc.CodeHostBaseURL, payload any) error {
		return g.handlePushEvent(ctx, db, payload)
	}, extsvc.KindGitHub, "push")
}

func (g *GitHubHandler) handlePushEvent(ctx context.Context, db database.DB, payload any) error {
	return handlePushEvent[*gh.PushEvent](ctx, db, g.logger, payload, gitHubCloneURLFromEvent)
}

func gitHubCloneURLFromEvent(event *gh.PushEvent) (string, error) {
	if event == nil || event.Repo == nil || event.Repo.CloneURL == nil {
		return "", errors.New("URL for repository not found")
	}
	return event.GetRepo().GetCloneURL(), nil
}

type GitLabHandler struct {
	logger log.Logger
}

func NewGitLabHandler() *GitLabHandler {
	return &GitLabHandler{
		logger: log.Scoped("webhooks.GitLabHandler"),
	}
}

func (g *GitLabHandler) Register(router *webhooks.Router) {
	router.Register(func(ctx context.Context, db database.DB, _ extsvc.CodeHostBaseURL, payload any) error {
		return g.handlePushEvent(ctx, db, payload)
	}, extsvc.KindGitLab, "push")
}

func (g *GitLabHandler) handlePushEvent(ctx context.Context, db database.DB, payload any) error {
	return handlePushEvent[*gitlabwebhooks.PushEvent](ctx, db, g.logger, payload, gitLabCloneURLFromEvent)
}

func gitLabCloneURLFromEvent(event *gitlabwebhooks.PushEvent) (string, error) {
	if event == nil {
		return "", errors.New("nil PushEvent received")
	}
	return event.Repository.GitSSHURL, nil
}

type BitbucketServerHandler struct {
	logger log.Logger
}

func NewBitbucketServerHandler() *BitbucketServerHandler {
	return &BitbucketServerHandler{
		logger: log.Scoped("webhooks.BitbucketServerHandler"),
	}
}

func (g *BitbucketServerHandler) Register(router *webhooks.Router) {
	router.Register(func(ctx context.Context, db database.DB, _ extsvc.CodeHostBaseURL, payload any) error {
		return g.handlePushEvent(ctx, db, payload)
	}, extsvc.KindBitbucketServer, "repo:refs_changed")
}

func (g *BitbucketServerHandler) handlePushEvent(ctx context.Context, db database.DB, payload any) error {
	return handlePushEvent[*bitbucketserver.PushEvent](ctx, db, g.logger, payload, bitbucketServerCloneURLFromEvent)
}

func bitbucketServerCloneURLFromEvent(event *bitbucketserver.PushEvent) (string, error) {
	if event == nil {
		return "", errors.New("nil PushEvent received")
	}
	for _, link := range event.Repository.Links.Clone {
		// The ssh link is the closest to our repo name
		if link.Name != "ssh" {
			continue
		}
		return link.Href, nil
	}
	return "", errors.New("no ssh URLs found")
}

type BitbucketCloudHandler struct {
	logger log.Logger
}

func NewBitbucketCloudHandler() *BitbucketCloudHandler {
	return &BitbucketCloudHandler{
		logger: log.Scoped("webhooks.BitbucketCloudHandler"),
	}
}

func (g *BitbucketCloudHandler) Register(router *webhooks.Router) {
	router.Register(func(ctx context.Context, db database.DB, _ extsvc.CodeHostBaseURL, payload any) error {
		return g.handlePushEvent(ctx, db, payload)
	}, extsvc.KindBitbucketCloud, "repo:push")
}

func (g *BitbucketCloudHandler) handlePushEvent(ctx context.Context, db database.DB, payload any) error {
	return handlePushEvent[*bitbucketcloud.PushEvent](ctx, db, g.logger, payload, bitbucketCloudCloneURLFromEvent)
}

func bitbucketCloudCloneURLFromEvent(event *bitbucketcloud.PushEvent) (string, error) {
	if event == nil {
		return "", errors.New("nil PushEvent received")
	}
	href := event.Repository.Links.HTML.Href
	if href == "" {
		return "", errors.New("clone url is empty")
	}
	return href, nil
}

// handlePushEvent takes a push payload and a function to extract the repo
// clone URL from the event. It then uses the clone URL to find a repo and queues
// a repo update.
func handlePushEvent[T any](ctx context.Context, db database.DB, logger log.Logger, payload any, cloneURLGetter func(event T) (string, error)) error {
	event, ok := payload.(T)
	if !ok {
		return errors.Newf("incorrect event type: %T", payload)
	}

	cloneURL, err := cloneURLGetter(event)
	if err != nil {
		return errors.Wrap(err, "getting clone URL from event")
	}

	repoName, err := cloneurls.RepoSourceCloneURLToRepoName(ctx, db, cloneURL)
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
			logger.Warn("push webhook received for unknown repo", log.String("repo", string(repoName)))
			return nil
		}
		return errors.Wrap(err, "handlePushEvent: EnqueueRepoUpdate failed")
	}

	logger.Info("successfully updated", log.String("name", resp.Name))
	return nil
}
