package main

import (
	"context"
	"os"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	batchesstore "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	SourcegraphEndpoint = env.Get("SOURCEGRAPH_BASE_URL", "http://127.0.0.1:7080", "Sourcegraph frontend endpoint")
	githubToken         = env.Get("GITHUB_TOKEN", "", "GITHUB_TOKEN to clone the repositories")
)

func main() {
	ctx := context.Background()
	logfuncs := log.Init(log.Resource{
		Name: "executors test runner",
	})
	defer logfuncs.Sync()

	logger := log.Scoped("init", "runner initialization process")

	db, err := initDB(logger)
	if err != nil {
		logger.Fatal("failed to connect to DB", log.Error(err))
	}

	bstore := batchesstore.New(db, &observation.TestContext, nil)

	// Verify the database connection works.
	count, err := bstore.CountBatchChanges(ctx, batchesstore.CountBatchChangesOpts{})
	if err != nil {
		logger.Fatal("failed to count batch changes", log.Error(err))
	}

	if count != 0 {
		logger.Fatal("instance has preexisting batch changes")
	}

	logger.Info("Instance is clean")

	var client *gqltestutil.Client
	client, err = initAndAuthenticate()
	if err != nil {
		logger.Fatal("Failed to set up user", log.Error(err))
	}

	// Make sure repos are cloned in the instance.
	if err := ensureRepos(client); err != nil {
		logger.Fatal("Ensuring repos exist in the instance", log.Error(err))
	}

	// Now that we have our repositories synced and cloned into the instance, we
	// can start testing.

	tests := []TestFunc{
		RunTestBasicExample,
		RunTest2ConditionalStaticallySkippedSteps,
	}
	anyFailed := false
	for _, tf := range tests {
		if pass := tf(ctx, client, bstore, logger); !pass {
			anyFailed = true
		}
		cleanupDB(ctx, db, logger)
	}

	if anyFailed {
		os.Exit(1)
	}
}

type TestFunc = func(ctx context.Context, client *gqltestutil.Client, bstore *store.Store, logger log.Logger) bool

func initDB(logger log.Logger) (database.DB, error) {
	// This call to SetProviders is here so that calls to GetProviders don't block.
	authz.SetProviders(true, []authz.Provider{})

	obsCtx := observation.TestContext
	obsCtx.Logger = logger
	sqlDB, err := connections.RawNewFrontendDB(&obsCtx, "postgres://sg@127.0.0.1:5433/sg", "")
	if err != nil {
		return nil, errors.Errorf("failed to connect to database: %s", err)
	}

	logger.Info("Connected to database!")

	return database.NewDB(logger, sqlDB), nil
}

func cleanupDB(ctx context.Context, db database.DB, logger log.Logger) {
	q := sqlf.Sprintf(cleanupDBQueryFmtstr)
	if _, err := db.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...); err != nil {
		logger.Fatal("failed to cleanup DB state", log.Error(err))
	}
}

const cleanupDBQueryFmtstr = `
TRUNCATE batch_changes RESTART IDENTITY CASCADE;
TRUNCATE batch_spec_execution_cache_entries RESTART IDENTITY CASCADE;
TRUNCATE batch_spec_resolution_jobs RESTART IDENTITY CASCADE;
TRUNCATE batch_spec_workspace_execution_jobs RESTART IDENTITY CASCADE;
TRUNCATE batch_spec_workspace_execution_last_dequeues RESTART IDENTITY CASCADE;
TRUNCATE batch_spec_workspace_files RESTART IDENTITY CASCADE;
TRUNCATE batch_spec_workspaces RESTART IDENTITY CASCADE;
TRUNCATE batch_specs RESTART IDENTITY CASCADE;
TRUNCATE changeset_specs RESTART IDENTITY CASCADE;
TRUNCATE changesets RESTART IDENTITY CASCADE;
`
