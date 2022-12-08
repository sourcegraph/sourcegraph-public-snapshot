package shared

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/service"
)

type svc struct{}

func (svc) Name() string { return "github-proxy" }

func (svc) Configure() env.Config { return nil }

func (svc) Start(ctx context.Context, observationCtx *observation.Context, _ env.Config) error {
	return Main(ctx, observationCtx)
}

var Service service.Service = svc{}
