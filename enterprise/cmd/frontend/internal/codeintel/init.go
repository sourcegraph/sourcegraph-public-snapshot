package codeintel

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	codeintelapi "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/api"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	codeintelhttpapi "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/httpapi"
	codeintelresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/resolvers"
	codeintelgqlresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/resolvers/graphql"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

var bundleManagerURL = env.Get("PRECISE_CODE_INTEL_BUNDLE_MANAGER_URL", "", "HTTP address for internal LSIF bundle manager server.")
var rawHunkCacheSize = env.Get("PRECISE_CODE_INTEL_HUNK_CACHE_CAPACITY", "1000", "Maximum number of git diff hunk objects that can be loaded into the hunk cache at once.")

func Init(ctx context.Context, enterpriseServices *enterprise.Services) error {
	if err := initOnce(ctx); err != nil {
		return err
	}

	hunkCacheSize, err := strconv.ParseInt(rawHunkCacheSize, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid int %q for PRECISE_CODE_INTEL_HUNK_CACHE_CAPACITY: %s", rawHunkCacheSize, err)
	}

	hunkCache, err := codeintelresolvers.NewHunkCache(int(hunkCacheSize))
	if err != nil {
		return fmt.Errorf("failed to initialize hunk cache: %s", err)
	}

	resolver := codeintelgqlresolvers.NewResolver(codeintelresolvers.NewResolver(
		services.store,
		services.bundleManagerClient,
		services.api,
		hunkCache,
	))

	internalHandler, err := NewCodeIntelUploadHandler(ctx, true)
	if err != nil {
		return err
	}

	externalHandler, err := NewCodeIntelUploadHandler(ctx, false)
	if err != nil {
		return err
	}

	enterpriseServices.CodeIntelResolver = resolver
	enterpriseServices.NewCodeIntelUploadHandler = func(internal bool) http.Handler {
		if internal {
			return internalHandler
		}

		return externalHandler
	}

	return nil
}

func NewCodeIntelUploadHandler(ctx context.Context, internal bool) (http.Handler, error) {
	if err := initOnce(ctx); err != nil {
		return nil, err
	}

	return codeintelhttpapi.NewUploadHandler(services.store, services.bundleManagerClient, internal), nil
}

var once sync.Once
var services struct {
	store               store.Store
	bundleManagerClient bundles.BundleManagerClient
	api                 codeintelapi.CodeIntelAPI
	err                 error
}

func initOnce(ctx context.Context) error {
	once.Do(func() {
		if bundleManagerURL == "" {
			services.err = fmt.Errorf("invalid value for PRECISE_CODE_INTEL_BUNDLE_MANAGER_URL: no value supplied")
			return
		}

		observationContext := &observation.Context{
			Logger:     log15.Root(),
			Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
			Registerer: prometheus.DefaultRegisterer,
		}

		codeIntelDB := mustInitializeCodeIntelDatabase()
		store := store.NewObserved(store.NewWithDB(dbconn.Global), observationContext)
		bundleManagerClient := bundles.New(codeIntelDB, observationContext, bundleManagerURL)
		api := codeintelapi.NewObserved(codeintelapi.New(store, bundleManagerClient, gitserver.DefaultClient), observationContext)

		services.store = store
		services.bundleManagerClient = bundleManagerClient
		services.api = api
	})

	return services.err
}

func mustInitializeCodeIntelDatabase() *sql.DB {
	postgresDSN := conf.Get().ServiceConnections.CodeIntelPostgresDSN
	conf.Watch(func() {
		if newDSN := conf.Get().ServiceConnections.CodeIntelPostgresDSN; postgresDSN != newDSN {
			log.Fatalf("detected database DSN change, restarting to take effect: %s", newDSN)
		}
	})

	db, err := dbconn.New(postgresDSN, "_codeintel")
	if err != nil {
		log.Fatalf("failed to connect to codeintel database: %s", err)
	}

	if err := dbconn.MigrateDB(db, "codeintel"); err != nil {
		log.Fatalf("failed to perform codeintel database migration: %s", err)
	}

	return db
}
