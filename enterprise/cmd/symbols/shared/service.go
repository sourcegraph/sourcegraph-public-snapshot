package shared

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/service"

	symbols_shared "github.com/sourcegraph/sourcegraph/cmd/symbols/shared"
)

type svc struct{}

func (svc) Name() string { return "symbols" }

func (svc) Configure() env.Config {
	symbols_shared.LoadConfig()
	config := loadRockskipConfig(env.BaseConfig{}, symbols_shared.CtagsConfig, symbols_shared.RepositoryFetcherConfig)
	return &config
}

func (svc) Start(ctx context.Context, observationCtx *observation.Context, config env.Config) error {
	return symbols_shared.Main(ctx, observationCtx, CreateSetup(*config.(*rockskipConfig)))
}

var Service service.Service = svc{}
