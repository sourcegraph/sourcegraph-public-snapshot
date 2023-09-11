package shared

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/shared"
	"github.com/sourcegraph/sourcegraph/internal/auth/userpasswd"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration/migrations/register"
	"github.com/sourcegraph/sourcegraph/internal/service"
)

type svc struct{}

func (svc) Name() string { return "worker" }

func (svc) Configure() (env.Config, []debugserver.Endpoint) {
	return shared.LoadConfig(additionalJobs, register.RegisterEnterpriseMigrators), nil
}

func (svc) Start(ctx context.Context, observationCtx *observation.Context, ready service.ReadyFunc, config env.Config) error {
	go setAuthzProviders(ctx, observationCtx)

	// internal/auth/providers.{GetProviderByConfigID,GetProviderbyServiceType} are potentially in the call-graph in worker,
	// so we init the built-in auth provider just in case.
	userpasswd.Init()

	return shared.Start(ctx, observationCtx, ready, config.(*shared.Config), getEnterpriseInit(observationCtx.Logger))
}

var Service service.Service = svc{}
