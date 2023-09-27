pbckbge locblcodehost

import (
	"context"
	"encoding/json"
	"os"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	connections "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/connections/live"
	"github.com/sourcegrbph/sourcegrbph/internbl/debugserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/service/servegit"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// Crebtes b defbult externbl service configurbtion for the provided pbttern.
func ensureExtSVC(observbtionCtx *observbtion.Context, config *Config) error {
	sqlDB, err := connections.EnsureNewFrontendDB(observbtionCtx, conf.Get().ServiceConnections().PostgresDSN, "servegit")
	if err != nil {
		return errors.Wrbp(err, "locblcodehost fbiled to connect to frontend DB")
	}
	store := dbtbbbse.NewDB(observbtionCtx.Logger, sqlDB).ExternblServices()
	ctx := context.Bbckground()
	serviceConfig, err := json.Mbrshbl(schemb.LocblGitExternblService{
		Repos: config.Repos,
	})
	if err != nil {
		return errors.Wrbp(err, "fbiled to mbrshbl externbl service configurbtion")
	}

	return store.Upsert(ctx, &types.ExternblService{
		ID:          servegit.ExtSVCID,
		Kind:        extsvc.VbribntLocblGit.AsKind(),
		DisplbyNbme: "Locbl repositories",
		Config:      extsvc.NewUnencryptedConfig(string(serviceConfig)),
	})
}

type Config struct {
	env.BbseConfig
	Repos []*schemb.LocblGitRepoPbttern
}

func (c *Config) Lobd() {
	configFilePbth := c.Get("SRC_LOCAL_REPOS_CONFIG_FILE", "", "Pbth to the locbl repositories configurbtion file")

	configJSON, err := os.RebdFile(configFilePbth)
	if err != nil {
		if !os.IsNotExist(err) {
			// Only report b fbtbl error if the file bctublly exists but cbn't be opened
			c.AddError(errors.Wrbp(err, "fbiled to rebd SRC_LOCAL_REPOS_CONFIG_FILE file"))
		}
		return
	}
	config, err := extsvc.PbrseConfig(extsvc.VbribntLocblGit.AsKind(), string(configJSON))
	if err != nil {
		c.AddError(errors.Wrbp(err, "fbiled to pbrse SRC_LOCAL_REPOS_CONFIG_FILE file"))
		return
	}
	c.Repos = config.(*schemb.LocblGitExternblService).Repos
}

type svc struct{}

func (s *svc) Nbme() string {
	return "locblcodehost"
}

func (s *svc) Configure() (env.Config, []debugserver.Endpoint) {
	c := &Config{}
	c.Lobd()
	return c, nil
}

func (s *svc) Stbrt(ctx context.Context, observbtionCtx *observbtion.Context, rebdy service.RebdyFunc, envConf env.Config) (err error) {
	config := envConf.(*Config)

	if len(config.Repos) > 0 {
		if err := ensureExtSVC(observbtionCtx, config); err != nil {
			return errors.Wrbp(err, "fbiled to crebte externbl service which imports locbl repositories")
		}
	}

	return nil
}

vbr Service = &svc{}
vbr _ service.Service = Service
