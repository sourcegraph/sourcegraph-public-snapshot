package app

import (
	"io/ioutil"
	"log"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/oauth2client"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/ui"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/feature"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/csp"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/eventsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httptrace"
	httpapiauth "sourcegraph.com/sourcegraph/sourcegraph/services/httpapi/auth"

	// Import for side effects.
	_ "sourcegraph.com/sourcegraph/sourcegraph/app/internal/redirects"
)

// NewHandler returns a new app handler that uses the provided app
// router (or creates a new one if nil).
func NewHandler(r *router.Router) http.Handler {
	if r == nil {
		r = router.New(nil)
	}

	auth.InitSessionStore(conf.AppURL.Scheme == "https")

	var mw []handlerutil.Middleware
	mw = append(mw, httpapiauth.AuthorizationMiddleware, auth.CookieMiddleware)
	mw = append(mw, eventsutil.AgentMiddleware)
	mw = append(mw, internal.Middleware...)

	m := http.NewServeMux()
	if conf.GetenvBool("SG_USE_CSP") {
		cspHandler := csp.NewHandler(cspConfig)
		cspHandler.ReportLog = log.New(ioutil.Discard, "", 0)
		mw = append(mw, cspHandler.Middleware)
	}

	m.Handle("/", r)

	r.Get(router.RobotsTxt).Handler(httptrace.TraceRoute(http.HandlerFunc(robotsTxt)))
	r.Get(router.Favicon).Handler(httptrace.TraceRoute(http.HandlerFunc(favicon)))

	r.Get(router.SitemapIndex).Handler(httptrace.TraceRoute(internal.Handler(serveSitemapIndex)))
	r.Get(router.RepoSitemap).Handler(httptrace.TraceRoute(internal.Handler(serveRepoSitemap)))

	r.Get(router.Logout).Handler(httptrace.TraceRoute(internal.Handler(serveLogout)))

	// Redirects
	r.Get(router.OldToolsRedirect).Handler(httptrace.TraceRoute(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/beta", 301)
	})))

	r.Get(router.UI).Handler(ui.Router())

	r.Get(router.GitHubOAuth2Initiate).Handler(httptrace.TraceRoute(internal.Handler(oauth2client.ServeGitHubOAuth2Initiate)))
	r.Get(router.GitHubOAuth2Receive).Handler(httptrace.TraceRoute(internal.Handler(oauth2client.ServeGitHubOAuth2Receive)))

	if feature.Features.GodocRefs {
		r.Get(router.GDDORefs).Handler(httptrace.TraceRoute(internal.Handler(serveGDDORefs)))
	}
	return handlerutil.WithMiddleware(m, mw...)
}

// cspConfig is the Content Security Policy config for app handlers.
var cspConfig = csp.Config{
	// Strict because API responses should never be treated as page
	// content.
	PolicyReportOnly: &csp.Policy{
		DefaultSrc: []string{"'self'"},
		FrameSrc:   []string{"https://www.youtube.com", "https://speakerdeck.com"},
		FontSrc:    []string{"'self'", "https://s3-us-west-2.amazonaws.com/sourcegraph-assets/fonts/"},
		ScriptSrc:  []string{"'self'", "https://www.google-analytics.com", "https://platform.twitter.com", "https://speakerdeck.com"},
		ImgSrc:     []string{"*"},
		StyleSrc:   []string{"*"},
		ReportURI:  "/.csp-report",
	},
}

func serveLogout(w http.ResponseWriter, r *http.Request) error {
	auth.DeleteSession(w, r)
	http.Redirect(w, r, "/", http.StatusFound)
	return nil
}
