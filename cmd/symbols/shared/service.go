package shared

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/service"
)

type svc struct{}

func (svc) Name() string { return "symbols" }

func (svc) Configure() (env.Config, []debugserver.Endpoint) {
	LoadConfig()
	return nil, []debugserver.Endpoint{GRPCWebUIDebugEndpoint()}
}

func (svc) Start(ctx context.Context, observationCtx *observation.Context, ready service.ReadyFunc, config env.Config) error {
	return Main(ctx, observationCtx, ready, SetupSqlite)
}

var Service service.Service = svc{}
