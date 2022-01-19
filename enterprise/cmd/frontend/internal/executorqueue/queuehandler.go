package executorqueue

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

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
		expectedAccessToken := accessToken()
		if expectedAccessToken == "" {
			w.WriteHeader(http.StatusInternalServerError)
			log15.Error("executors.accessToken not configured in site config")
			return
		}

		token, headerSupplied, err := executorToken(r)
		if err != nil {
			if !headerSupplied {
				// This header is required to be present with 401 responses in order to prompt the client
				// to retry the request with basic auth credentials. If we do not send this header, the
				// git fetch/clone flow will break against the internal gitservice with a permanent 401.
				w.Header().Add("WWW-Authenticate", `Basic realm="Sourcegraph"`)
			}

			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if token != expectedAccessToken {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

const SchemeExecutorToken = "token-executor"

var (
	errNoTokenSupplied    = errors.New("no token value in the HTTP Authorization request header (recommended) or basic auth (deprecated)")
	errMalformedToken     = errors.Errorf(`HTTP Authorization request header value must be of the following form: '%s "TOKEN"'`, SchemeExecutorToken)
	errUnrecognizedScheme = errors.Errorf("unrecognized HTTP Authorization request header scheme (supported values: %q)", SchemeExecutorToken)
)

func executorToken(r *http.Request) (_ string, headerSupplied bool, _ error) {
	headerValue := r.Header.Get("Authorization")
	if headerValue == "" {
		if _, password, ok := r.BasicAuth(); ok {
			return password, false, nil
		}

		return "", false, errNoTokenSupplied
	}

	parts := strings.Split(headerValue, " ")
	if len(parts) != 2 {
		return "", true, errMalformedToken
	}

	if parts[0] != SchemeExecutorToken {
		return "", true, errUnrecognizedScheme
	}

	return parts[1], true, nil
}
