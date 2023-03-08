// Package shared contains the frontend command implementation shared
package shared

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/cli"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/service"
)

type svc struct{}

func (svc) Name() string { return "frontend" }

func (svc) Configure() (env.Config, []debugserver.Endpoint) {
	CLILoadConfig()
	return nil, GRPCWebUIDebugEndpoints()
}

func (svc) Start(ctx context.Context, observationCtx *observation.Context, ready service.ReadyFunc, config env.Config) error {
	ossSetupHook := func(_ database.DB, _ conftypes.UnifiedWatchable) enterprise.Services {
		return enterprise.DefaultServices()
	}
	return CLIMain(ctx, observationCtx, ready, ossSetupHook)
}

var Service service.Service = svc{}

// Reexported to get around `internal` package.
var (
	CLILoadConfig = cli.LoadConfig
	CLIMain       = cli.Main
)
