package shared

import (
	"context"

	frontend_shared "github.com/sourcegraph/sourcegraph/cmd/frontend/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/service"

	_ "github.com/sourcegraph/sourcegraph/cmd/frontend/registry/api"
	_ "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/registry"
)

type svc struct{}

func (svc) Name() string { return "frontend" }

func (svc) Configure() env.Config {
	frontend_shared.CLIConfigureTODO()
	codeintel.LoadConfig()
	return nil
}

func (svc) Start(ctx context.Context, observationCtx *observation.Context, config env.Config) error {
	return frontend_shared.CLIMainTODO(ctx, observationCtx, EnterpriseSetupHook)
}

var Service service.Service = svc{}
