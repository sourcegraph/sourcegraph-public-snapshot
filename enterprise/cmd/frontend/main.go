// Command frontend contains the enterprise frontend implementation.
//
// It wraps the open source frontend command and merely injects a few
// proprietary things into it via e.g. blank/underscore imports in this file
// which register side effects with the frontend package.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/oobmigration"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codemonitors"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/dotcom"
	executor "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue"
	licensing "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/licensing/init"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/notebooks"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/orgrepos"
	_ "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/registry"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/searchcontexts"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
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
	"enterprise":     orgrepos.Init,
	"notebooks":      notebooks.Init,
}

func enterpriseSetupHook(db database.DB, conf conftypes.UnifiedWatchable) enterprise.Services {
	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	if debug {
		log.Println("enterprise edition")
	}

	auth.Init(db)

	ctx := context.Background()
	enterpriseServices := enterprise.DefaultServices()

	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	services, err := codeintel.NewServices(ctx, conf, db)
	if err != nil {
		log.Fatal(err)
	}

	if err := codeintel.Init(ctx, db, conf, &enterpriseServices, observationContext, services); err != nil {
		log.Fatal(fmt.Sprintf("failed to initialize codeintel: %s", err))
	}

	// Initialize executor-specific services with the code-intel services.
	if err := executor.Init(ctx, db, conf, &enterpriseServices, observationContext, services.InternalUploadHandler); err != nil {
		log.Fatal(fmt.Sprintf("failed to initialize executor: %s", err))
	}

	// Initialize all the enterprise-specific services that do not need the codeintel-specific services.
	for name, fn := range initFunctions {
		if err := fn(ctx, db, conf, &enterpriseServices, observationContext); err != nil {
			log.Fatal(fmt.Sprintf("failed to initialize %s: %s", name, err))
		}
	}

	return enterpriseServices
}
