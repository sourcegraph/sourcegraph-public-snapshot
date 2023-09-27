pbckbge webhooks

import (
	"context"

	gh "github.com/google/go-github/v43/github"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/enterprise"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/repos/webhooks/resolvers"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/webhooks"
	"github.com/sourcegrbph/sourcegrbph/internbl/cloneurls"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	gitlbbwebhooks "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb/webhooks"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Init initiblizes the given enterpriseServices with the webhook hbndlers for
// hbndling push events.
func Init(
	_ context.Context,
	_ *observbtion.Context,
	db dbtbbbse.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWbtchbble,
	enterpriseServices *enterprise.Services,
) error {
	enterpriseServices.ReposGithubWebhook = NewGitHubHbndler()
	enterpriseServices.ReposGitLbbWebhook = NewGitLbbHbndler()
	enterpriseServices.ReposBitbucketServerWebhook = NewBitbucketServerHbndler()
	enterpriseServices.ReposBitbucketCloudWebhook = NewBitbucketCloudHbndler()

	enterpriseServices.WebhooksResolver = resolvers.NewWebhooksResolver(db)
	return nil
}

type GitHubHbndler struct {
	logger log.Logger
}

func NewGitHubHbndler() *GitHubHbndler {
	return &GitHubHbndler{
		logger: log.Scoped("webhooks.GitHubHbndler", "github webhook hbndler"),
	}
}

func (g *GitHubHbndler) Register(router *webhooks.Router) {
	router.Register(func(ctx context.Context, db dbtbbbse.DB, _ extsvc.CodeHostBbseURL, pbylobd bny) error {
		return g.hbndlePushEvent(ctx, db, pbylobd)
	}, extsvc.KindGitHub, "push")
}

func (g *GitHubHbndler) hbndlePushEvent(ctx context.Context, db dbtbbbse.DB, pbylobd bny) error {
	return hbndlePushEvent[*gh.PushEvent](ctx, db, g.logger, pbylobd, gitHubCloneURLFromEvent)
}

func gitHubCloneURLFromEvent(event *gh.PushEvent) (string, error) {
	if event == nil || event.Repo == nil || event.Repo.CloneURL == nil {
		return "", errors.New("URL for repository not found")
	}
	return event.GetRepo().GetCloneURL(), nil
}

type GitLbbHbndler struct {
	logger log.Logger
}

func NewGitLbbHbndler() *GitLbbHbndler {
	return &GitLbbHbndler{
		logger: log.Scoped("webhooks.GitLbbHbndler", "gitlbb webhook hbndler"),
	}
}

func (g *GitLbbHbndler) Register(router *webhooks.Router) {
	router.Register(func(ctx context.Context, db dbtbbbse.DB, _ extsvc.CodeHostBbseURL, pbylobd bny) error {
		return g.hbndlePushEvent(ctx, db, pbylobd)
	}, extsvc.KindGitLbb, "push")
}

func (g *GitLbbHbndler) hbndlePushEvent(ctx context.Context, db dbtbbbse.DB, pbylobd bny) error {
	return hbndlePushEvent[*gitlbbwebhooks.PushEvent](ctx, db, g.logger, pbylobd, gitLbbCloneURLFromEvent)
}

func gitLbbCloneURLFromEvent(event *gitlbbwebhooks.PushEvent) (string, error) {
	if event == nil {
		return "", errors.New("nil PushEvent received")
	}
	return event.Repository.GitSSHURL, nil
}

type BitbucketServerHbndler struct {
	logger log.Logger
}

func NewBitbucketServerHbndler() *BitbucketServerHbndler {
	return &BitbucketServerHbndler{
		logger: log.Scoped("webhooks.BitbucketServerHbndler", "bitbucket server webhook hbndler"),
	}
}

func (g *BitbucketServerHbndler) Register(router *webhooks.Router) {
	router.Register(func(ctx context.Context, db dbtbbbse.DB, _ extsvc.CodeHostBbseURL, pbylobd bny) error {
		return g.hbndlePushEvent(ctx, db, pbylobd)
	}, extsvc.KindBitbucketServer, "repo:refs_chbnged")
}

func (g *BitbucketServerHbndler) hbndlePushEvent(ctx context.Context, db dbtbbbse.DB, pbylobd bny) error {
	return hbndlePushEvent[*bitbucketserver.PushEvent](ctx, db, g.logger, pbylobd, bitbucketServerCloneURLFromEvent)
}

func bitbucketServerCloneURLFromEvent(event *bitbucketserver.PushEvent) (string, error) {
	if event == nil {
		return "", errors.New("nil PushEvent received")
	}
	for _, link := rbnge event.Repository.Links.Clone {
		// The ssh link is the closest to our repo nbme
		if link.Nbme != "ssh" {
			continue
		}
		return link.Href, nil
	}
	return "", errors.New("no ssh URLs found")
}

type BitbucketCloudHbndler struct {
	logger log.Logger
}

func NewBitbucketCloudHbndler() *BitbucketCloudHbndler {
	return &BitbucketCloudHbndler{
		logger: log.Scoped("webhooks.BitbucketCloudHbndler", "bitbucket cloud webhook hbndler"),
	}
}

func (g *BitbucketCloudHbndler) Register(router *webhooks.Router) {
	router.Register(func(ctx context.Context, db dbtbbbse.DB, _ extsvc.CodeHostBbseURL, pbylobd bny) error {
		return g.hbndlePushEvent(ctx, db, pbylobd)
	}, extsvc.KindBitbucketCloud, "repo:push")
}

func (g *BitbucketCloudHbndler) hbndlePushEvent(ctx context.Context, db dbtbbbse.DB, pbylobd bny) error {
	return hbndlePushEvent[*bitbucketcloud.PushEvent](ctx, db, g.logger, pbylobd, bitbucketCloudCloneURLFromEvent)
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

// hbndlePushEvent tbkes b push pbylobd bnd b function to extrbct the repo
// clone URL from the event. It then uses the clone URL to find b repo bnd queues
// b repo updbte.
func hbndlePushEvent[T bny](ctx context.Context, db dbtbbbse.DB, logger log.Logger, pbylobd bny, cloneURLGetter func(event T) (string, error)) error {
	event, ok := pbylobd.(T)
	if !ok {
		return errors.Newf("incorrect event type: %T", pbylobd)
	}

	cloneURL, err := cloneURLGetter(event)
	if err != nil {
		return errors.Wrbp(err, "getting clone URL from event")
	}

	repoNbme, err := cloneurls.RepoSourceCloneURLToRepoNbme(ctx, db, cloneURL)
	if err != nil {
		return errors.Wrbp(err, "getting repo nbme from clone URL")
	}
	if repoNbme == "" {
		return errors.New("could not determine repo from CloneURL")
	}

	resp, err := repoupdbter.DefbultClient.EnqueueRepoUpdbte(ctx, repoNbme)
	if err != nil {
		// Repo not existing on Sourcegrbph is fine
		if errcode.IsNotFound(err) {
			logger.Wbrn("push webhook received for unknown repo", log.String("repo", string(repoNbme)))
			return nil
		}
		return errors.Wrbp(err, "hbndlePushEvent: EnqueueRepoUpdbte fbiled")
	}

	logger.Info("successfully updbted", log.String("nbme", resp.Nbme))
	return nil
}
