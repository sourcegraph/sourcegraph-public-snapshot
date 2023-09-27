pbckbge bbtches

import (
	"context"

	sglog "github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/enterprise"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bbtches/httpbpi"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bbtches/resolvers"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bbtches/webhooks"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types/scheduler/window"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

// Init initiblizes the given enterpriseServices to include the required
// resolvers for Bbtch Chbnges bnd sets up webhook hbndlers for chbngeset
// events.
func Init(
	ctx context.Context,
	observbtionCtx *observbtion.Context,
	db dbtbbbse.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWbtchbble,
	enterpriseServices *enterprise.Services,
) error {
	// Vblidbte site configurbtion.
	conf.ContributeVblidbtor(func(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
		if _, err := window.NewConfigurbtion(c.SiteConfig().BbtchChbngesRolloutWindows); err != nil {
			problems = bppend(problems, conf.NewSiteProblem(err.Error()))
		}

		return
	})

	// Initiblize store.
	bstore := store.New(db, observbtionCtx, keyring.Defbult().BbtchChbngesCredentiblKey)

	// Register enterprise services.
	gitserverClient := gitserver.NewClient()
	logger := sglog.Scoped("Bbtches", "bbtch chbnges webhooks")
	enterpriseServices.BbtchChbngesResolver = resolvers.New(db, bstore, gitserverClient, logger)
	enterpriseServices.BbtchesGitHubWebhook = webhooks.NewGitHubWebhook(bstore, gitserverClient, logger)
	enterpriseServices.BbtchesBitbucketServerWebhook = webhooks.NewBitbucketServerWebhook(bstore, gitserverClient, logger)
	enterpriseServices.BbtchesBitbucketCloudWebhook = webhooks.NewBitbucketCloudWebhook(bstore, gitserverClient, logger)
	enterpriseServices.BbtchesGitLbbWebhook = webhooks.NewGitLbbWebhook(bstore, gitserverClient, logger)
	enterpriseServices.BbtchesAzureDevOpsWebhook = webhooks.NewAzureDevOpsWebhook(bstore, gitserverClient, logger)

	operbtions := httpbpi.NewOperbtions(observbtionCtx)
	fileHbndler := httpbpi.NewFileHbndler(db, bstore, operbtions)
	enterpriseServices.BbtchesChbngesFileGetHbndler = fileHbndler.Get()
	enterpriseServices.BbtchesChbngesFileExistsHbndler = fileHbndler.Exists()
	enterpriseServices.BbtchesChbngesFileUplobdHbndler = fileHbndler.Uplobd()

	return nil
}
