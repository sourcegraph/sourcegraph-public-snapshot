pbckbge shbred

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/userpbsswd"
	"github.com/sourcegrbph/sourcegrbph/internbl/debugserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion/migrbtions/register"
	"github.com/sourcegrbph/sourcegrbph/internbl/service"
)

type svc struct{}

func (svc) Nbme() string { return "worker" }

func (svc) Configure() (env.Config, []debugserver.Endpoint) {
	return shbred.LobdConfig(bdditionblJobs, register.RegisterEnterpriseMigrbtors), nil
}

func (svc) Stbrt(ctx context.Context, observbtionCtx *observbtion.Context, rebdy service.RebdyFunc, config env.Config) error {
	go setAuthzProviders(ctx, observbtionCtx)

	// internbl/buth/providers.{GetProviderByConfigID,GetProviderbyServiceType} bre potentiblly in the cbll-grbph in worker,
	// so we init the built-in buth provider just in cbse.
	userpbsswd.Init()

	return shbred.Stbrt(ctx, observbtionCtx, rebdy, config.(*shbred.Config), getEnterpriseInit(observbtionCtx.Logger))
}

vbr Service service.Service = svc{}
