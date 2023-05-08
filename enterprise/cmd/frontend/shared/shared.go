// Package shared is the enterprise frontend program's shared main entrypoint.
//
// It lets the invoker of the OSS frontend shared entrypoint injects a few
// proprietary things into it via e.g. blank/underscore imports in this file
// which register side effects with the frontend package.
package shared

import (
	"context"
	"os"
	"strconv"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth"
	githubapp "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/githubappauth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches"
	codeintelinit "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codemonitors"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/completions"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/compute"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/embeddings"
	executor "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/insights"
	licensing "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/licensing/init"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/notebooks"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/own"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/rbac"
	_ "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/registry"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/repos/webhooks"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/search"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/searchcontexts"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel"
	codeintelshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/scim"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type EnterpriseInitializer = func(context.Context, *observation.Context, database.DB, codeintel.Services, conftypes.UnifiedWatchable, *enterprise.Services) error

var initFunctions = map[string]EnterpriseInitializer{
	"authz":          authz.Init,
	"batches":        batches.Init,
	"codeintel":      codeintelinit.Init,
	"codemonitors":   codemonitors.Init,
	"compute":        compute.Init,
	"dotcom":         dotcom.Init,
	"insights":       insights.Init,
	"licensing":      licensing.Init,
	"notebooks":      notebooks.Init,
	"scim":           scim.Init,
	"searchcontexts": searchcontexts.Init,
	"repos.webhooks": webhooks.Init,
	"embeddings":     embeddings.Init,
	"rbac":           rbac.Init,
	"own":            own.Init,
	"completions":    completions.Init,
	"githubapp":      githubapp.Init,
}

func EnterpriseSetupHook(db database.DB, conf conftypes.UnifiedWatchable) enterprise.Services {
	logger := log.Scoped("enterprise", "frontend enterprise edition")
	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	if debug {
		logger.Debug("enterprise edition")
	}

	auth.Init(logger, db)

	ctx := context.Background()
	enterpriseServices := enterprise.DefaultServices()

	observationCtx := observation.NewContext(logger)

	codeIntelServices, err := codeintel.NewServices(codeintel.ServiceDependencies{
		DB:             db,
		CodeIntelDB:    mustInitializeCodeIntelDB(logger),
		ObservationCtx: observationCtx,
	})
	if err != nil {
		logger.Fatal("failed to initialize code intelligence", log.Error(err))
	}

	// Initialize search first, as we require enterprise search jobs to exist already
	// when other initializers are called.
	if err := search.Init(ctx, observationCtx, db, codeIntelServices, conf, &enterpriseServices); err != nil {
		logger.Fatal("failed to initialize search", log.Error(err))
	}

	for name, fn := range initFunctions {
		if err := fn(ctx, observationCtx, db, codeIntelServices, conf, &enterpriseServices); err != nil {
			logger.Fatal("failed to initialize", log.String("name", name), log.Error(err))
		}
	}

	// Inititalize executor last, as we require code intel and batch changes services to be
	// already populated on the enterpriseServices object.
	if err := executor.Init(observationCtx, db, conf, &enterpriseServices); err != nil {
		logger.Fatal("failed to initialize executor", log.Error(err))
	}

	return enterpriseServices
}

func mustInitializeCodeIntelDB(logger log.Logger) codeintelshared.CodeIntelDB {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.CodeIntelPostgresDSN
	})

	db, err := connections.EnsureNewCodeIntelDB(observation.NewContext(logger), dsn, "frontend")
	if err != nil {
		logger.Fatal("Failed to connect to codeintel database", log.Error(err))
	}

	return codeintelshared.NewCodeIntelDB(logger, db)
}
