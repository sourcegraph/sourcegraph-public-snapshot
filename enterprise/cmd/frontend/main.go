// Command frontend contains the enterprise frontend implementation.
//
// It wraps the open source frontend command and merely injects a few
// proprietary things into it via e.g. blank/underscore imports in this file
// which register side effects with the frontend package.
package main

import (
	"context"
	"os"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/app"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codemonitors"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/compute"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/dotcom"
	executor "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue"
	licensing "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/licensing/init"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/notebooks"
	_ "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/registry"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/searchcontexts"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func main() {
	shared.Main(enterpriseSetupHook)
}

func init() {
	oobmigration.ReturnEnterpriseMigrations = true
}

type EnterpriseInitializer = func(context.Context, database.DB, conftypes.UnifiedWatchable, *enterprise.Services, *observation.Context) error

var initFunctions = map[string]EnterpriseInitializer{
	"authz":          authz.Init,
	"licensing":      licensing.Init,
	"insights":       insights.Init,
	"batches":        batches.Init,
	"codemonitors":   codemonitors.Init,
	"dotcom":         dotcom.Init,
	"searchcontexts": searchcontexts.Init,
	"notebooks":      notebooks.Init,
	"compute":        compute.Init,
}

var codeIntelConfig = &codeintel.Config{}

func init() {
	codeIntelConfig.Load()
}

func enterpriseSetupHook(db database.DB, conf conftypes.UnifiedWatchable) enterprise.Services {
	logger := log.Scoped("enterprise", "frontend enterprise edition")
	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	if debug {
		logger.Debug("enterprise edition")
	}

	auth.Init(logger, db)

	ctx := context.Background()
	enterpriseServices := enterprise.DefaultServices()

	observationContext := &observation.Context{
		Logger:     logger,
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}

	if err := codeIntelConfig.Validate(); err != nil {
		logger.Fatal("failed to load codeintel config", log.Error(err))
	}

	services, err := codeintel.NewServices(ctx, codeIntelConfig, conf, db)
	if err != nil {
		logger.Fatal(err.Error())
	}

	if err := codeintel.Init(ctx, db, codeIntelConfig, &enterpriseServices, services); err != nil {
		logger.Fatal("failed to initialize codeintel", log.Error(err))
	}

	// Initialize executor-specific services with the code-intel services.
	if err := executor.Init(ctx, db, conf, &enterpriseServices, observationContext, services.InternalUploadHandler); err != nil {
		logger.Fatal("failed to initialize executor", log.Error(err))
	}

	if err := app.Init(db, conf, &enterpriseServices); err != nil {
		logger.Fatal("failed to initialize app", log.Error(err))
	}

	// Initialize all the enterprise-specific services that do not need the codeintel-specific services.
	for name, fn := range initFunctions {
		initLogger := logger.Scoped(name, "")
		if err := fn(ctx, db, conf, &enterpriseServices, observationContext); err != nil {
			initLogger.Fatal("failed to initialize", log.Error(err))
		}
	}

	return enterpriseServices
}
