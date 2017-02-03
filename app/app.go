package app

import (
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/oauth2client"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/redirects"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/ui"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/eventsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httptrace"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github"
	httpapiauth "sourcegraph.com/sourcegraph/sourcegraph/services/httpapi/auth"
)

// NewHandler returns a new app handler that uses the provided app
// router (or creates a new one if nil).
func NewHandler(r *router.Router) http.Handler {
	if r == nil {
		r = router.New(nil)
	}

	auth.InitSessionStore(conf.AppURL.Scheme == "https")

	m := http.NewServeMux()

	m.Handle("/", r)

	r.Get(router.RobotsTxt).Handler(httptrace.TraceRoute(http.HandlerFunc(robotsTxt)))
	r.Get(router.Favicon).Handler(httptrace.TraceRoute(http.HandlerFunc(favicon)))

	r.Get(router.SitemapIndex).Handler(httptrace.TraceRoute(internal.Handler(serveSitemapIndex)))
	r.Get(router.RepoSitemap).Handler(httptrace.TraceRoute(internal.Handler(serveRepoSitemap)))
	r.Get(router.RepoBadge).Handler(httptrace.TraceRoute(internal.Handler(serveRepoBadge)))

	r.Get(router.Logout).Handler(httptrace.TraceRoute(internal.Handler(serveLogout)))

	// Redirects
	r.Get(router.OldToolsRedirect).Handler(httptrace.TraceRoute(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/beta", 301)
	})))

	r.Get(router.UI).Handler(ui.Router())

	r.Get(router.GitHubOAuth2Initiate).Handler(httptrace.TraceRoute(internal.Handler(oauth2client.ServeGitHubOAuth2Initiate)))
	r.Get(router.GitHubOAuth2Receive).Handler(httptrace.TraceRoute(internal.Handler(oauth2client.ServeGitHubOAuth2Receive)))
	r.Get(router.GoogleOAuth2Initiate).Handler(httptrace.TraceRoute(internal.Handler(oauth2client.ServeGoogleOAuth2Initiate)))
	r.Get(router.GoogleOAuth2Receive).Handler(httptrace.TraceRoute(internal.Handler(oauth2client.ServeGoogleOAuth2Receive)))

	r.Get(router.GDDORefs).Handler(httptrace.TraceRoute(internal.Handler(serveGDDORefs)))

	r.Get(router.ShowAuth).Handler(httptrace.TraceRoute(internal.Handler(serveShowAuth)))

	var h http.Handler = m
	h = redirects.RedirectsMiddleware(h)
	h = eventsutil.AgentMiddleware(h)
	h = githubAuthMiddleware(h)
	h = auth.CookieMiddleware(h)
	h = httpapiauth.AuthorizationMiddleware(h)

	return h
}

func serveLogout(w http.ResponseWriter, r *http.Request) error {
	auth.DeleteSession(w, r)
	http.Redirect(w, r, "/", http.StatusFound)
	return nil
}

func githubAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := github.NewContextWithAuthedClient(r.Context())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
