package app

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/justinas/nosurf"
	"sourcegraph.com/sourcegraph/csp"
	"sourcegraph.com/sourcegraph/sourcegraph/app/appconf"
	appauth "sourcegraph.com/sourcegraph/sourcegraph/app/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/app/coverage"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	_ "sourcegraph.com/sourcegraph/sourcegraph/app/internal/ui"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/authutil"
	"sourcegraph.com/sourcegraph/sourcegraph/conf"
	httpapiauth "sourcegraph.com/sourcegraph/sourcegraph/httpapi/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/util/eventsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
	"sourcegraph.com/sourcegraph/sourcegraph/util/metricutil"
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
	h := nosurf.New(handler)
	// Prevent setting a different cookie for subpaths if someone
	// directly visits a subpath.
	h.SetBaseCookie(http.Cookie{
		Path: "/",
	})
	h.ExemptRegexps("^/login/oauth/", "git-[\\w-]+$")
	h.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpctx.SetRouteName(r, "")
		internal.HandleError(w, r, http.StatusForbidden, errors.New("CSRF check failed"))
	}))
	return h
}

// NewHandler returns a new app handler that uses the provided app
// router (or creates a new one if nil).
func NewHandler(r *router.Router) http.Handler {
	if r == nil {
		r = router.New(nil)
	}

	var mw []handlerutil.Middleware
	if authutil.ActiveFlags.HasAccessControl() {
		mw = append(mw, httpapiauth.OAuth2AccessTokenMiddleware)
	}
	if authutil.ActiveFlags.HasUserAccounts() {
		mw = append(mw, appauth.CookieMiddleware, handlerutil.UserMiddleware)
	}
	if !metricutil.DisableMetricsCollection() {
		mw = append(mw, eventsutil.AgentMiddleware)
		mw = append(mw, eventsutil.DeviceIdMiddleware)
	}
	mw = append(mw, internal.Middleware...)

	m := http.NewServeMux()
	if conf.GetenvBool("SG_USE_CSP") {
		cspHandler := csp.NewHandler(cspConfig)
		cspHandler.ReportLog = log.New(ioutil.Discard, "", 0)
		mw = append(mw, cspHandler.ServeHTTP)
	}

	m.Handle("/", handlerutil.WithMiddleware(r, tmplReloadMiddleware))

	// Add git transport routes
	gitserver.AddHandlers(&r.Router)

	r.Get(router.Builds).Handler(internal.Handler(serveBuilds))

	r.Get(router.RobotsTxt).HandlerFunc(robotsTxt)
	r.Get(router.Favicon).HandlerFunc(favicon)

	r.Get(router.SitemapIndex).Handler(internal.Handler(serveSitemapIndex))

	r.Get(router.Def).Handler(internal.Handler(serveDef))
	r.Get(router.DefRefs).Handler(internal.Handler(serveDef))
	r.Get(router.RepoAppFrame).Handler(internal.Handler(serveRepoFrame))
	r.Get(router.Home).Handler(internal.Handler(serveHomeDashboard))
	r.Get(router.LogOut).Handler(internal.Handler(serveLogOut))

	r.Get(router.UserSettingsProfile).Handler(internal.Handler(serveUserSettingsProfile))

	r.Get(router.Repo).Handler(internal.Handler(serveRepo))
	r.Get(router.RepoBuild).Handler(internal.Handler(serveRepoBuild))
	r.Get(router.RepoBuildUpdate).Handler(internal.Handler(serveRepoBuildUpdate))
	r.Get(router.RepoBuildTaskLog).Handler(internal.Handler(serveRepoBuildTaskLog))
	r.Get(router.RepoBuilds).Handler(internal.Handler(serveRepoBuilds))
	r.Get(router.RepoBuildsCreate).Handler(internal.Handler(serveRepoBuildsCreate))
	r.Get(router.RepoTree).Handler(internal.Handler(serveRepoTree))
	r.Get(router.RepoSitemap).Handler(internal.Handler(serveRepoSitemap))

	r.Get(router.RepoCommit).Handler(internal.Handler(serveRepoCommit))
	r.Get(router.RepoRevCommits).Handler(internal.Handler(serveRepoCommits))
	r.Get(router.RepoTags).Handler(internal.Handler(serveRepoTags))
	r.Get(router.RepoBranches).Handler(internal.Handler(serveRepoBranches))

	for route, handlerFunc := range internal.Handlers {
		r.Get(route).Handler(internal.Handler(handlerFunc))
	}

	coverage.AddRoutes(r)

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

func tmplReloadMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if appconf.Flags.ReloadAssets {
		tmpl.Load()
	}
	next(w, r)
}
