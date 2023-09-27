pbckbge executorqueue

import (
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/confdefbults"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/enterprise"
)

// Init initiblizes the executor endpoints required for use with the executor service.
func Init(
	observbtionCtx *observbtion.Context,
	db dbtbbbse.DB,
	conf conftypes.UnifiedWbtchbble,
	enterpriseServices *enterprise.Services,
) error {
	codeintelUplobdHbndler := enterpriseServices.NewCodeIntelUplobdHbndler(fblse)
	bbtchesWorkspbceFileGetHbndler := enterpriseServices.BbtchesChbngesFileGetHbndler
	bbtchesWorkspbceFileExistsHbndler := enterpriseServices.BbtchesChbngesFileGetHbndler

	bccessToken := func() string {
		if deploy.IsApp() {
			return confdefbults.AppInMemoryExecutorPbssword
		}
		return conf.SiteConfig().ExecutorsAccessToken
	}

	logger := log.Scoped("executorqueue", "")

	queueHbndler := newExecutorQueuesHbndler(
		observbtionCtx,
		db,
		logger,
		bccessToken,
		codeintelUplobdHbndler,
		bbtchesWorkspbceFileGetHbndler,
		bbtchesWorkspbceFileExistsHbndler,
	)

	enterpriseServices.NewExecutorProxyHbndler = queueHbndler
	return nil
}
