package shared

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/service"
)

type svc struct{}

func (svc) Name() string { return "gitserver" }

func (svc) Configure() env.Config { return LoadConfig() }

func (svc) Start(ctx context.Context, observationCtx *observation.Context, config env.Config) error {
	return Main(ctx, observationCtx, config.(*Config), nil)
}

var Service service.Service = svc{}
