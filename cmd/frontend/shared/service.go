// Package shared contains the frontend command implementation shared
package shared

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/cli"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration/migrations/register"
	"github.com/sourcegraph/sourcegraph/internal/service"

	_ "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/registry"
	_ "github.com/sourcegraph/sourcegraph/cmd/frontend/registry/api"
)

type svc struct{}

func (svc) Name() string { return "frontend" }

func (svc) Configure() (env.Config, []debugserver.Endpoint) {
	CLILoadConfig()
	codeintel.LoadConfig()
	search.LoadConfig()
	return nil, CreateDebugServerEndpoints()
}

func (svc) Start(ctx context.Context, observationCtx *observation.Context, ready service.ReadyFunc, config env.Config) error {
	return CLIMain(ctx, observationCtx, ready, EnterpriseSetupHook, register.RegisterEnterpriseMigratorsUsingConfAndStoreFactory)
}

var Service service.Service = svc{}

// Reexported to get around `internal` package.
var (
	CLILoadConfig   = cli.LoadConfig
	CLIMain         = cli.Main
	AutoUpgradeDone = cli.AutoUpgradeDone
)
