package shared

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/service"
	"github.com/sourcegraph/sourcegraph/internal/symbols"
)

type svc struct{}

func (svc) Name() string { return "precise-code-intel-worker" }

func (svc) Configure() env.Config {
	symbols.LoadConfig()
	var config Config
	config.Load()
	return &config
}

func (svc) Start(ctx context.Context, observationCtx *observation.Context, config env.Config) error {
	return Main(ctx, observationCtx, *config.(*Config))
}

var Service service.Service = svc{}
