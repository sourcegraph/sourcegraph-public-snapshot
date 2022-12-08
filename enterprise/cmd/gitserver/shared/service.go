package shared

import (
	"context"

	shared "github.com/sourcegraph/sourcegraph/cmd/gitserver/shared"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/service"
)

type svc struct{}

func (svc) Name() string { return "gitserver" }

func (svc) Configure() env.Config {
	return shared.LoadConfig()
}

func (svc) Start(ctx context.Context, observationCtx *observation.Context, config env.Config) error {
	return shared.Main(ctx, observationCtx, config.(*shared.Config), enterpriseInit)
}

var Service service.Service = svc{}
