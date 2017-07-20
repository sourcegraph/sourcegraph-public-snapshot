package app

import (
	"fmt"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/errorutil"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/oauth2client"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/redirects"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/ui"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/ui2"
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

	r.Get(router.SitemapIndex).Handler(traceutil.TraceRoute(errorutil.Handler(serveSitemapIndex)))
	r.Get(router.RepoSitemap).Handler(traceutil.TraceRoute(errorutil.Handler(serveRepoSitemap)))
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

	// Our top level UI handler chooses between our legacy UI and our new
	// "streamlined web app" UI. A user can opt in via:
	//
	//  document.cookie="streamlined=true;path=/"
	//
	// And opt out via:
	//
	//  document.cookie="streamlined=false;path=/"
	//
	uiRouter := ui.Router()
	ui2Router := ui2.Router()
	r.Get(router.UI).Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("streamlined")
		if err == nil && cookie.Value == "true" {
			// User wants beta streamlined interface.
			ui2Router.ServeHTTP(w, r)
			return
		}
		uiRouter.ServeHTTP(w, r)
	}))

	r.Get(router.GitHubOAuth2Initiate).Handler(traceutil.TraceRoute(errorutil.Handler(oauth2client.ServeGitHubOAuth2Initiate)))
	r.Get(router.GitHubOAuth2Receive).Handler(traceutil.TraceRoute(errorutil.Handler(oauth2client.ServeGitHubOAuth2Receive)))
	r.Get(router.GitHubAppInstalled).Handler(traceutil.TraceRoute(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		githubutil.ClearCacheForCurrentUser(r.Context())
		http.Redirect(w, r, "/", 301)
	})))

	r.Get(router.GDDORefs).Handler(traceutil.TraceRoute(errorutil.Handler(serveGDDORefs)))
	r.Get(router.Editor).Handler(traceutil.TraceRoute(errorutil.Handler(serveEditor)))

	var h http.Handler = m
	h = redirects.RedirectsMiddleware(h)
	h = session.CookieMiddleware(h)
	h = httpapiauth.AuthorizationMiddleware(h)

	return h
}

func serveLogout(w http.ResponseWriter, r *http.Request) error {
	session.DeleteSession(w, r)
	http.Redirect(w, r, "/", http.StatusFound)
	return nil
}
