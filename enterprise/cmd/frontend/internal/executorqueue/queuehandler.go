package executorqueue

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/log"
	"golang.org/x/exp/slices"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/queues/batches"
	codeintelqueue "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/queues/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor/store"
	executorstore "github.com/sourcegraph/sourcegraph/enterprise/internal/executor/store"
	executortypes "github.com/sourcegraph/sourcegraph/enterprise/internal/executor/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	metricsstore "github.com/sourcegraph/sourcegraph/internal/metrics/store"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var validQueues = []string{"batches", "codeintel"}

func validateQueues(queues []string) []string {
	var invalidQueues []string
	for _, queue := range queues {
		if !slices.Contains(validQueues, queue) {
			invalidQueues = append(invalidQueues, queue)
		}
	}
	return invalidQueues
}

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
	// Note: In order register a new queue type please change the validate() check code in enterprise/cmd/executor/config.go
	codeIntelQueueHandler := codeintelqueue.QueueHandler(observationCtx, db, accessToken)
	batchesQueueHandler := batches.QueueHandler(observationCtx, db, accessToken)

	codeintelHandler := handler.NewHandler(executorStore, jobTokenStore, metricsStore, codeIntelQueueHandler)
	batchesHandler := handler.NewHandler(executorStore, jobTokenStore, metricsStore, batchesQueueHandler)
	handlers := []handler.ExecutorHandler{codeintelHandler, batchesHandler}

	gitserverClient := gitserver.NewClient()

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
		gitRouter.Path("/{RepoName:.*}/info/refs").Handler(gitserverProxy(logger, gitserverClient, "/info/refs"))
		gitRouter.Path("/{RepoName:.*}/git-upload-pack").Handler(gitserverProxy(logger, gitserverClient, "/git-upload-pack"))
		// The git routes are treated as internal actor. Additionally, each job comes with a short-lived token that is
		// checked by jobAuthMiddleware.
		gitRouter.Use(withInternalActor, jobAuthMiddleware(logger, routeGit, jobTokenStore, executorStore))

		// Serve the executor queue APIs.
		queueRouter := base.PathPrefix("/queue").Subrouter()
		// The queue route are treated as an internal actor and require the executor access token to authenticate.
		queueRouter.Use(withInternalActor, executorAuth)

		jobRouter := base.PathPrefix("/queue").Subrouter()
		// The job routes are treated as internal actor. Additionally, each job comes with a short-lived token that is
		// checked by jobAuthMiddleware.
		jobRouter.Use(withInternalActor, jobAuthMiddleware(logger, routeQueue, jobTokenStore, executorStore))

		for _, h := range handlers {
			handler.SetupRoutes(h, queueRouter)
			handler.SetupJobRoutes(h, jobRouter)
		}

		queueRouter.Path("/dequeue").Methods(http.MethodPost).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//var req dequeueRequest
			var req dequeueRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				// TODO: should we also log errors here? Not sure
				http.Error(w, fmt.Sprintf("Failed to unmarshal payload: %s", err.Error()), http.StatusBadRequest)
			}

			// TODO: simply exported this method: I guess all of this will move into the handler package anyway so temp solution
			if err := handler.ValidateWorkerHostname(req.WorkerHostName); err != nil {
				// TODO
			}

			version2Supported := false
			if req.Version != "" {
				var err error
				version2Supported, err = api.CheckSourcegraphVersion(req.Version, "4.3.0-0", "2022-11-24")
				if err != nil {
					// TODO: should we also log errors here? Not sure
					http.Error(w, fmt.Sprintf("Failed to check Sourcegraph version: %s", err.Error()), http.StatusInternalServerError)
				}
			}

			if invalidQueues := validateQueues(req.Queues); len(invalidQueues) != 0 {
				// TODO: should we also log errors here? Not sure
				http.Error(w, fmt.Sprintf("Invalid queue name(s) '%s' found. Supported queue names are '%s'. ", strings.Join(invalidQueues, ", "), strings.Join(validQueues, ", ")), http.StatusBadRequest)
			}

			resourceMetadata := handler.ResourceMetadata{
				NumCPUs:   req.NumCPUs,
				Memory:    req.Memory,
				DiskSpace: req.DiskSpace,
			}
			var job executortypes.Job
			// TODO - impl fairness later
			for _, queue := range req.Queues {
				// TODO: basically replicating error handling of handler.dequeue() here
				switch queue {
				case "batches":
					record, _, err := batchesQueueHandler.Store.Dequeue(r.Context(), req.WorkerHostName, nil)
					if err != nil {
						logger.Error("Handler returned an error", log.Error(err))
						http.Error(w, fmt.Sprintf("Failed to dequeue from queue %s: %s", queue, errors.Wrap(err, "dbworkerstore.Dequeue").Error()), http.StatusInternalServerError)
					}

					job, err = batchesQueueHandler.RecordTransformer(r.Context(), req.Version, record, resourceMetadata)
					if err != nil {
						if _, err = batchesQueueHandler.Store.MarkFailed(r.Context(), record.RecordID(), fmt.Sprintf("failed to transform record: %s", err), dbworkerstore.MarkFinalOptions{}); err != nil {
							logger.Error("Failed to mark record as failed",
								log.Int("recordID", record.RecordID()),
								log.Error(err))
						}

						http.Error(w, fmt.Sprintf("Failed to transform %s record into job: %s", queue, errors.Wrap(err, "RecordTransformer")), http.StatusInternalServerError)
					}
				case "codeintel":
					record, _, err := codeIntelQueueHandler.Store.Dequeue(r.Context(), req.WorkerHostName, nil)
					if err != nil {
						logger.Error("Handler returned an error", log.Error(err))
						http.Error(w, fmt.Sprintf("Failed to dequeue from queue %s: %s", queue, errors.Wrap(err, "dbworkerstore.Dequeue").Error()), http.StatusInternalServerError)
					}
					job, err = codeIntelQueueHandler.RecordTransformer(r.Context(), req.Version, record, resourceMetadata)
					if err != nil {
						if _, err = codeIntelQueueHandler.Store.MarkFailed(r.Context(), record.RecordID(), fmt.Sprintf("failed to transform record: %s", err), dbworkerstore.MarkFinalOptions{}); err != nil {
							logger.Error("Failed to mark record as failed",
								log.Int("recordID", record.RecordID()),
								log.Error(err))
						}

						http.Error(w, fmt.Sprintf("Failed to transform %s record into job: %s", queue, errors.Wrap(err, "RecordTransformer")), http.StatusInternalServerError)
					}
				}
				if job.ID != 0 {
					break
				}
				// If this executor supports v2, return a v2 payload. Based on this field,
				// marshalling will be switched between old and new payload.
				if version2Supported {
					job.Version = 2
				}

				token, err := jobTokenStore.Create(r.Context(), job.ID, queue, job.RepositoryName)
				if err != nil {
					if errors.Is(err, executorstore.ErrJobTokenAlreadyCreated) {
						// Token has already been created, regen it.
						token, err = jobTokenStore.Regenerate(r.Context(), job.ID, queue)
						if err != nil {
							http.Error(w, fmt.Sprintf("Failed to regenerate token: %s", errors.Wrap(err, "RegenerateToken").Error()), http.StatusInternalServerError)
						}
					} else {
						http.Error(w, fmt.Sprintf("Failed to create token: %s", errors.Wrap(err, "CreateToken").Error()), http.StatusInternalServerError)
					}
				}
				job.Token = token
				job.Queue = queue
			}

			// TODO - does this actually work?
			if err := json.NewEncoder(w).Encode(job); err != nil {
				logger.Error("Failed to serialize payload", log.Error(err))
				http.Error(w, fmt.Sprintf("Failed to serialize payload: %s", err), http.StatusInternalServerError)
			}
		})

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

// TODO: fairly sure this is basically executortypes.DequeueRequest with Queues extended
// (and WorkerHostName == ExecutorName?)
type dequeueRequest struct {
	Queues         []string `json:"queues"`
	WorkerHostName string   `json:"workerHostName"`
	Version        string   `json:"version"`
	NumCPUs        int      `json:"numCPUs,omitempty"`
	Memory         string   `json:"memory,omitempty"`
	DiskSpace      string   `json:"diskSpace,omitempty"`
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
