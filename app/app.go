package app

import (
	"io/ioutil"
	"log"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/app/appconf"
	appauth "sourcegraph.com/sourcegraph/sourcegraph/app/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
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

	var mw []handlerutil.Middleware
	mw = append(mw, httpapiauth.OAuth2AccessTokenMiddleware, appauth.CookieMiddleware)
	mw = append(mw, eventsutil.AgentMiddleware)
	mw = append(mw, internal.Middleware...)

	m := http.NewServeMux()
	if conf.GetenvBool("SG_USE_CSP") {
		cspHandler := csp.NewHandler(cspConfig)
		cspHandler.ReportLog = log.New(ioutil.Discard, "", 0)
		mw = append(mw, cspHandler.Middleware)
	}

	m.Handle("/", handlerutil.WithMiddleware(r, tmplReloadMiddleware))

	// Add git transport routes
	gitserver.AddHandlers(&r.Router)

	r.Get(router.RobotsTxt).Handler(httptrace.TraceRoute(http.HandlerFunc(robotsTxt)))
	r.Get(router.Favicon).Handler(httptrace.TraceRoute(http.HandlerFunc(favicon)))

	r.Get(router.SitemapIndex).Handler(httptrace.TraceRoute(internal.Handler(serveSitemapIndex)))
	r.Get(router.RepoSitemap).Handler(httptrace.TraceRoute(internal.Handler(serveRepoSitemap)))

	r.Get(router.Logout).Handler(httptrace.TraceRoute(internal.Handler(serveLogout)))

	// Redirects
	r.Get(router.OldToolsRedirect).Handler(httptrace.TraceRoute(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/beta", 301)
	})))

	for route, handlerFunc := range internal.Handlers {
		r.Get(route).Handler(httptrace.TraceRoute(handlerFunc))
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

func tmplReloadMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if appconf.Flags.ReloadAssets {
			tmpl.Load()
		}
		next.ServeHTTP(w, r)
	})
}

func serveLogout(w http.ResponseWriter, r *http.Request) error {
	appauth.DeleteSessionCookie(w)
	http.Redirect(w, r, "/", http.StatusFound)
	return nil
}
