package cli

import (
	"net/http"
	"strings"

	"github.com/NYTimes/gziphandler"
	gcontext "github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/hooks"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/assetsutil"
	internalauth "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/cli/middleware"
	internalhttpapi "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi/router"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/deviceid"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	tracepkg "github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

// newExternalHTTPHandler creates and returns the HTTP handler that serves the app and API pages to
// external clients.
func newExternalHTTPHandler(db database.DB, schema *graphql.Schema, gitHubWebhook webhooks.Registerer, gitLabWebhook, bitbucketServerWebhook http.Handler, newCodeIntelUploadHandler enterprise.NewCodeIntelUploadHandler, newExecutorProxyHandler enterprise.NewExecutorProxyHandler, rateLimitWatcher graphqlbackend.LimitWatcher) (http.Handler, error) {
	// Each auth middleware determines on a per-request basis whether it should be enabled (if not, it
	// immediately delegates the request to the next middleware in the chain).
	authMiddlewares := auth.AuthMiddleware()

	// HTTP API handler, the call order of middleware is LIFO.
	r := router.New(mux.NewRouter().PathPrefix("/.api/").Subrouter())
	apiHandler := internalhttpapi.NewHandler(db, r, schema, gitHubWebhook, gitLabWebhook, bitbucketServerWebhook, newCodeIntelUploadHandler, rateLimitWatcher)
	if hooks.PostAuthMiddleware != nil {
		// 🚨 SECURITY: These all run after the auth handler so the client is authenticated.
		apiHandler = hooks.PostAuthMiddleware(apiHandler)
	}
	apiHandler = featureflag.Middleware(database.FeatureFlags(db), apiHandler)
	apiHandler = authMiddlewares.API(apiHandler) // 🚨 SECURITY: auth middleware
	// 🚨 SECURITY: The HTTP API should not accept cookies as authentication, except from trusted
	// origins, to avoid CSRF attacks. See session.CookieMiddlewareWithCSRFSafety for details.
	apiHandler = session.CookieMiddlewareWithCSRFSafety(db, apiHandler, corsAllowHeader, isTrustedOrigin) // API accepts cookies with special header
	apiHandler = internalhttpapi.AccessTokenAuthMiddleware(db, apiHandler)                                // API accepts access tokens
	apiHandler = gziphandler.GzipHandler(apiHandler)
	if envvar.SourcegraphDotComMode() {
		apiHandler = deviceid.Middleware(apiHandler)
	}

	// 🚨 SECURITY: This handler implements its own token auth inside enterprise
	executorProxyHandler := newExecutorProxyHandler()

	// App handler (HTML pages), the call order of middleware is LIFO.
	appHandler := app.NewHandler(db)
	if hooks.PostAuthMiddleware != nil {
		// 🚨 SECURITY: These all run after the auth handler so the client is authenticated.
		appHandler = hooks.PostAuthMiddleware(appHandler)
	}
	appHandler = featureflag.Middleware(database.FeatureFlags(db), appHandler)
	appHandler = authMiddlewares.App(appHandler)                           // 🚨 SECURITY: auth middleware
	appHandler = session.CookieMiddleware(db, appHandler)                  // app accepts cookies
	appHandler = internalhttpapi.AccessTokenAuthMiddleware(db, appHandler) // app accepts access tokens
	if envvar.SourcegraphDotComMode() {
		appHandler = deviceid.Middleware(appHandler)
	}
	// Mount handlers and assets.
	sm := http.NewServeMux()
	sm.Handle("/.api/", secureHeadersMiddleware(apiHandler, crossOriginPolicyAPI))
	sm.Handle("/.executors/", secureHeadersMiddleware(executorProxyHandler, crossOriginPolicyNever))
	sm.Handle("/", secureHeadersMiddleware(appHandler, crossOriginPolicyNever))
	assetsutil.Mount(sm)

	var h http.Handler = sm

	// Wrap in middleware, first line is last to run.
	//
	// 🚨 SECURITY: Auth middleware that must run before other auth middlewares.
	// OverrideAuthMiddleware allows us to inject an authentication token via an
	// environment variable, for testing. This is true only when a site-config
	// change is explicitly made, to enable this token.
	h = middleware.Trace(h)
	h = gcontext.ClearHandler(h)
	h = healthCheckMiddleware(h)
	h = middleware.BlackHole(h)
	h = middleware.SourcegraphComGoGetHandler(h)
	h = internalauth.ForbidAllRequestsMiddleware(h)
	h = internalauth.OverrideAuthMiddleware(db, h)
	h = tracepkg.HTTPMiddleware(h, conf.DefaultClient())
	h = ot.HTTPMiddleware(h)

	return h, nil
}

func healthCheckMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/healthz", "/__version":
			_, _ = w.Write([]byte(version.Version()))
		default:
			next.ServeHTTP(w, r)
		}
	})
}

// newInternalHTTPHandler creates and returns the HTTP handler for the internal API (accessible to
// other internal services).
func newInternalHTTPHandler(schema *graphql.Schema, db database.DB, newCodeIntelUploadHandler enterprise.NewCodeIntelUploadHandler, rateLimitWatcher graphqlbackend.LimitWatcher) http.Handler {
	internalMux := http.NewServeMux()
	internalMux.Handle("/.internal/", gziphandler.GzipHandler(
		actor.HTTPMiddleware(
			featureflag.Middleware(database.FeatureFlags(db),
				internalhttpapi.NewInternalHandler(
					router.NewInternal(mux.NewRouter().PathPrefix("/.internal/").Subrouter()),
					db,
					schema,
					newCodeIntelUploadHandler,
					rateLimitWatcher,
				),
			),
		),
	))
	h := http.Handler(internalMux)
	h = gcontext.ClearHandler(h)
	h = tracepkg.HTTPMiddleware(h, conf.DefaultClient())
	h = ot.HTTPMiddleware(h)
	return h
}

// corsAllowHeader is the HTTP header that, if present (and assuming secureHeadersMiddleware is
// used), indicates that the incoming HTTP request is either same-origin or is from an allowed
// origin. See
// https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)_Prevention_Cheat_Sheet#Protecting_REST_Services:_Use_of_Custom_Request_Headers
// for more information on this technique.
const corsAllowHeader = "X-Requested-With"

// crossOriginPolicy describes the cross-origin policy the middleware should be enforcing.
type crossOriginPolicy string

const (
	// crossOriginPolicyAPI describes that the middleware should handle cross origin requests as a
	// public API. That is, cross-origin requests are allowed from any domain but
	// cookie/session-based authentication is only allowed if the origin is in the configured
	/// allow-list of origins. Otherwise, only access token authentication is permitted.
	//
	// This is to be used for all /.api routes, such as our GraphQL and search streaming APIs as we
	// want third-party websites (such as e.g. github1s.com, or internal tools for on-prem
	// customers) to be able to leverage our API. Their users will need to provide an access token,
	// or the website would need to be added to Sourcegraph's CORS allow list in order to be granted
	// cookie/session-based authentication (which is dangerous to expose to untrusted domains.)
	crossOriginPolicyAPI crossOriginPolicy = "API"

	// crossOriginPolicyNever describes that the middleware should handle cross origin requests by
	// never allowing them. This makes sense for e.g. routes such as e.g. sign out pages, where
	// cookie based authentication is needed and requests should never come from a domain other than
	// the Sourcegraph instance itself.
	//
	// Important: This only applies to cross-origin requests issued by clients that respect CORS,
	// such as browsers. So for example Code Intelligence /.executors, despite being "an API",
	// should use this policy unless they intend to get cross-origin requests _from browsers_.
	crossOriginPolicyNever crossOriginPolicy = "never"
)

// secureHeadersMiddleware adds and checks for HTTP security-related headers.
//
// 🚨 SECURITY: This handler is served to all clients, even on private servers to clients who have
// not authenticated. It must not reveal any sensitive information.
func secureHeadersMiddleware(next http.Handler, policy crossOriginPolicy) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// headers for security
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "DENY")
		// no cache by default
		w.Header().Set("Cache-Control", "no-cache, max-age=0")

		// Write CORS headers and potentially handle the requests if it is a OPTIONS request.
		if handled := handleCORSRequest(w, r, policy); handled {
			return // request was handled, do not invoke next handler
		}

		next.ServeHTTP(w, r)
	})
}

// handleCORSRequest handles checking the Origin header and writing CORS Access-Control-Allow-*
// headers. In some cases, it may handle OPTIONS CORS preflight requests in which case the function
// returns true and the request should be considered fully served.
func handleCORSRequest(w http.ResponseWriter, r *http.Request, policy crossOriginPolicy) (handled bool) {
	// If this route is one which should never allow cross-origin requests, then we should return
	// early. We do not write ANY Access-Control-Allow-* CORS headers, which triggers the browsers
	// default (and strict) behavior of not allowing cross-origin requests.
	//
	// We could instead parse the domain from conf.Get().ExternalURL and use that in the response,
	// to make things more explicit, but it would add more logic here to think about and you would
	// also want to think about whether or not `OPTIONS` requests should be handled and if the other
	// headers (-Credentials, -Methods, -Headers, etc.) should be sent back in such a situation.
	// Instead, it's easier to reason about the code by just saying "we send back nothing in this
	// case, and so the browser enforces no cross-origin requests".
	//
	// This is in compliance with section 7.2 "Resource Sharing Check" of the CORS standard: https://www.w3.org/TR/2020/SPSD-cors-20200602/#resource-sharing-check-0
	// It states:
	//
	// > If the response includes zero or more than one Access-Control-Allow-Origin header values,
	// > return fail and terminate this algorithm.
	//
	// And you may also see the type of error the browser would produce in this instance at e.g.
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS/Errors/CORSMissingAllowOrigin
	//
	if policy == crossOriginPolicyNever {
		return false
	}

	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if isTrustedOrigin(r) {
		// The request came from a trusted origin, either from the browser extension's fixed origin
		// identifier or from an origin present in the site configuration `corsOrigin` allow list.
		//
		// This means we should allow the exact origin to make the request, the browser will grant
		// their request the ability to authenticate via a session cookie.
		//
		// Note: API users also rely on this codepath handling wildcards properly. For example, if
		// Sourcegraph is behind a corporate VPN an admin may choose to set the CORS origin to "*"
		// (via a proxy, not what a browser would ever send) and would expect Sourcegraph to respond
		// appropriately to any Origin request header:
		//
		// 	"Origin: *" -> "Access-Control-Allow-Origin: *"
		// 	"Origin: https://foobar.com" -> "Access-Control-Allow-Origin: https://foobar.com"
		//
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	}

	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		// Only trusted origins are allowed to send the secure X-Requested-With and X-Sourcegraph-Client
		// headers, which indicate the client passed CORS AND is a trusted origin.
		//
		// In the future, untrusted origins will be allowed to send cross-origin requests with
		// authentication, but we will only respect session authentication iff the secure header
		// X-Requested-With is present, indicating the request came from a trusted origin or a
		// client that does not respect CORS (e.g. curl.)
		if isTrustedOrigin(r) {
			w.Header().Set("Access-Control-Allow-Headers", corsAllowHeader+", X-Sourcegraph-Client, Content-Type, Authorization, X-Sourcegraph-Should-Trace")
		} else {
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Sourcegraph-Should-Trace")
		}
		w.WriteHeader(http.StatusOK)
		return true // we handled the request
	}
	return false
}

// isTrustedOrigin returns whether the HTTP request's Origin is trusted to initiate authenticated
// cross-origin requests.
func isTrustedOrigin(r *http.Request) bool {
	requestOrigin := r.Header.Get("Origin")

	isExtensionRequest := requestOrigin == devExtension || requestOrigin == prodExtension

	var isCORSAllowedRequest bool
	if corsOrigin := conf.Get().CorsOrigin; corsOrigin != "" {
		isCORSAllowedRequest = isAllowedOrigin(requestOrigin, strings.Fields(corsOrigin))
	}

	if externalURL := strings.TrimSuffix(conf.Get().ExternalURL, "/"); externalURL != "" && requestOrigin == externalURL {
		isCORSAllowedRequest = true
	}

	return isExtensionRequest || isCORSAllowedRequest
}
