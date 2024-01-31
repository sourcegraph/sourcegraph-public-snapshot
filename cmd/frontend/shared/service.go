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

type svc struct {
	ready                chan struct{}
	debugserverEndpoints cli.LazyDebugserverEndpoint
}

func (svc) Name() string { return "frontend" }

func (s *svc) Configure() (env.Config, []debugserver.Endpoint) {
	CLILoadConfig()
	codeintel.LoadConfig()
	search.LoadConfig()
	// Signals health of startup.
	s.ready = make(chan struct{})

	return nil, createDebugServerEndpoints(s.ready, &s.debugserverEndpoints)
}

func (s *svc) Start(ctx context.Context, observationCtx *observation.Context, signalReadyToParent service.ReadyFunc, config env.Config) error {
	// This service's debugserver endpoints should start responding when this service is ready (and
	// not ewait for *all* services to be ready). Therefore, we need to track whether we are ready
	// separately.
	ready := service.ReadyFunc(func() {
		close(s.ready)
		signalReadyToParent()
	})

	return CLIMain(ctx, observationCtx, ready, &s.debugserverEndpoints, EnterpriseSetupHook, register.RegisterEnterpriseMigratorsUsingConfAndStoreFactory)
}

var Service service.Service = &svc{}

// Reexported to get around `internal` package.
var (
	CLILoadConfig   = cli.LoadConfig
	CLIMain         = cli.Main
	AutoUpgradeDone = cli.AutoUpgradeDone
)
