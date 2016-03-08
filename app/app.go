package app

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/justinas/nosurf"
	"sourcegraph.com/sourcegraph/csp"
	"src.sourcegraph.com/sourcegraph/app/appconf"
	appauth "src.sourcegraph.com/sourcegraph/app/auth"
	"src.sourcegraph.com/sourcegraph/app/coverage"
	"src.sourcegraph.com/sourcegraph/app/internal"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/gitserver"
	httpapiauth "src.sourcegraph.com/sourcegraph/httpapi/auth"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
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
	r.Get(router.Download).Handler(internal.Handler(serveDownload))
	r.Get(router.DownloadInstall).Handler(internal.Handler(serveDownloadInstall))

	r.Get(router.RobotsTxt).HandlerFunc(robotsTxt)
	r.Get(router.Favicon).HandlerFunc(favicon)

	r.Get(router.SitemapIndex).Handler(internal.Handler(serveSitemapIndex))

	if !appconf.Flags.DisableUserContent {
		r.Get(router.UserContent).Handler(internal.Handler(serveUserContent))
	}

	r.Get(router.Def).Handler(internal.Handler(serveDef))
	r.Get(router.DefExamples).Handler(internal.Handler(serveDefExamples))
	r.Get(router.RepoAppFrame).Handler(internal.Handler(serveRepoFrame))
	r.Get(router.Home).Handler(internal.Handler(serveHomeDashboard))
	r.Get(router.LogOut).Handler(internal.Handler(serveLogOut))

	r.Get(router.UserSettingsProfile).Handler(internal.Handler(serveUserSettingsProfile))
	r.Get(router.UserSettingsProfileAvatar).Handler(internal.Handler(serveUserSettingsProfileAvatar))

	r.Get(router.Repo).Handler(internal.Handler(serveRepo))
	r.Get(router.RepoBuild).Handler(internal.Handler(serveRepoBuild))
	r.Get(router.RepoBuildUpdate).Handler(internal.Handler(serveRepoBuildUpdate))
	r.Get(router.RepoBuildTaskLog).Handler(internal.Handler(serveRepoBuildTaskLog))
	r.Get(router.RepoBuilds).Handler(internal.Handler(serveRepoBuilds))
	r.Get(router.RepoBuildsCreate).Handler(internal.Handler(serveRepoBuildsCreate))
	r.Get(router.RepoSearch).Handler(internal.Handler(serveRepoSearch))
	r.Get(router.RepoTree).Handler(internal.Handler(serveRepoTree))
	r.Get(router.RepoSitemap).Handler(internal.Handler(serveRepoSitemap))

	r.Get(router.RepoCommit).Handler(internal.Handler(serveRepoCommit))
	r.Get(router.RepoRevCommits).Handler(internal.Handler(serveRepoCommits))
	r.Get(router.RepoTags).Handler(internal.Handler(serveRepoTags))
	r.Get(router.RepoBranches).Handler(internal.Handler(serveRepoBranches))

	// This route dispatches to registered SearchFrames.
	r.Get(router.RepoPlatformSearch).Handler(internal.Handler(serveRepoPlatformSearchResults))

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
