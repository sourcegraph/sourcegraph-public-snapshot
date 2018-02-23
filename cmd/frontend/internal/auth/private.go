package auth

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/ui"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
)

// newUserRequiredAuthzHandler wraps the handler and requires an authenticated client for all HTTP requests, except those
// whitelisted by allowAnonymousRequest.
//
// 🚨 SECURITY: Any change to this function could introduce security exploits.
func newUserRequiredAuthzHandler(handler http.Handler) http.Handler {
	return session.CookieOrSessionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor := actor.FromContext(r.Context())

		// If an anonymous user tries to access an endpoint that requires authentication, prevent access and
		// redirect them to the login page.
		if !actor.IsAuthenticated() && !allowAnonymousRequest(r) {
			if isAPIRequest := strings.HasPrefix(r.URL.Path, "/.api/"); isAPIRequest {
				// Report 401 Unauthorized for API requests. Redirect 302 Found for web page requests.
				http.Error(w, "Private mode requires authentication.", http.StatusUnauthorized)
			} else {
				// Redirect 302 Found for web page requests.
				q := url.Values{}
				q.Set("returnTo", r.URL.String())
				http.Redirect(w, r, "/sign-in?"+q.Encode(), http.StatusFound)
			}
			return
		}

		// The client is authenticated, or the request is accessible to anonymous clients.
		handler.ServeHTTP(w, r)
	}))
}

var (
	// 🚨 SECURITY: These maps define route names that anonymous users can access. They MUST NOT leak any sensitive
	// data or allow unprivileged users to perform undesired actions.
	anonymousAccessibleAPIRoutes = map[string]struct{}{
		router.RobotsTxt:         struct{}{},
		router.Favicon:           struct{}{},
		router.Logout:            struct{}{},
		router.SignUp:            struct{}{},
		router.SiteInit:          struct{}{},
		router.SignIn:            struct{}{},
		router.SignOut:           struct{}{},
		router.ResetPasswordInit: struct{}{},
		router.ResetPassword:     struct{}{},
	}
	anonymousAccessibleUIRoutes = map[string]struct{}{
		ui.RouteSignIn:        struct{}{},
		ui.RouteSignUp:        struct{}{},
		ui.RoutePasswordReset: struct{}{},
	}
)

func matchedRouteName(req *http.Request, router *mux.Router) string {
	var m mux.RouteMatch
	if !router.Match(req, &m) || m.Route == nil {
		return ""
	}
	return m.Route.GetName()
}

// allowAnonymousRequest reports whether an anonymous user is allowed to make the given HTTP request.
//
// 🚨 SECURITY: This func MUST return false if handling req would leak any sensitive data or allow unprivileged
// users to perform undesired actions.
func allowAnonymousRequest(req *http.Request) bool {
	if strings.HasPrefix(req.URL.Path, "/.assets/") || strings.HasPrefix(req.URL.Path, "/.api/telemetry/log/v1/") {
		return true
	}
	apiRouteName := matchedRouteName(req, router.Router())
	if apiRouteName == router.UI {
		// Test against UI router. (Some of its handlers inject private data into the title or meta tags.)
		uiRouteName := matchedRouteName(req, ui.Router())
		_, ok := anonymousAccessibleUIRoutes[uiRouteName]
		return ok
	}
	_, ok := anonymousAccessibleAPIRoutes[apiRouteName]
	return ok
}
