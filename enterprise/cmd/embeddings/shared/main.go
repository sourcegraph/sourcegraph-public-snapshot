package shared

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"

	emb "github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
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
	// sqlDB := mustInitializeFrontendDB(observationCtx)
	// db := database.NewDB(logger, sqlDB)

	// Run setup
	gitserverClient := gitserver.NewClient()
	uploadStore, err := emb.NewEmbeddingsUploadStore(ctx, observationCtx, config.EmbeddingsUploadStoreConfig)
	if err != nil {
		return err
	}

	// Create HTTP server
	handler := NewHandler(ctx, gitserverClient, uploadStore)
	handler = handlePanic(logger, handler)
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
	ctx context.Context,
	gitserverClient gitserver.Client,
	uploadStore uploadstore.Store,
) http.Handler {
	// Initialize the legacy JSON API server
	mux := http.NewServeMux()
	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		// TODO: This has to be a POST request
		// TODO: Use LRU cache to save 3 recent indices (should be thread-safe?)
		// TODO: Get repo name from query params
		// TODO: Check repo name exists
		// TODO: Convert repo name to fs-safe string + hash
		// TODO: Check if repo index exists in lru cache
		// TODO: Otherwise download index and store it in cache
		// TODO: Check sub repo permissions
		// file, err := uploadStore.Get(ctx, "index")
		// fmt.Println(err)
		// if err != nil {
		// 	// todo error
		// 	return
		// }
		// fileBytes, err := ioutil.ReadAll(file)
		// fmt.Println(err)
		// if err != nil {
		// 	// todo error
		// 	return
		// }
		// var embeddingIndex emb.EmbeddingIndex
		// err = json.Unmarshal(fileBytes, &embeddingIndex)
		// fmt.Println(err)
		// if err != nil {
		// 	return
		// }
		// fmt.Println(embeddingIndex.RepoName, embeddingIndex.Revision)
		if r.Method != "POST" {
			http.Error(w, fmt.Sprintf("unsupported method %s", r.Method), http.StatusBadRequest)
			return
		}

		var args emb.EmbeddingsSearchParameters
		err := json.NewDecoder(r.Body).Decode(&args)
		if err != nil {
			http.Error(w, "could not parse request body", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(emb.EmbeddingSearchResults{
			CodeResults: []emb.EmbeddingSearchResult{{"code/path", 0, 1, "code/content"}},
			TextResults: []emb.EmbeddingSearchResult{{"text/path", 0, 1, "text/content"}},
		})
		// fmt.Println("SEARCH!")
		// revision, err := gitserverClient.ResolveRevision(ctx, api.RepoName("github.com/sourcegraph/sourcegraph"), "", gitserver.ResolveRevisionOptions{})
		// fmt.Println(revision, err)
		// res, err := gitserverClient.ReadFile(ctx, nil, api.RepoName("github.com/sourcegraph/sourcegraph"), revision, "package.json")
		// fmt.Println(err)
		// fmt.Println(string(res))
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
