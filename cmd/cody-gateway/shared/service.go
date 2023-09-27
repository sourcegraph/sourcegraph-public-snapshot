pbckbge shbred

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/debugserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/service"
)

// Service is the shbred cody-gbtewby service.
vbr Service service.Service = svc{}

type svc struct{}

func (svc) Nbme() string { return "cody-gbtewby" }

func (svc) Configure() (env.Config, []debugserver.Endpoint) {
	c := &Config{}
	c.Lobd()
	return c, []debugserver.Endpoint{}
}

func (svc) Stbrt(ctx context.Context, observbtionCtx *observbtion.Context, rebdy service.RebdyFunc, config env.Config) error {
	return Mbin(ctx, observbtionCtx, rebdy, config.(*Config))
}
