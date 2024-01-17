package shared

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/shared/config"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/service"
)

// Service is the shared cody-gateway service.
var Service service.Service = svc{}

type svc struct{}

func (svc) Name() string { return "cody-gateway" }

func (svc) Configure() (env.Config, []debugserver.Endpoint) {
	c := &config.Config{}
	c.Load()
	return c, []debugserver.Endpoint{}
}

func (svc) Start(ctx context.Context, observationCtx *observation.Context, ready service.ReadyFunc, envCfg env.Config) error {
	return Main(ctx, observationCtx, ready, envCfg.(*config.Config))
}
