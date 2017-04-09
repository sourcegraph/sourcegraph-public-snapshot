package app

import (
	"fmt"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/errorutil"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/oauth2client"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/redirects"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/ui"
	httpapiauth "sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/eventsutil"
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

	r.Get(router.SitemapIndex).Handler(traceutil.TraceRoute(errorutil.Handler(serveSitemapIndex)))
	r.Get(router.RepoSitemap).Handler(traceutil.TraceRoute(errorutil.Handler(serveRepoSitemap)))
	r.Get(router.RepoBadge).Handler(traceutil.TraceRoute(errorutil.Handler(serveRepoBadge)))

	r.Get(router.Logout).Handler(traceutil.TraceRoute(errorutil.Handler(serveLogout)))

	// Redirects
	r.Get(router.OldToolsRedirect).Handler(traceutil.TraceRoute(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/beta", 301)
	})))

	r.Get(router.UI).Handler(ui.Router())

	r.Get(router.GitHubOAuth2Initiate).Handler(traceutil.TraceRoute(errorutil.Handler(oauth2client.ServeGitHubOAuth2Initiate)))
	r.Get(router.GitHubOAuth2Receive).Handler(traceutil.TraceRoute(errorutil.Handler(oauth2client.ServeGitHubOAuth2Receive)))

	r.Get(router.InstallZap).Handler(traceutil.TraceRoute(errorutil.Handler(serveInstallZap)))

	r.Get(router.GDDORefs).Handler(traceutil.TraceRoute(errorutil.Handler(serveGDDORefs)))

	r.Get(router.ShowAuth).Handler(traceutil.TraceRoute(errorutil.Handler(serveShowAuth)))

	var h http.Handler = m
	h = redirects.RedirectsMiddleware(h)
	h = eventsutil.AgentMiddleware(h)
	h = session.CookieMiddleware(h)
	h = httpapiauth.AuthorizationMiddleware(h)

	return h
}

func serveLogout(w http.ResponseWriter, r *http.Request) error {
	session.DeleteSession(w, r)
	http.Redirect(w, r, "/", http.StatusFound)
	return nil
}
