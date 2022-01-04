package executorqueue

import (
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/cockroachdb/errors"
	"github.com/gorilla/mux"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	executor "github.com/sourcegraph/sourcegraph/internal/services/executors/store"
)

func newExecutorQueueHandler(executorStore executor.Store, queueOptions []handler.QueueOptions, accessToken func() string, uploadHandler http.Handler) (func() http.Handler, error) {
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
		handler.SetupRoutes(executorStore, queueOptions, base.PathPrefix("/queue/").Subrouter())

		// Upload LSIF indexes without a sudo access token or github tokens.
		base.Path("/lsif/upload").Methods("POST").Handler(uploadHandler)

		return basicAuthMiddleware(accessToken, base)
	}

	return factory, nil
}

// basicAuthMiddleware rejects requests that do not have a basic auth username and password matching
// the expected username and password. This should only be used for internal _services_, not users,
// in which a shared key exchange can be done so safely.
func basicAuthMiddleware(accessToken func() string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// We don't care about the username. Only the password matters here.
		_, password, ok := r.BasicAuth()
		if !ok {
			// This header is required to be present with 401 responses in order to prompt the client
			// to retry the request with basic auth credentials. If we do not send this header, the
			// git fetch/clone flow will break against the internal gitservice with a permanent 401.
			w.Header().Add("WWW-Authenticate", `Basic realm="Sourcegraph"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		ac := accessToken()
		if ac == "" {
			w.WriteHeader(http.StatusInternalServerError)
			log15.Error("executors.accessToken not configured in site config")
			return
		}
		if password != ac {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
