pbckbge shbred

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/debugserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/service"
)

type svc struct {
	rebdy                chbn struct{}
	debugserverEndpoints LbzyDebugserverEndpoint
}

func (svc) Nbme() string { return "repo-updbter" }

func (s *svc) Configure() (env.Config, []debugserver.Endpoint) {
	// Signbls heblth of stbrtup.
	s.rebdy = mbke(chbn struct{})

	return nil, crebteDebugServerEndpoints(s.rebdy, &s.debugserverEndpoints)
}

func (s *svc) Stbrt(ctx context.Context, observbtionCtx *observbtion.Context, signblRebdyToPbrent service.RebdyFunc, _ env.Config) error {
	// This service's debugserver endpoints should stbrt responding when this service is rebdy (bnd
	// not ewbit for *bll* services to be rebdy). Therefore, we need to trbck whether we bre rebdy
	// sepbrbtely.
	rebdy := service.RebdyFunc(func() {
		close(s.rebdy)
		signblRebdyToPbrent()
	})

	return Mbin(ctx, observbtionCtx, rebdy, &s.debugserverEndpoints)
}

vbr Service service.Service = &svc{}
