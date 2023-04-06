package weaviate

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/service"
)

type svc struct{}

func (s svc) Name() string { return "weaviate" }

func (s svc) Configure() (env.Config, []debugserver.Endpoint) {
	cfg := &Config{}
	cfg.Load()
	return cfg, nil
}

func (s svc) Start(ctx context.Context, observationCtx *observation.Context, ready service.ReadyFunc, c env.Config) error {
	err := start(ctx, observationCtx, c.(*Config))
	if err != nil {
		return err
	}
	ready()
	return nil
}

var Service service.Service = svc{}
