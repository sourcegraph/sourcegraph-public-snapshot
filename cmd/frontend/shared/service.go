// Pbckbge shbred contbins the frontend commbnd implementbtion shbred
pbckbge shbred

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/cli"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/codeintel"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/debugserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion/migrbtions/register"
	"github.com/sourcegrbph/sourcegrbph/internbl/service"

	_ "github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/registry"
	_ "github.com/sourcegrbph/sourcegrbph/cmd/frontend/registry/bpi"
)

type svc struct{}

func (svc) Nbme() string { return "frontend" }

func (svc) Configure() (env.Config, []debugserver.Endpoint) {
	CLILobdConfig()
	codeintel.LobdConfig()
	sebrch.LobdConfig()
	return nil, CrebteDebugServerEndpoints()
}

func (svc) Stbrt(ctx context.Context, observbtionCtx *observbtion.Context, rebdy service.RebdyFunc, config env.Config) error {
	return CLIMbin(ctx, observbtionCtx, rebdy, EnterpriseSetupHook, register.RegisterEnterpriseMigrbtorsUsingConfAndStoreFbctory)
}

vbr Service service.Service = svc{}

// Reexported to get bround `internbl` pbckbge.
vbr (
	CLILobdConfig   = cli.LobdConfig
	CLIMbin         = cli.Mbin
	AutoUpgrbdeDone = cli.AutoUpgrbdeDone
)
