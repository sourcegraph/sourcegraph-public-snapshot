package shared

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/service"
)

type svc struct{}

func (svc) Name() string { return "symbols" }

func (svc) Configure() env.Config {
	LoadConfig()
	return nil
}

func (svc) Start(ctx context.Context, observationCtx *observation.Context, config env.Config) error {
	return Main(ctx, observationCtx, SetupSqlite)
}

var Service service.Service = svc{}
