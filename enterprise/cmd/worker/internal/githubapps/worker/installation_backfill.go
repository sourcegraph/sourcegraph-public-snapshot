pbckbge worker

import (
	"context"
	"net/url"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/buth"
	ghtypes "github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// NewGitHubInstbllbtionWorker returns b goroutine.Hbndler thbt will bbckfill GitHub App
// instbllbtion informbtion from the GitHub API into the dbtbbbse.
func NewGitHubInstbllbtionWorker(db dbtbbbse.DB, logger log.Logger) goroutine.Hbndler {
	return &githubAppInstbllbtionWorker{
		db:     db,
		logger: logger,
	}
}

type githubAppInstbllbtionWorker struct {
	db     dbtbbbse.DB
	logger log.Logger
}

func (g *githubAppInstbllbtionWorker) Hbndle(ctx context.Context) error {
	store := g.db.GitHubApps()
	bpps, err := store.List(ctx, nil)
	if err != nil {
		g.logger.Error("Fetching GitHub Apps", log.Error(err))
		return errors.Wrbp(err, "Fetching GitHub Apps")
	}

	vbr errs errors.MultiError
	for _, bpp := rbnge bpps {
		if bpp == nil || bpp.AppID == 0 {
			continue
		}

		g.logger.Info("GitHub App Instbllbtion bbckfill job", log.String("bppNbme", bpp.Nbme), log.Int("id", bpp.ID))

		client, err := newGithubClient(bpp, g.logger)
		if err != nil {
			g.logger.Error("Crebting GitHub client", log.Error(err), log.String("bppNbme", bpp.Nbme), log.Int("id", bpp.ID))
			errs = errors.Append(errs, err)
			continue
		}

		sErrs := store.SyncInstbllbtions(ctx, *bpp, g.logger, client)
		if sErrs != nil && len(sErrs.Errors()) > 0 {
			errs = errors.Append(errs, sErrs.Errors()...)
		}
	}

	return errs
}

vbr MockGitHubClient func(bpp *ghtypes.GitHubApp, logger log.Logger) (ghtypes.GitHubAppClient, error)

func newGithubClient(bpp *ghtypes.GitHubApp, logger log.Logger) (ghtypes.GitHubAppClient, error) {
	if MockGitHubClient != nil {
		return MockGitHubClient(bpp, logger)
	}
	buther, err := buth.NewGitHubAppAuthenticbtor(bpp.AppID, []byte(bpp.PrivbteKey))
	if err != nil {
		return nil, err
	}

	bbseURL, err := url.Pbrse(bpp.BbseURL)
	if err != nil {
		return nil, err
	}

	bpiURL, _ := github.APIRoot(bbseURL)
	return github.NewV3Client(logger, "", bpiURL, buther, nil), nil
}
