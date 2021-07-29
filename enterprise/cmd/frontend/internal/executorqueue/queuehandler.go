package executorqueue

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/cockroachdb/errors"
	"github.com/gorilla/mux"
	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func newExecutorQueueHandler(db dbutil.DB, observationContext *observation.Context, queueOptions map[string]handler.QueueOptions, uploadHandler http.Handler) (func() http.Handler, error) {
	host, port, err := net.SplitHostPort(envvar.HTTPAddrInternal)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to parse internal API address %q", envvar.HTTPAddrInternal))
	}

	frontendOrigin, err := url.Parse(fmt.Sprintf("http://%s:%s/.internal/git", host, port))
	if err != nil {
		return nil, errors.Wrap(err, "failed to construct the origin for the internal frontend")
	}

	factory := func() http.Handler {
		// 🚨 SECURITY: These routes are secured by checking a token shared between services.
		base := mux.NewRouter().PathPrefix("/.executors/").Subrouter()
		base.StrictSlash(true)

		// Proxy only info/refs and git-upload-pack for gitservice (git clone/fetch).
		base.Path("/git/{rest:.*/(?:info/refs|git-upload-pack)}").Handler(reverseProxy(frontendOrigin))

		// Serve the executor queue API.
		subRouter := base.PathPrefix("/queue/").Subrouter()
		initQueues(db, observationContext, queueOptions, subRouter)

		// Upload LSIF indexes without a sudo access token or github tokens.
		base.Path("/lsif/upload").Methods("POST").Handler(uploadHandler)

		return basicAuthMiddleware(base)
	}

	return factory, nil
}

func initQueues(db dbutil.DB, observationContext *observation.Context, queueOptions map[string]handler.QueueOptions, router *mux.Router) {
	for queueName, options := range queueOptions {
		// Make local copy of queue name for capture below
		queueName, store := queueName, options.Store

		prometheus.DefaultRegisterer.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name:        "src_executor_total",
			Help:        "Total number of jobs in the queued state.",
			ConstLabels: map[string]string{"queue": queueName},
		}, func() float64 {
			// TODO(efritz) - do not count soft-deleted code intel index records
			count, err := store.QueuedCount(context.Background(), nil)
			if err != nil {
				log15.Error("Failed to get queued job count", "queue", queueName, "error", err)
			}

			return float64(count)
		}))
	}

	handler.SetupRoutes(queueOptions, router)
}

// basicAuthMiddleware rejects requests that do not have a basic auth username and password matching
// the expected username and password. This should only be used for internal _services_, not users,
// in which a shared key exchange can be done so safely.
func basicAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			// This header is required to be present with 401 responses in order to prompt the client
			// to retry the request with basic auth credentials. If we do not send this header, the
			// git fetch/clone flow will break against the internal gitservice with a permanent 401.
			w.Header().Add("WWW-Authenticate", `Basic realm="Sourcegraph"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if sharedConfig.FrontendUsername == "" {
			w.WriteHeader(http.StatusInternalServerError)
			log15.Error("invalid value for EXECUTOR_FRONTEND_USERNAME: no value supplied")
			return
		}
		if sharedConfig.FrontendPassword == "" {
			w.WriteHeader(http.StatusInternalServerError)
			log15.Error("invalid value for EXECUTOR_FRONTEND_PASSWORD: no value supplied")
			return
		}
		if username != sharedConfig.FrontendUsername || password != sharedConfig.FrontendPassword {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
