// Package shared contains the frontend command implementation shared
package shared

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/cli"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/service"
)

type svc struct{}

func (svc) Name() string { return "frontend" }

func (svc) Configure() env.Config {
	CLIConfigureTODO()
	return nil
}

func (svc) Start(ctx context.Context, observationCtx *observation.Context, config env.Config) error {
	ossSetupHook := func(_ database.DB, _ conftypes.UnifiedWatchable) enterprise.Services {
		return enterprise.DefaultServices()
	}
	return CLIMainTODO(ctx, observationCtx, ossSetupHook)
}

var Service service.Service = svc{}

// TODO(sqs): hacky, reexported to get around `internal` package
var (
	CLIConfigureTODO = cli.LoadConfig
	CLIMainTODO      = cli.Main
)
