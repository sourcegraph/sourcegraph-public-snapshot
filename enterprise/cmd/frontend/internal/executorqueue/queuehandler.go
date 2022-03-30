package executorqueue

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	executor "github.com/sourcegraph/sourcegraph/internal/services/executors/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

		return actor.HTTPMiddleware(authMiddleware(accessToken, base))
	}

	return factory, nil
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
	}

	if token != expectedAccessToken {
		w.WriteHeader(http.StatusForbidden)
		return false
	}

	return true
}
