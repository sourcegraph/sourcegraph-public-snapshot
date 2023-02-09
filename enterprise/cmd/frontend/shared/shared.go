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
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/app"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	ossauthz "github.com/sourcegraph/sourcegraph/internal/authz"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches"
	codeintelinit "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codemonitors"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/compute"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/dotcom"
	executor "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/insights"
	licensing "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/licensing/init"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/notebooks"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/rbac"
	_ "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/registry"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/repos/webhooks"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/searchcontexts"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
	codeintelshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/scim"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/graph"
	sgtypes "github.com/sourcegraph/sourcegraph/internal/types"
)

type EnterpriseInitializer = func(context.Context, *observation.Context, database.DB, codeintel.Services, conftypes.UnifiedWatchable, *enterprise.Services) error

var initFunctions = map[string]EnterpriseInitializer{
	"app":            app.Init,
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
	"rbac":           rbac.Init,
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
	c, err := codenav.NewHunkCache(100)
	if err != nil {
		logger.Fatal("failed to initialize code intelligence", log.Error(err))
	}
	graph.RegisterStore(&codeNavShim{
		svc:       codeIntelServices,
		gs:        gitserver.New(&observation.TestContext, db),
		hunkCache: c,
	})

	for name, fn := range initFunctions {
		if err := fn(ctx, observationCtx, db, codeIntelServices, conf, &enterpriseServices); err != nil {
			logger.Fatal("failed to initialize", log.String("name", name), log.Error(err))
		}
	}

	// Inititalize executor last, as we require code intel and batch changes services to be
	// already populated on the enterpriseServices object.
	if err := executor.Init(ctx, observationCtx, db, conf, &enterpriseServices); err != nil {
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

type codeNavShim struct {
	svc codeintel.Services
	gs  *gitserver.Client

	hunkCache codenav.HunkCache
}

func toLocations(uls []types.UploadLocation) []sgtypes.CodeIntelLocation {
	ls := make([]sgtypes.CodeIntelLocation, len(uls))
	for i, l := range uls {
		ls[i] = sgtypes.CodeIntelLocation{
			Path:         l.Path,
			TargetCommit: l.TargetCommit,
			TargetRange: sgtypes.CodeIntelRange{
				Start: sgtypes.CodeIntelPosition(l.TargetRange.Start),
				End:   sgtypes.CodeIntelPosition(l.TargetRange.End),
			},
		}
	}
	return ls
}

func (s *codeNavShim) GetReferences(ctx context.Context, repo sgtypes.MinimalRepo, args sgtypes.CodeIntelRequestArgs) (_ []sgtypes.CodeIntelLocation, err error) {
	uploads, err := s.svc.CodenavService.GetClosestDumpsForBlob(ctx, args.RepositoryID, args.Commit, args.Path, true, "")
	if err != nil || len(uploads) == 0 {
		return nil, err
	}

	reqState := codenav.NewRequestState(uploads, ossauthz.DefaultSubRepoPermsChecker, s.gs, repo.ToRepo(), string(args.Commit), args.Path, 10, s.hunkCache)

	locs, _, err := s.svc.CodenavService.GetReferences(ctx, shared.RequestArgs(args), reqState, shared.ReferencesCursor{
		Phase: "local",
	})
	return toLocations(locs), err
}

func (s *codeNavShim) GetImplementations(ctx context.Context, repo sgtypes.MinimalRepo, args sgtypes.CodeIntelRequestArgs) (_ []sgtypes.CodeIntelLocation, err error) {
	uploads, err := s.svc.CodenavService.GetClosestDumpsForBlob(ctx, args.RepositoryID, args.Commit, args.Path, true, "")
	if err != nil || len(uploads) == 0 {
		return nil, err
	}

	reqState := codenav.NewRequestState(uploads, ossauthz.DefaultSubRepoPermsChecker, s.gs, repo.ToRepo(), string(args.Commit), args.Path, 10, s.hunkCache)

	locs, _, err := s.svc.CodenavService.GetImplementations(ctx, shared.RequestArgs(args), reqState, shared.ImplementationsCursor{
		Phase: "local",
	})
	return toLocations(locs), err
}
