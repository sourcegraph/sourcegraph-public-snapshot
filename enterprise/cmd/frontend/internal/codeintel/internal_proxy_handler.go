package codeintel

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"path"

	"github.com/gorilla/mux"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

var indexerURL = env.Get("PRECISE_CODE_INTEL_INDEX_MANAGER_URL", "", "HTTP address for the internal precise-code-intel-indexer-manager.")
var internalProxyAuthToken = env.Get("PRECISE_CODE_INTEL_INTERNAL_PROXY_AUTH_TOKEN", "", "The auth token used to secure communication between the precise-code-intel-indexer service and the internal API provided by this proxy.")

func makeInternalProxyHandlerFactory() (func() http.Handler, error) {
	host, port, err := net.SplitHostPort(envvar.HTTPAddrInternal)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to parse internal API address '%s'", envvar.HTTPAddrInternal))
	}
	if host == "" {
		host = "127.0.0.1"
	}

	frontendOrigin, err := url.Parse(fmt.Sprintf("http://%s:%s/.internal/git", host, port))
	if err != nil {
		return nil, errors.Wrap(err, "failed to construct the origin for the internal frontend")
	}

	if indexerURL == "" {
		return nil, fmt.Errorf("invalid value for PRECISE_CODE_INTEL_INDEX_MANAGER_URL: no value supplied")
	}
	indexerOrigin, err := url.Parse(indexerURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to construct the origin for the precise-code-intel-index-manager")
	}

	factory := func() http.Handler {
		// 🚨 SECURITY: These routes are secured by checking a token shared between services.
		base := mux.NewRouter().PathPrefix("/.internal-code-intel/").Subrouter()
		base.StrictSlash(true)

		base.Path("/git/{rest:.*}").Handler(internalProxyAuthTokenMiddleware(reverseProxy(frontendOrigin)))
		base.Path("/index-queue/{rest:.*}").Handler(internalProxyAuthTokenMiddleware(reverseProxy(indexerOrigin)))
		return base
	}

	return factory, nil
}

// internalProxyAuthTokenMiddleware rejects requests that do not have a basic password matching
// the configured internal proxy auth token.
func internalProxyAuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, token, ok := r.BasicAuth()
		if !ok {
			w.Header().Add("WWW-Authenticate", `Basic realm="Sourcegraph"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if token != internalProxyAuthToken {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// TODO(efritz) - add tracing, metrics
var client = http.DefaultClient

// reverseProxy creates an HTTP handler that will proxy requests to the given target URL. See
// makeProxyRequest for details on how the request URI is constructed.
//
// Note that we do not use httputil.ReverseProxy. We need to ensure that there are no redirect
// requests to unreachable (internal) routes sent back to the client, and the built-in reverse
// proxy does not give sufficient control over redirects.
func reverseProxy(target *url.URL) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r, err := makeProxyRequest(r, target)
		if err != nil {
			log15.Error("Failed to construct proxy request", "err", err)
			http.Error(w, fmt.Sprintf("failed to construct proxy request: %s", err), http.StatusInternalServerError)
			return
		}

		resp, err := client.Do(r)
		if err != nil {
			log15.Error("Failed to perform proxy request", "err", err)
			http.Error(w, fmt.Sprintf("failed to perform proxy request: %s", err), http.StatusInternalServerError)
			return
		}

		writeResponse(w, resp)
	})
}

// getRest returns the "rest" segment of the request's URL. This is a function variable so
// we can swap it out easily during testing. The gorilla/mux does have a testing function to
// set variables on a request context, but the context gets lost somewhere between construction
// of the request and the default client's handling of the request.
var getRest = func(r *http.Request) string {
	return mux.Vars(r)["rest"]
}

// makeProxyRequest returns a new HTTP request object with the given HTTP request's headers
// and body. The resulting request's URL is a URL constructed with the given target URL as
// a base, and the text matching the "{rest:.*}" portion of the given request's route as the
// additional path segment. The resulting request's GetBody field is populated so that a
// 307 Temporary Redirect can be followed when making POST requests. This is necessary to
// properly proxy git service operations without being redirected to an inaccessible API.
func makeProxyRequest(r *http.Request, target *url.URL) (*http.Request, error) {
	getBody, err := makeReaderFactory(r.Body)
	if err != nil {
		return nil, err
	}

	u := r.URL
	u.Scheme = target.Scheme
	u.Host = target.Host
	u.Path = path.Join("/", target.Path, getRest(r))

	req, err := http.NewRequest(r.Method, u.String(), getBody())
	if err != nil {
		return nil, err
	}

	copyHeader(req.Header, r.Header)
	req.GetBody = func() (io.ReadCloser, error) { return getBody(), nil }
	return req, nil
}

// makeReaderFactory returns a function that returns a copy of the given reader on each
// invocation.
func makeReaderFactory(r io.Reader) (func() io.ReadCloser, error) {
	content, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	factory := func() io.ReadCloser {
		return ioutil.NopCloser(bytes.NewReader(content))
	}

	return factory, nil
}

// writeResponse writes the headers, status code, and body of the given response to the
// given response writer.
func writeResponse(w http.ResponseWriter, resp *http.Response) {
	defer resp.Body.Close()
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)

	if _, err := io.Copy(w, resp.Body); err != nil {
		log15.Error("Failed to write payload to client", "err", err)
	}
}

// copyHeader adds the header values from src to dst.
func copyHeader(dst, src http.Header) {
	for k, vs := range src {
		for _, v := range vs {
			dst.Add(k, v)
		}
	}
}
