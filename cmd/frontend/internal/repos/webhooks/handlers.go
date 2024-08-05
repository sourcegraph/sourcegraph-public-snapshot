package webhooks

import (
	"context"
	"strconv"

	gh "github.com/google/go-github/v55/github"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/repos/webhooks/resolvers"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
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
	router.Register(func(ctx context.Context, db database.DB, baseURL extsvc.CodeHostBaseURL, payload any) error {
		return g.handlePushEvent(ctx, db, baseURL, payload)
	}, extsvc.KindGitHub, "push")
}

func (g *GitHubHandler) handlePushEvent(ctx context.Context, db database.DB, baseURL extsvc.CodeHostBaseURL, payload any) error {
	return handlePushEvent[*gh.PushEvent](ctx, db, g.logger, extsvc.TypeGitHub, baseURL, payload, gitHubExternalIDFromEvent)
}

func gitHubExternalIDFromEvent(event *gh.PushEvent) (string, error) {
	if event == nil || event.GetRepo() == nil || event.GetRepo().GetID() == 0 {
		return "", errors.New("external ID for repository not found")
	}
	return event.GetRepo().GetNodeID(), nil
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
	router.Register(func(ctx context.Context, db database.DB, baseURL extsvc.CodeHostBaseURL, payload any) error {
		return g.handlePushEvent(ctx, db, baseURL, payload)
	}, extsvc.KindGitLab, "push")
}

func (g *GitLabHandler) handlePushEvent(ctx context.Context, db database.DB, baseURL extsvc.CodeHostBaseURL, payload any) error {
	return handlePushEvent[*gitlabwebhooks.PushEvent](ctx, db, g.logger, extsvc.TypeGitLab, baseURL, payload, gitlabExternalIDFromEvent)
}

func gitlabExternalIDFromEvent(event *gitlabwebhooks.PushEvent) (string, error) {
	if event == nil {
		return "", errors.New("nil PushEvent received")
	}
	if event.ProjectID == 0 {
		return "", errors.New("project ID not found in PushEvent")
	}
	return strconv.Itoa(event.ProjectID), nil
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
	router.Register(func(ctx context.Context, db database.DB, baseURL extsvc.CodeHostBaseURL, payload any) error {
		return g.handlePushEvent(ctx, db, baseURL, payload)
	}, extsvc.KindBitbucketServer, "repo:refs_changed")
}

func (g *BitbucketServerHandler) handlePushEvent(ctx context.Context, db database.DB, baseURL extsvc.CodeHostBaseURL, payload any) error {
	return handlePushEvent[*bitbucketserver.PushEvent](ctx, db, g.logger, extsvc.TypeBitbucketServer, baseURL, payload, bitbucketServerExternalIDFromEvent)
}

func bitbucketServerExternalIDFromEvent(event *bitbucketserver.PushEvent) (string, error) {
	if event == nil {
		return "", errors.New("nil PushEvent received")
	}
	if event.Repository.ID == 0 {
		return "", errors.New("repository ID not found in PushEvent")
	}
	return strconv.Itoa(event.Repository.ID), nil
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
	router.Register(func(ctx context.Context, db database.DB, baseURL extsvc.CodeHostBaseURL, payload any) error {
		return g.handlePushEvent(ctx, db, baseURL, payload)
	}, extsvc.KindBitbucketCloud, "repo:push")
}

func (g *BitbucketCloudHandler) handlePushEvent(ctx context.Context, db database.DB, baseURL extsvc.CodeHostBaseURL, payload any) error {
	return handlePushEvent[*bitbucketcloud.PushEvent](ctx, db, g.logger, extsvc.TypeBitbucketCloud, baseURL, payload, bitbucketCloudExternalIDFromEvent)
}

func bitbucketCloudExternalIDFromEvent(event *bitbucketcloud.PushEvent) (string, error) {
	if event == nil {
		return "", errors.New("nil PushEvent received")
	}
	return event.Repository.UUID, nil
}

// handlePushEvent takes a push payload and a function to extract the repo
// clone URL from the event. It then uses the clone URL to find a repo and queues
// a repo update.
func handlePushEvent[T any](ctx context.Context, db database.DB, logger log.Logger, externalServiceType string, externalServiceURL extsvc.CodeHostBaseURL, payload any, externalIDGetter func(event T) (string, error)) error {
	event, ok := payload.(T)
	if !ok {
		return errors.Newf("incorrect event type: %T", payload)
	}

	externalID, err := externalIDGetter(event)
	if err != nil {
		return errors.Wrap(err, "getting external ID from event")
	}

	rs, err := db.Repos().List(ctx, database.ReposListOptions{
		ExternalRepos: []api.ExternalRepoSpec{
			{
				ID:          externalID,
				ServiceType: externalServiceType,
				ServiceID:   externalServiceURL.String(),
			},
		},
	})
	if err != nil {
		return errors.Wrap(err, "handlePushEvent: ListRepos failed")
	}

	if len(rs) == 0 {
		logger.Debug("push webhook received for unknown repo", log.String("repo", string(externalID)))
		return nil
	}

	resp, err := repoupdater.DefaultClient.EnqueueRepoUpdate(ctx, rs[0].Name)
	if err != nil {
		return errors.Wrap(err, "handlePushEvent: EnqueueRepoUpdate failed")
	}

	logger.Debug("successfully updated repo from webhook", log.String("name", resp.Name))
	return nil
}
