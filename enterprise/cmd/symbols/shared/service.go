package shared

import (
	"context"

	symbols_shared "github.com/sourcegraph/sourcegraph/cmd/symbols/shared"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/service"
)

type svc struct{}

func (svc) Name() string { return "symbols" }

func (svc) Configure() (env.Config, []debugserver.Endpoint) {
	symbols_shared.LoadConfig()
	config := loadRockskipConfig(env.BaseConfig{}, symbols_shared.CtagsConfig, symbols_shared.RepositoryFetcherConfig)
	return &config, []debugserver.Endpoint{symbols_shared.GRPCWebUIDebugEndpoint()}
}

func (svc) Start(ctx context.Context, observationCtx *observation.Context, ready service.ReadyFunc, config env.Config) error {
	return symbols_shared.Main(ctx, observationCtx, ready, CreateSetup(*config.(*rockskipConfig)))
}

var Service service.Service = svc{}
