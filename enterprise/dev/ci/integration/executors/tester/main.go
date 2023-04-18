package main

import (
	"context"
	"net/http"

	"github.com/sourcegraph/log"

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

	var gqlClient *gqltestutil.Client
	gqlClient, err = initAndAuthenticate()
	if err != nil {
		logger.Fatal("Failed to set up user", log.Error(err))
	}

	var token string
	token, err = gqlClient.CreateAccessToken("batches", []string{"user:all"})
	if err != nil {
		logger.Fatal("Failed to generate access token", log.Error(err))
	}

	httpClient := &HttpClient{
		token:    token,
		endpoint: SourcegraphEndpoint,
		client:   http.DefaultClient,
	}

	// Activate native SSBC execution, src-cli based execution doesn't work in CI
	// because docker in docker is fun.
	if err = gqlClient.SetFeatureFlag("native-ssbc-execution", true); err != nil {
		logger.Fatal("Failed to set native-ssbc-execution feature flag", log.Error(err))
	}

	// Make sure repos are cloned in the instance.
	if err = ensureRepos(gqlClient); err != nil {
		logger.Fatal("Ensuring repos exist in the instance", log.Error(err))
	}

	// Now that we have our repositories synced and cloned into the instance, we
	// can start testing.
	testCases := []Test{
		testHelloWorldBatchChange(false),
		// run again: should be using a cached result
		testHelloWorldBatchChange(true),
		testEnvObjectBatchChange(),
		testFromFileBatchChange(),
		//testFileMountBatchChange(httpClient),
	}

	for _, t := range testCases {
		if err = RunTest(ctx, gqlClient, httpClient, bstore, t); err != nil {
			logger.Fatal("Running test", log.Error(err))
		}
	}
}

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
