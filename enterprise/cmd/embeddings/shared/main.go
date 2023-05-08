package shared

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	eiauthz "github.com/sourcegraph/sourcegraph/enterprise/internal/authz"
	srp "github.com/sourcegraph/sourcegraph/enterprise/internal/authz/subrepoperms"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/instrumentation"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/service"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

const addr = ":9991"

func Main(ctx context.Context, observationCtx *observation.Context, ready service.ReadyFunc, config *Config) error {
	logger := observationCtx.Logger

	// Initialize tracing/metrics
	observationCtx = observation.NewContext(logger, observation.Honeycomb(&honey.Dataset{
		Name:       "embeddings",
		SampleRate: 20,
	}))

	// Initialize main DB connection.
	sqlDB := mustInitializeFrontendDB(observationCtx)
	db := database.NewDB(logger, sqlDB)

	go setAuthzProviders(ctx, db)

	repoStore := db.Repos()
	repoEmbeddingJobsStore := repo.NewRepoEmbeddingJobsStore(db)

	// Run setup
	gitserverClient := gitserver.NewClient()
	uploadStore, err := embeddings.NewEmbeddingsUploadStore(ctx, observationCtx, config.EmbeddingsUploadStoreConfig)
	if err != nil {
		return err
	}

	authz.DefaultSubRepoPermsChecker, err = srp.NewSubRepoPermsClient(edb.NewEnterpriseDB(db).SubRepoPerms())
	if err != nil {
		return errors.Wrap(err, "creating sub-repo client")
	}

	readFile := func(ctx context.Context, repoName api.RepoName, revision api.CommitID, fileName string) ([]byte, error) {
		return gitserverClient.ReadFile(ctx, authz.DefaultSubRepoPermsChecker, repoName, revision, fileName)
	}

	indexGetter, err := NewCachedEmbeddingIndexGetter(
		repoStore,
		repoEmbeddingJobsStore,
		func(ctx context.Context, repoEmbeddingIndexName embeddings.RepoEmbeddingIndexName) (*embeddings.RepoEmbeddingIndex, error) {
			return embeddings.DownloadRepoEmbeddingIndex(ctx, uploadStore, string(repoEmbeddingIndexName))
		},
		config.EmbeddingsCacheSize,
	)
	if err != nil {
		return err
	}

	getQueryEmbedding, err := getCachedQueryEmbeddingFn()
	if err != nil {
		return err
	}

	weaviate := newWeaviateClient(
		logger,
		readFile,
		getQueryEmbedding,
		config.WeaviateURL,
	)

	getContextDetectionEmbeddingIndex := getCachedContextDetectionEmbeddingIndex(uploadStore)

	// Create HTTP server
	handler := NewHandler(logger, readFile, indexGetter.Get, getQueryEmbedding, weaviate, getContextDetectionEmbeddingIndex)
	handler = handlePanic(logger, handler)
	handler = featureflag.Middleware(db.FeatureFlags(), handler)
	handler = trace.HTTPMiddleware(logger, handler, conf.DefaultClient())
	handler = instrumentation.HTTPMiddleware("", handler)
	handler = actor.HTTPMiddleware(logger, handler)
	server := httpserver.NewFromAddr(addr, &http.Server{
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Handler:      handler,
	})

	// Mark health server as ready and go!
	ready()

	goroutine.MonitorBackgroundRoutines(ctx, server)

	return nil
}

func NewHandler(
	logger log.Logger,
	readFile readFileFn,
	getRepoEmbeddingIndex getRepoEmbeddingIndexFn,
	getQueryEmbedding getQueryEmbeddingFn,
	weaviate *weaviateClient,
	getContextDetectionEmbeddingIndex getContextDetectionEmbeddingIndexFn,
) http.Handler {
	// Initialize the legacy JSON API server
	mux := http.NewServeMux()
	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, fmt.Sprintf("unsupported method %s", r.Method), http.StatusBadRequest)
			return
		}

		var args embeddings.EmbeddingsSearchParameters
		err := json.NewDecoder(r.Body).Decode(&args)
		if err != nil {
			http.Error(w, "could not parse request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		res, err := searchRepoEmbeddingIndex(r.Context(), logger, args, readFile, getRepoEmbeddingIndex, getQueryEmbedding, weaviate)
		if errcode.IsNotFound(err) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err != nil {
			logger.Error("error searching embedding index", log.Error(err))
			http.Error(w, fmt.Sprintf("error searching embedding index: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	})

	mux.HandleFunc("/isContextRequiredForChatQuery", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, fmt.Sprintf("unsupported method %s", r.Method), http.StatusBadRequest)
			return
		}

		var args embeddings.IsContextRequiredForChatQueryParameters
		err := json.NewDecoder(r.Body).Decode(&args)
		if err != nil {
			http.Error(w, "could not parse request body", http.StatusBadRequest)
			return
		}

		isRequired, err := isContextRequiredForChatQuery(r.Context(), getQueryEmbedding, getContextDetectionEmbeddingIndex, args.Query)
		if err != nil {
			logger.Error("error detecting if context is required for query", log.Error(err))
			http.Error(w, fmt.Sprintf("error detecting if context is required for query: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(embeddings.IsContextRequiredForChatQueryResult{IsRequired: isRequired})
	})

	return mux
}

func mustInitializeFrontendDB(observationCtx *observation.Context) *sql.DB {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.PostgresDSN
	})

	db, err := connections.EnsureNewFrontendDB(observationCtx, dsn, "embeddings")
	if err != nil {
		observationCtx.Logger.Fatal("failed to connect to database", log.Error(err))
	}

	return db
}

// SetAuthzProviders periodically refreshes the global authz providers. This changes the repositories that are visible for reads based on the
// current actor stored in an operation's context, which is likely an internal actor for many of
// the jobs configured in this service. This also enables repository update operations to fetch
// permissions from code hosts.
func setAuthzProviders(ctx context.Context, db database.DB) {
	// authz also relies on UserMappings being setup.
	globals.WatchPermissionsUserMapping()

	for range time.NewTicker(eiauthz.RefreshInterval()).C {
		allowAccessByDefault, authzProviders, _, _, _ := eiauthz.ProvidersFromConfig(ctx, conf.Get(), db.ExternalServices(), db)
		authz.SetProviders(allowAccessByDefault, authzProviders)
	}
}

func handlePanic(logger log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				err := fmt.Sprintf("%v", rec)
				http.Error(w, fmt.Sprintf("%v", rec), http.StatusInternalServerError)
				logger.Error("recovered from panic", log.String("err", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
