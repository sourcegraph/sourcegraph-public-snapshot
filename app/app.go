package app

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/justinas/nosurf"

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
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httputil/httpctx"
	httpapiauth "sourcegraph.com/sourcegraph/sourcegraph/services/httpapi/auth"

	// Import for side effects.
	_ "sourcegraph.com/sourcegraph/sourcegraph/app/internal/redirects"
)

// NewHandlerWithCSRFProtection creates a new handler that uses the provided
// handler. It additionally adds support for cross-site request forgery. To make
// your forms compliant you will have to include a hidden input which contains
// the CSRFToken that is made available to you in the template via tmpl.Common.
//
// Example:
// 	<input type="hidden" name="csrf_token" value="{{$.CSRFToken}}">
//
func NewHandlerWithCSRFProtection(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := nosurf.New(handler)
		// Prevent setting a different cookie for subpaths if someone
		// directly visits a subpath.
		h.SetBaseCookie(http.Cookie{
			Path:     "/",
			HttpOnly: true,
			Secure:   appauth.OnlySecureCookies(httpctx.FromRequest(r)),
		})
		h.ExemptRegexps("^/login/oauth/", "git-[\\w-]+$")
		h.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpctx.SetRouteName(r, "")
			internal.HandleError(w, r, http.StatusForbidden, errors.New("CSRF check failed"))
		}))
		h.ServeHTTP(w, r)
	})
}

// NewHandler returns a new app handler that uses the provided app
// router (or creates a new one if nil).
func NewHandler(r *router.Router) http.Handler {
	if r == nil {
		r = router.New(nil)
	}

	var mw []handlerutil.Middleware
	mw = append(mw, httpapiauth.OAuth2AccessTokenMiddleware, appauth.CookieMiddleware, handlerutil.ActorMiddleware)
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

	r.Get(router.RobotsTxt).HandlerFunc(robotsTxt)
	r.Get(router.Favicon).HandlerFunc(favicon)

	r.Get(router.SitemapIndex).Handler(internal.Handler(serveSitemapIndex))

	r.Get(router.RepoSitemap).Handler(internal.Handler(serveRepoSitemap))

	for route, handlerFunc := range internal.Handlers {
		r.Get(route).Handler(handlerFunc)
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
