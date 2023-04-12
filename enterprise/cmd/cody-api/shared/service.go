package shared

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/service"
)

type svc struct{}

func (svc) Name() string { return "cody-api" }

func (svc) Configure() (env.Config, []debugserver.Endpoint) {
	return nil, []debugserver.Endpoint{debugserver.NewGRPCWebUIEndpoint("cody-api", addr)}
}

func (svc) Start(ctx context.Context, observationCtx *observation.Context, ready service.ReadyFunc, config env.Config) error {
	// TODO - setup connections to db, codeintel apis, etc

	ready()
	goroutine.MonitorBackgroundRoutines(ctx, makeServer())
	return nil
}

var Service service.Service = svc{}
