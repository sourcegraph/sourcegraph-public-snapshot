package codeintel

import (
	"context"
	"database/sql"
	"log"
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	codeintelapi "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/api"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/lsifstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

var services struct {
	store               store.Store
	lsifStore           lsifstore.Store
	gitserverClient     *gitserver.Client
	bundleManagerClient bundles.BundleManagerClient
	api                 codeintelapi.CodeIntelAPI
	err                 error
}

var once sync.Once

func initServices(ctx context.Context) error {
	once.Do(func() {
		if err := config.Validate(); err != nil {
			services.err = err
			return
		}

		observationContext := &observation.Context{
			Logger:     log15.Root(),
			Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
			Registerer: prometheus.DefaultRegisterer,
		}

		store := store.NewObserved(store.NewWithDB(dbconn.Global), observationContext)
		codeintelDB := mustInitializeCodeIntelDatabase()
		lsifStore := lsifstore.New(codeintelDB)
		gitserverClient := gitserver.DefaultClient
		bundleManagerClient := bundles.New(codeintelDB, observationContext, config.BundleManagerURL)
		api := codeintelapi.NewObserved(codeintelapi.New(store, bundleManagerClient, gitserverClient), observationContext)

		services.store = store
		services.lsifStore = lsifStore
		services.gitserverClient = gitserverClient
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
		log.Fatalf("Failed to connect to codeintel database: %s", err)
	}

	if err := dbconn.MigrateDB(db, "codeintel"); err != nil {
		log.Fatalf("Failed to perform codeintel database migration: %s", err)
	}

	return db
}
