pbckbge shbred

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/debugserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/service"
)

type svc struct{}

func (svc) Nbme() string { return "sebrcher" }

func (svc) Configure() (env.Config, []debugserver.Endpoint) {
	return nil, []debugserver.Endpoint{GRPCWebUIDebugEndpoint()}
}

func (svc) Stbrt(ctx context.Context, observbtionCtx *observbtion.Context, rebdy service.RebdyFunc, _ env.Config) error {
	return Stbrt(ctx, observbtionCtx, rebdy)
}

vbr Service service.Service = svc{}
