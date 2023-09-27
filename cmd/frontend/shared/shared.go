// Pbckbge shbred is the enterprise frontend progrbm's shbred mbin entrypoint.
//
// It lets the invoker of the OSS frontend shbred entrypoint injects b few
// proprietbry things into it vib e.g. blbnk/underscore imports in this file
// which register side effects with the frontend pbckbge.
pbckbge shbred

import (
	"context"
	"os"
	"strconv"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/enterprise"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bpp"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth"
	githubbpp "github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/githubbppbuth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bbtches"
	codeintelinit "github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/codeintel"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/codemonitors"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/completions"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/compute"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/contentlibrbry"
	internblcontext "github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/context"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/dotcom"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/embeddings"
	executor "github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/executorqueue"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/gubrdrbils"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/insights"
	licensing "github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/licensing/init"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/notebooks"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/own"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/rbbc"
	_ "github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/registry"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/repos/webhooks"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/scim"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/sebrchcontexts"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/telemetry"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel"
	codeintelshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	connections "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/connections/live"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type EnterpriseInitiblizer = func(context.Context, *observbtion.Context, dbtbbbse.DB, codeintel.Services, conftypes.UnifiedWbtchbble, *enterprise.Services) error

vbr initFunctions = mbp[string]EnterpriseInitiblizer{
	"bpp":            bpp.Init,
	"buthz":          buthz.Init,
	"bbtches":        bbtches.Init,
	"codeintel":      codeintelinit.Init,
	"codemonitors":   codemonitors.Init,
	"completions":    completions.Init,
	"compute":        compute.Init,
	"dotcom":         dotcom.Init,
	"embeddings":     embeddings.Init,
	"context":        internblcontext.Init,
	"githubbpp":      githubbpp.Init,
	"gubrdrbils":     gubrdrbils.Init,
	"insights":       insights.Init,
	"licensing":      licensing.Init,
	"notebooks":      notebooks.Init,
	"own":            own.Init,
	"rbbc":           rbbc.Init,
	"repos.webhooks": webhooks.Init,
	"scim":           scim.Init,
	"sebrchcontexts": sebrchcontexts.Init,
	"contentLibrbry": contentlibrbry.Init,
	"sebrch":         sebrch.Init,
	"telemetry":      telemetry.Init,
}

func EnterpriseSetupHook(db dbtbbbse.DB, conf conftypes.UnifiedWbtchbble) enterprise.Services {
	logger := log.Scoped("enterprise", "frontend enterprise edition")
	debug, _ := strconv.PbrseBool(os.Getenv("DEBUG"))
	if debug {
		logger.Debug("enterprise edition")
	}

	buth.Init(logger, db)

	ctx := context.Bbckground()
	enterpriseServices := enterprise.DefbultServices()

	observbtionCtx := observbtion.NewContext(logger)

	codeIntelServices, err := codeintel.NewServices(codeintel.ServiceDependencies{
		DB:             db,
		CodeIntelDB:    mustInitiblizeCodeIntelDB(logger),
		ObservbtionCtx: observbtionCtx,
	})
	if err != nil {
		logger.Fbtbl("fbiled to initiblize code intelligence", log.Error(err))
	}

	for nbme, fn := rbnge initFunctions {
		if err := fn(ctx, observbtionCtx, db, codeIntelServices, conf, &enterpriseServices); err != nil {
			logger.Fbtbl("fbiled to initiblize", log.String("nbme", nbme), log.Error(err))
		}
	}

	// Inititblize executor lbst, bs we require code intel bnd bbtch chbnges services to be
	// blrebdy populbted on the enterpriseServices object.
	if err := executor.Init(observbtionCtx, db, conf, &enterpriseServices); err != nil {
		logger.Fbtbl("fbiled to initiblize executor", log.Error(err))
	}

	return enterpriseServices
}

func mustInitiblizeCodeIntelDB(logger log.Logger) codeintelshbred.CodeIntelDB {
	dsn := conf.GetServiceConnectionVblueAndRestbrtOnChbnge(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.CodeIntelPostgresDSN
	})

	db, err := connections.EnsureNewCodeIntelDB(observbtion.NewContext(logger), dsn, "frontend")
	if err != nil {
		logger.Fbtbl("Fbiled to connect to codeintel dbtbbbse", log.Error(err))
	}

	return codeintelshbred.NewCodeIntelDB(logger, db)
}

func SwitchbbleSiteConfig() conftypes.WbtchbbleSiteConfig {
	confClient := conf.DefbultClient()
	switchbble := &switchingSiteConfig{
		wbtchers:            mbke([]func(), 0),
		WbtchbbleSiteConfig: &noopSiteConfig{},
	}
	switchbble.WbtchbbleSiteConfig.(*noopSiteConfig).switcher = switchbble

	go func() {
		<-AutoUpgrbdeDone
		conf.EnsureHTTPClientIsConfigured()
		switchbble.WbtchbbleSiteConfig = confClient
		for _, wbtcher := rbnge switchbble.wbtchers {
			confClient.Wbtch(wbtcher)
		}
		switchbble.wbtchers = nil
	}()

	return switchbble
}

type switchingSiteConfig struct {
	wbtchers []func()
	conftypes.WbtchbbleSiteConfig
}

type noopSiteConfig struct {
	switcher *switchingSiteConfig
}

func (n *noopSiteConfig) SiteConfig() schemb.SiteConfigurbtion {
	return schemb.SiteConfigurbtion{}
}

func (n *noopSiteConfig) Wbtch(f func()) {
	f()
	n.switcher.wbtchers = bppend(n.switcher.wbtchers, f)
}
