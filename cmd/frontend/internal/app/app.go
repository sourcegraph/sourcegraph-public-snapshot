package app

import (
	"fmt"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/errorutil"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/oauth2client"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/redirects"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/ui2"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth0"
	httpapiauth "sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/githubutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
)

// NewHandler returns a new app handler that uses the provided app
// router.
func NewHandler(r *router.Router) http.Handler {
	session.InitSessionStore(conf.AppURL.Scheme == "https")

	m := http.NewServeMux()

	m.Handle("/", r)

	m.Handle("/__version", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, env.Version)
	}))

	r.Get(router.RobotsTxt).Handler(traceutil.TraceRoute(http.HandlerFunc(robotsTxt)))
	r.Get(router.Favicon).Handler(traceutil.TraceRoute(http.HandlerFunc(favicon)))
	r.Get(router.OpenSearch).Handler(traceutil.TraceRoute(http.HandlerFunc(openSearch)))

	r.Get(router.RepoBadge).Handler(traceutil.TraceRoute(errorutil.Handler(serveRepoBadge)))

	r.Get(router.Logout).Handler(traceutil.TraceRoute(errorutil.Handler(serveLogout)))

	// Redirects
	r.Get(router.OldToolsRedirect).Handler(traceutil.TraceRoute(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/beta", 301)
	})))

	r.Get(router.GopherconLiveBlog).Handler(traceutil.TraceRoute(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://about.sourcegraph.com/go", 302)
	})))

	r.Get(router.GoSymbolURL).Handler(traceutil.TraceRoute(errorutil.Handler(serveGoSymbolURL)))

	r.Get(router.UI).Handler(ui2.Router())

	signInURL := "https://" + auth0.Domain + "/authorize?response_type=code&client_id=" + auth0.Config.ClientID + "&connection=Sourcegraph&redirect_uri=" + conf.AppURL.String() + "/-/auth0/sign-in"
	r.Get(router.SignIn).Handler(traceutil.TraceRoute(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, signInURL, http.StatusSeeOther)
	})))
	r.Get(router.SignOut).Handler(traceutil.TraceRoute(http.HandlerFunc(serveSignOut)))
	r.Get(router.Auth0Signin).Handler(traceutil.TraceRoute(errorutil.Handler(ServeAuth0SignIn)))

	r.Get(router.GitHubOAuth2Initiate).Handler(traceutil.TraceRoute(errorutil.Handler(oauth2client.ServeGitHubOAuth2Initiate))) // DEPRECATED
	r.Get(router.GitHubOAuth2Receive).Handler(traceutil.TraceRoute(errorutil.Handler(oauth2client.ServeGitHubOAuth2Receive)))   // DEPRECATED
	r.Get(router.GitHubAppInstalled).Handler(traceutil.TraceRoute(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		githubutil.ClearCacheForCurrentUser(r.Context())
		http.Redirect(w, r, "/", 301)
	})))

	r.Get(router.GDDORefs).Handler(traceutil.TraceRoute(errorutil.Handler(serveGDDORefs)))
	r.Get(router.Editor).Handler(traceutil.TraceRoute(errorutil.Handler(serveEditor)))

	r.Get(router.DebugHeaders).Handler(traceutil.TraceRoute(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del("Cookie")
		r.Header.Write(w)
	})))

	var h http.Handler = m
	h = redirects.RedirectsMiddleware(h)
	h = session.CookieMiddleware(h)
	h = httpapiauth.AuthorizationMiddleware(h)

	return h
}

func serveSignOut(w http.ResponseWriter, r *http.Request) {
	session.DeleteSession(w, r)
	http.Redirect(w, r, "https://"+auth0.Domain+"/v2/logout?"+conf.AppURL.String(), http.StatusSeeOther)
}

// DEPRECATED
func serveLogout(w http.ResponseWriter, r *http.Request) error {
	session.DeleteSession(w, r)
	http.Redirect(w, r, "/", http.StatusFound)
	return nil
}
