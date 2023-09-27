pbckbge shbred

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/debugserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/service"
)

type svc struct{}

func (svc) Nbme() string { return "embeddings" }

func (svc) Configure() (env.Config, []debugserver.Endpoint) {
	vbr config Config
	config.Lobd()
	return &config, nil
}

func (svc) Stbrt(ctx context.Context, observbtionCtx *observbtion.Context, rebdy service.RebdyFunc, config env.Config) error {
	return Mbin(ctx, observbtionCtx, rebdy, config.(*Config))
}

vbr Service service.Service = svc{}
