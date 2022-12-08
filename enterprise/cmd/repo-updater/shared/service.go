package shared

import (
	"context"

	shared "github.com/sourcegraph/sourcegraph/cmd/repo-updater/shared"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/service"
)

type svc struct{}

func (svc) Name() string { return "repo-updater" }

func (svc) Configure() env.Config {
	repoupdater.LoadConfig()
	return nil
}

func (svc) Start(ctx context.Context, observationCtx *observation.Context, config env.Config) error {
	return shared.Main(ctx, observationCtx, EnterpriseInit)
}

var Service service.Service = svc{}
