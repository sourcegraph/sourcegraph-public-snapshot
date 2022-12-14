package executorqueue

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	metricsstore "github.com/sourcegraph/sourcegraph/internal/metrics/store"
)

func newExecutorQueueHandler(logger log.Logger, db database.DB, queueHandlers []handler.ExecutorHandler, accessToken func() string, uploadHandler http.Handler, batchesWorkspaceFileGetHandler http.Handler, batchesWorkspaceFileExistsHandler http.Handler) func() http.Handler {
	metricsStore := metricsstore.NewDistributedStore("executors:")
	executorStore := db.Executors()
	gitserverClient := gitserver.NewClient(db)

	factory := func() http.Handler {
		// 🚨 SECURITY: These routes are secured by checking a token shared between services.
		base := mux.NewRouter().PathPrefix("/.executors/").Subrouter()
		base.StrictSlash(true)

		// Proxy /info/refs and /git-upload-pack to gitservice for git clone/fetch.
		base.Path("/git/{RepoName:.*}/info/refs").Handler(gitserverProxy(logger, gitserverClient, "/info/refs"))
		base.Path("/git/{RepoName:.*}/git-upload-pack").Handler(gitserverProxy(logger, gitserverClient, "/git-upload-pack"))

		// Serve the executor queue API.
		handler.SetupRoutes(executorStore, metricsStore, queueHandlers, base.PathPrefix("/queue/").Subrouter())

		// Upload LSIF indexes without a sudo access token or github tokens.
		base.Path("/lsif/upload").Methods("POST").Handler(uploadHandler)

		base.Path("/files/batch-changes/{spec}/{file}").Methods("GET").Handler(batchesWorkspaceFileGetHandler)
		base.Path("/files/batch-changes/{spec}/{file}").Methods("HEAD").Handler(batchesWorkspaceFileExistsHandler)

		// Make sure requests to these endpoints are treated as an internal actor.
		// We treat executors as internal and the executor secret is an internal actor
		// access token.
		// Also ensure that proper executor authentication is provided.
		return authMiddleware(accessToken, withInternalActor(base))
	}

	return factory
}

// authMiddleware rejects requests that do not have a Authorization header set
// with the correct "token-executor <token>" value. This should only be used
// for internal _services_, not users, in which a shared key exchange can be
// done so safely.
func authMiddleware(accessToken func() string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if validateExecutorToken(w, r, accessToken()) {
			next.ServeHTTP(w, r)
		}
	})
}

const SchemeExecutorToken = "token-executor"

func validateExecutorToken(w http.ResponseWriter, r *http.Request, expectedAccessToken string) bool {
	if expectedAccessToken == "" {
		log15.Error("executors.accessToken not configured in site config")
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

	if token != expectedAccessToken {
		w.WriteHeader(http.StatusForbidden)
		return false
	}

	return true
}

// withInternalActor ensures that the request handling is running as an internal actor.
func withInternalActor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		next.ServeHTTP(rw, req.WithContext(actor.WithInternalActor(ctx)))
	})
}
