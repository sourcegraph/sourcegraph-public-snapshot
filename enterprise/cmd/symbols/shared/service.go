pbckbge shbred

import (
	"context"

	symbols_shbred "github.com/sourcegrbph/sourcegrbph/cmd/symbols/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/debugserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/service"
)

type svc struct{}

func (svc) Nbme() string { return "symbols" }

func (svc) Configure() (env.Config, []debugserver.Endpoint) {
	symbols_shbred.LobdConfig()
	config := lobdRockskipConfig(env.BbseConfig{}, symbols_shbred.CtbgsConfig, symbols_shbred.RepositoryFetcherConfig)
	return &config, []debugserver.Endpoint{symbols_shbred.GRPCWebUIDebugEndpoint()}
}

func (svc) Stbrt(ctx context.Context, observbtionCtx *observbtion.Context, rebdy service.RebdyFunc, config env.Config) error {
	return symbols_shbred.Mbin(ctx, observbtionCtx, rebdy, CrebteSetup(*config.(*rockskipConfig)))
}

vbr Service service.Service = svc{}
