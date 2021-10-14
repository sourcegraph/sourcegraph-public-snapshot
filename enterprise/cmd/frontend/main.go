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
	_ "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/registry"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/searchcontexts"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
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

type EnterpriseInitializer = func(context.Context, dbutil.DB, *oobmigration.Runner, *enterprise.Services, *observation.Context) error

var initFunctions = map[string]EnterpriseInitializer{
	"authz":          authz.Init,
	"licensing":      licensing.Init,
	"executor":       executor.Init,
	"codeintel":      codeintel.Init,
	"insights":       insights.Init,
	"batches":        batches.Init,
	"codemonitors":   codemonitors.Init,
	"dotcom":         dotcom.Init,
	"searchcontexts": searchcontexts.Init,
}

func enterpriseSetupHook(db dbutil.DB, outOfBandMigrationRunner *oobmigration.Runner) enterprise.Services {
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

	for name, fn := range initFunctions {
		if err := fn(ctx, db, outOfBandMigrationRunner, &enterpriseServices, observationContext); err != nil {
			log.Fatal(fmt.Sprintf("failed to initialize %s: %s", name, err))
		}
	}

	return enterpriseServices
}
