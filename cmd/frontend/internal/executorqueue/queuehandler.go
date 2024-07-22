package executorqueue

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/executorqueue/handler"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/executorqueue/queues/batches"
	codeintelqueue "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/executorqueue/queues/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/executor/store"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	metricsstore "github.com/sourcegraph/sourcegraph/internal/metrics/store"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func newExecutorQueuesHandler(
	observationCtx *observation.Context,
	db database.DB,
	logger log.Logger,
	accessToken func() string,
	uploadHandler http.Handler,
	batchesWorkspaceFileGetHandler http.Handler,
	batchesWorkspaceFileExistsHandler http.Handler,
) func() http.Handler {
	metricsStore := metricsstore.NewDistributedStore("executors:")
	executorStore := db.Executors()
	jobTokenStore := store.NewJobTokenStore(observationCtx, db)

	// Register queues. If this set changes, be sure to also update the list of valid
	// queue names in ./metrics/queue_allocation.go, and register a metrics exporter
	// in the worker.
	//
	// Note: In order register a new queue type please change the validate() check code in cmd/executor/config.go
	autoIndexQueueHandler := codeintelqueue.QueueHandler(observationCtx, db, accessToken)
	batchesQueueHandler := batches.QueueHandler(observationCtx, db, accessToken)

	codeintelHandler := handler.NewHandler(executorStore, jobTokenStore, metricsStore, autoIndexQueueHandler)
	batchesHandler := handler.NewHandler(executorStore, jobTokenStore, metricsStore, batchesQueueHandler)
	handlers := []handler.ExecutorHandler{codeintelHandler, batchesHandler}

	multiHandler := handler.NewMultiHandler(executorStore, jobTokenStore, metricsStore, autoIndexQueueHandler, batchesQueueHandler)

	// Auth middleware
	executorAuth := executorAuthMiddleware(logger, accessToken)

	factory := func() http.Handler {
		// 🚨 SECURITY: These routes are secured by checking a token shared between services.
		base := mux.NewRouter().PathPrefix("/.executors/").Subrouter()
		base.StrictSlash(true)

		// Used by code_intel_test.go to test authentication HTTP status codes.
		// Also used by `executor validate` to check whether a token is set.
		testRouter := base.PathPrefix("/test").Subrouter()
		testRouter.Path("/auth").Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("ok")); err != nil {
				logger.Error("failed to test authentication", log.Error(err))
			}

		})
		testRouter.Use(withInternalActor, executorAuth)

		// Proxy /info/refs and /git-upload-pack to gitservice for git clone/fetch.
		gitRouter := base.PathPrefix("/git").Subrouter()
		gitserverClient := gitserver.NewClient("http.executor.gitproxy")
		gitRouter.Path("/{RepoName:.*}/info/refs").Handler(gitserverProxy(logger, gitserverClient, "/info/refs"))
		gitRouter.Path("/{RepoName:.*}/git-upload-pack").Handler(gitserverProxy(logger, gitserverClient, "/git-upload-pack"))
		// The git routes are treated as internal actor. Additionally, each job comes with a short-lived token that is
		// checked by jobAuthMiddleware.
		gitRouter.Use(withInternalActor, jobAuthMiddleware(logger, routeGit, jobTokenStore, executorStore))

		// Serve the executor queue APIs.
		queueRouter := base.PathPrefix("/queue").Subrouter()
		// The queue route are treated as an internal actor and require the executor access token to authenticate.
		queueRouter.Use(withInternalActor, executorAuth)
		queueRouter.Path("/dequeue").Methods(http.MethodPost).HandlerFunc(multiHandler.HandleDequeue)
		queueRouter.Path("/heartbeat").Methods(http.MethodPost).HandlerFunc(multiHandler.HandleHeartbeat)

		jobRouter := base.PathPrefix("/queue").Subrouter()
		// The job routes are treated as internal actor. Additionally, each job comes with a short-lived token that is
		// checked by jobAuthMiddleware.
		jobRouter.Use(withInternalActor, jobAuthMiddleware(logger, routeQueue, jobTokenStore, executorStore))

		for _, h := range handlers {
			handler.SetupRoutes(h, queueRouter)
			handler.SetupJobRoutes(h, jobRouter)
		}

		// Upload LSIF indexes without a sudo access token or github tokens.
		lsifRouter := base.PathPrefix("/lsif").Name("executor-lsif").Subrouter()
		lsifRouter.Path("/upload").Methods("POST").Handler(uploadHandler)
		// The lsif route are treated as an internal actor and require the executor access token to authenticate.
		lsifRouter.Use(withInternalActor, executorAuth)

		// Upload SCIP indexes without a sudo access token or github tokens.
		scipRouter := base.PathPrefix("/scip").Name("executor-scip").Subrouter()
		scipRouter.Path("/upload").Methods("POST").Handler(uploadHandler)
		scipRouter.Path("/upload").Methods("HEAD").Handler(noopHandler)
		// The scip route are treated as an internal actor and require the executor access token to authenticate.
		scipRouter.Use(withInternalActor, executorAuth)

		filesRouter := base.PathPrefix("/files").Name("executor-files").Subrouter()
		batchChangesRouter := filesRouter.PathPrefix("/batch-changes").Subrouter()
		batchChangesRouter.Path("/{spec}/{file}").Methods(http.MethodGet).Handler(batchesWorkspaceFileGetHandler)
		batchChangesRouter.Path("/{spec}/{file}").Methods(http.MethodHead).Handler(batchesWorkspaceFileExistsHandler)
		// The files route are treated as an internal actor and require the executor access token to authenticate.
		filesRouter.Use(withInternalActor, jobAuthMiddleware(logger, routeFiles, jobTokenStore, executorStore))

		return base
	}

	return factory
}

type routeName string

const (
	routeFiles = "files"
	routeGit   = "git"
	routeQueue = "queue"
)

// withInternalActor ensures that the request handling is running as an internal actor.
func withInternalActor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		next.ServeHTTP(rw, req.WithContext(actor.WithInternalActor(ctx)))
	})
}

// executorAuthMiddleware rejects requests that do not have a Authorization header set
// with the correct "token-executor <token>" value. This should only be used
// for internal _services_, not users, in which a shared key exchange can be
// done so safely.
func executorAuthMiddleware(logger log.Logger, accessToken func() string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if validateExecutorToken(w, r, logger, accessToken()) {
				next.ServeHTTP(w, r)
			}
		})
	}
}

const SchemeExecutorToken = "token-executor"

func validateExecutorToken(w http.ResponseWriter, r *http.Request, logger log.Logger, expectedAccessToken string) bool {
	if expectedAccessToken == "" {
		logger.Error("executors.accessToken not configured in site config")
		http.Error(w, "Executors are not configured on this instance", http.StatusInternalServerError)
		return false
	}

	var token string
	if headerValue := r.Header.Get("Authorization"); headerValue != "" {
		parts := strings.Split(headerValue, " ")
		if len(parts) != 2 {
			http.Error(w, fmt.Sprintf(`HTTP Authorization request header value must be of the following form: '%s "TOKEN"'`, SchemeExecutorToken), http.StatusUnauthorized)
			return false
		}
		if parts[0] != SchemeExecutorToken {
			http.Error(w, fmt.Sprintf("unrecognized HTTP Authorization request header scheme (supported values: %q)", SchemeExecutorToken), http.StatusUnauthorized)
			return false
		}

		token = parts[1]
	}
	if token == "" {
		http.Error(w, "no token value in the HTTP Authorization request header (recommended) or basic auth (deprecated)", http.StatusUnauthorized)
		return false
	}

	// 🚨 SECURITY: Use constant-time comparisons to avoid leaking the verification
	// code via timing attack. It is not important to avoid leaking the *length* of
	// the code, because the length of verification codes is constant.
	if subtle.ConstantTimeCompare([]byte(token), []byte(expectedAccessToken)) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return false
	}

	return true
}

func jobAuthMiddleware(
	logger log.Logger,
	routeName routeName,
	tokenStore store.JobTokenStore,
	executorStore database.ExecutorStore,
) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if validateJobRequest(w, r, logger, routeName, tokenStore, executorStore) {
				next.ServeHTTP(w, r)
			}
		})
	}
}

func validateJobRequest(
	w http.ResponseWriter,
	r *http.Request,
	logger log.Logger,
	routeName routeName,
	tokenStore store.JobTokenStore,
	executorStore database.ExecutorStore,
) bool {
	// Get the auth token from the Authorization header.
	var tokenType string
	var authToken string
	if headerValue := r.Header.Get("Authorization"); headerValue != "" {
		parts := strings.Split(headerValue, " ")
		if len(parts) != 2 {
			http.Error(w, fmt.Sprintf(`HTTP Authorization request header value must be of the following form: '%s "TOKEN"' or '%s TOKEN'`, "Bearer", "token-executor"), http.StatusUnauthorized)
			return false
		}
		// Check what the token type is. For backwards compatibility sake, we should also accept the general executor
		// access token.
		tokenType = parts[0]
		if tokenType != "Bearer" && tokenType != "token-executor" {
			http.Error(w, fmt.Sprintf("unrecognized HTTP Authorization request header scheme (supported values: %q, %q)", "Bearer", "token-executor"), http.StatusUnauthorized)
			return false
		}

		authToken = parts[1]
	}
	if authToken == "" {
		http.Error(w, "no token value in the HTTP Authorization request header", http.StatusUnauthorized)
		return false
	}

	// If the general executor access token was provided, simply check the value.
	if tokenType == "token-executor" {
		// 🚨 SECURITY: Use constant-time comparisons to avoid leaking the verification
		// code via timing attack. It is not important to avoid leaking the *length* of
		// the code, because the length of verification codes is constant.
		if subtle.ConstantTimeCompare([]byte(authToken), []byte(conf.SiteConfig().ExecutorsAccessToken)) == 1 {
			return true
		} else {
			w.WriteHeader(http.StatusForbidden)
			return false
		}
	}

	var executorName string
	var jobId int64
	var queue string
	var repo string
	var err error

	// Each route is "special". Set additional information based on the route that is being worked with.
	switch routeName {
	case routeFiles:
		queue = "batches"
	case routeGit:
		repo = mux.Vars(r)["RepoName"]
	case routeQueue:
		queue = mux.Vars(r)["queueName"]
	default:
		logger.Error("unsupported route", log.String("route", string(routeName)))
		http.Error(w, "unsupported route", http.StatusBadRequest)
		return false
	}

	jobId, err = parseJobIdHeader(r)
	if err != nil {
		logger.Error("failed to parse jobId", log.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	// When the requester sets a User with a username, r.URL.User.Username() will return a blank value (always).
	// To get the username is by using BasicAuth(). Even if the requester does not use a reverse proxy, this is the
	// way to get the username.
	executorName = r.Header.Get("X-Sourcegraph-Executor-Name")

	// Since the payload partially deserialize, ensure the worker hostname is valid.
	if len(executorName) == 0 {
		http.Error(w, "worker hostname cannot be empty", http.StatusBadRequest)
		return false
	}

	jobToken, err := tokenStore.GetByToken(r.Context(), authToken)
	if err != nil {
		logger.Error("failed to retrieve token", log.Error(err))
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return false
	}

	// Ensure the token was generated for the correct job.
	if jobToken.JobID != jobId {
		logger.Error("job ID does not match")
		http.Error(w, "invalid token", http.StatusForbidden)
		return false
	}

	// Check if the token is associated with the correct queue or repo.
	if len(repo) > 0 {
		if jobToken.Repo != repo {
			logger.Error("repo does not match")
			http.Error(w, "invalid token", http.StatusForbidden)
			return false
		}
	} else {
		// Ensure the token was generated for the correct queue.
		if jobToken.Queue != queue {
			logger.Error("queue name does not match")
			http.Error(w, "invalid token", http.StatusForbidden)
			return false
		}
	}
	// Ensure the token came from a legit executor instance.
	if _, _, err = executorStore.GetByHostname(r.Context(), executorName); err != nil {
		logger.Error("failed to lookup executor by hostname", log.Error(err))
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return false
	}

	return true
}

func parseJobIdHeader(r *http.Request) (int64, error) {
	jobIdHeader := r.Header.Get("X-Sourcegraph-Job-ID")
	if len(jobIdHeader) == 0 {
		return 0, errors.New("job ID not provided in header 'X-Sourcegraph-Job-ID'")
	}
	id, err := strconv.Atoi(jobIdHeader)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to parse Job ID")
	}
	return int64(id), nil
}

var noopHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})
