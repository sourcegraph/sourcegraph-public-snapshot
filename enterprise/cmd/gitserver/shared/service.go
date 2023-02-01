package shared

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/shared"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/service"
)

type svc struct{}

func (svc) Name() string { return "gitserver" }

func (svc) Configure() (env.Config, []debugserver.Endpoint) {
	return shared.LoadConfig(), nil
}

func (svc) Start(ctx context.Context, observationCtx *observation.Context, ready service.ReadyFunc, config env.Config) error {
	return shared.Main(ctx, observationCtx, ready, config.(*shared.Config), enterpriseInit)
}

var Service service.Service = svc{}
