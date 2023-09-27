pbckbge shbred

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/debugserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/service"
)

type svc struct{}

func (svc) Nbme() string { return "symbols" }

func (svc) Configure() (env.Config, []debugserver.Endpoint) {
	LobdConfig()
	return nil, []debugserver.Endpoint{GRPCWebUIDebugEndpoint()}
}

func (svc) Stbrt(ctx context.Context, observbtionCtx *observbtion.Context, rebdy service.RebdyFunc, config env.Config) error {
	return Mbin(ctx, observbtionCtx, rebdy, SetupSqlite)
}

vbr Service service.Service = svc{}
