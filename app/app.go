package app

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/justinas/nosurf"
	"github.com/sourcegraph/mux"

	"sourcegraph.com/sourcegraph/csp"
	"src.sourcegraph.com/sourcegraph/app/appconf"
	appauth "src.sourcegraph.com/sourcegraph/app/auth"
	"src.sourcegraph.com/sourcegraph/app/internal"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/conf/feature"
	"src.sourcegraph.com/sourcegraph/gitserver"
	httpapiauth "src.sourcegraph.com/sourcegraph/httpapi/auth"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"

	_ "src.sourcegraph.com/sourcegraph/app/internal/markdown"
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

	mw := []handlerutil.Middleware{
		AssetsMiddleware(),
		handlerutil.CacheMiddleware,
		appauth.CookieMiddleware,
		httpapiauth.OAuth2AccessTokenMiddleware,
		handlerutil.UserMiddleware,
	}
	mw = append(mw, internal.Middleware...)

	m := http.NewServeMux()
	if conf.GetenvBool("SG_USE_CSP") {
		cspHandler := csp.NewHandler(cspConfig)
		cspHandler.ReportLog = log.New(ioutil.Discard, "", 0)
		mw = append(mw, cspHandler.ServeHTTP)
	}

	m.Handle("/_/route/", http.StripPrefix("/_/route", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var rtmatch mux.RouteMatch
		matched := r.Match(req, &rtmatch)
		if matched {
			_, err := w.Write([]byte(rtmatch.Route.GetName()))
			if err != nil {
				log.Printf("failed to write to response for route request for %s: %s", "/_/route"+req.URL.String(), err)
			}
		} else {
			http.Error(w, "", http.StatusNotFound)
		}
	})))

	m.Handle("/", handlerutil.WithMiddleware(r, tmplReloadMiddleware))

	// Add git transport routes
	gitserver.AddHandlers(&r.Router)

	// Set handlers for the installed routes.
	if appconf.Flags.Blog {
		r.Get(router.BlogIndex).Handler(internal.Handler(serveBlogIndex))
		r.Get(router.BlogIndexAtom).Handler(internal.Handler(serveBlogIndexAtom))
		r.Get(router.BlogPost).Handler(internal.Handler(serveBlogPost))
		r.Get(router.Liveblog).Handler(liveblogHandler)
	}

	r.Get(router.Builds).Handler(internal.Handler(serveBuilds))
	r.Get(router.Download).Handler(internal.Handler(serveDownload))
	r.Get(router.DownloadInstall).Handler(internal.Handler(serveDownloadInstall))
	r.Get(router.RepoCreate).Handler(internal.Handler(serveRepoCreate))

	r.Get(router.SitemapIndex).Handler(internal.Handler(serveSitemapIndex))

	if !appconf.Flags.DisableUserContent {
		r.Get(router.UserContent).Handler(internal.Handler(serveUserContent))
	}

	r.Get(router.Def).Handler(internal.Handler(serveDef))
	r.Get(router.DefExamples).Handler(internal.Handler(serveDefExamples))
	r.Get(router.DefPopover).Handler(internal.Handler(serveDefPopover))
	r.Get(router.DefShare).Handler(internal.Handler(serveDefShare))
	r.Get(router.RepoAppFrame).Handler(internal.Handler(serveRepoFrame))
	r.Get(router.Home).Handler(internal.Handler(serveHomeDashboard))
	r.Get(router.LogOut).Handler(internal.Handler(serveLogOut))

	r.Get(router.UserSettingsProfile).Handler(internal.Handler(serveUserSettingsProfile))
	r.Get(router.UserSettingsEmails).Handler(internal.Handler(serveUserSettingsEmails))
	if !appconf.Flags.DisableIntegrations {
		r.Get(router.UserSettingsIntegrations).Handler(internal.Handler(serveUserSettingsIntegrations))
		r.Get(router.UserSettingsIntegrationsUpdate).Handler(internal.Handler(serveUserSettingsIntegrationsUpdate))
	}
	if !authutil.ActiveFlags.DisableUserProfiles {
		r.Get(router.User).Handler(internal.Handler(serveUser))
		r.Get(router.UserOrgs).Handler(internal.Handler(serveUserOrgs))
		r.Get(router.OrgMembers).Handler(internal.Handler(serveOrgMembers))
	}

	r.Get(router.Repo).Handler(internal.Handler(serveRepo))
	r.Get(router.RepoBuild).Handler(internal.Handler(serveRepoBuild))
	r.Get(router.RepoBuildUpdate).Handler(internal.Handler(serveRepoBuildUpdate))
	r.Get(router.RepoBuildLog).Handler(internal.Handler(serveRepoBuildLog))
	r.Get(router.RepoBuildTaskLog).Handler(internal.Handler(serveRepoBuildTaskLog))
	r.Get(router.RepoBuilds).Handler(internal.Handler(serveRepoBuilds))
	r.Get(router.RepoBuildsCreate).Handler(internal.Handler(serveRepoBuildsCreate))
	r.Get(router.RepoCompare).Handler(internal.Handler(serveRepoCompare))
	r.Get(router.RepoCompareAll).Handler(internal.Handler(serveRepoCompare))
	r.Get(router.RepoRefresh).Handler(internal.Handler(serveRepoRefresh))
	r.Get(router.RepoSearch).Handler(internal.Handler(serveRepoSearch))
	r.Get(router.RepoTree).Handler(internal.Handler(serveRepoTree))
	r.Get(router.RepoSitemap).Handler(internal.Handler(serveRepoSitemap))
	r.Get(router.RepoTreeShare).Handler(internal.Handler(serveRepoTreeShare))
	r.Get(router.SourceboxDef).Handler(internal.Handler(serveSourceboxDef))
	r.Get(router.SourceboxFile).Handler(internal.Handler(serveSourceboxFile))

	if appconf.Flags.RepoBadgesAndCounters {
		r.Get(router.RepoBadge).Handler(internal.Handler(serveRepoBadge))
		r.Get(router.RepoBadges).Handler(internal.Handler(serveRepoBadges))
		r.Get(router.RepoCounter).Handler(internal.Handler(serveRepoCounter))
		r.Get(router.RepoCounters).Handler(internal.Handler(serveRepoCounters))
	}

	r.Get(router.RepoCommit).Handler(internal.Handler(serveRepoCommit))
	r.Get(router.RepoRevCommits).Handler(internal.Handler(serveRepoCommits))
	r.Get(router.RepoTags).Handler(internal.Handler(serveRepoTags))
	r.Get(router.RepoBranches).Handler(internal.Handler(serveRepoBranches))

	// This route dispatches to registered SearchFrames.
	r.Get(router.RepoPlatformSearch).Handler(internal.Handler(serveRepoPlatformSearchResults))

	for route, handlerFunc := range internal.Handlers {
		r.Get(route).Handler(internal.Handler(handlerFunc))
	}

	if feature.Features.NotificationCenter {
		r.Get(router.AppGlobalNotificationCenter).Handler(internal.Handler(serveAppGlobalNotificationCenter))
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
		ScriptSrc: []string{"'self'", "https://www.google-analytics.com", "heapanalytics.com", "https://cdn.heapanalytics.com", "https://platform.twitter.com", "https://speakerdeck.com",
			"https://resonancelabs.github.io", // Traceguide JS
			"'unsafe-eval'",                   // Required for Heap Analytics JS (their external script requires eval).
		},
		ImgSrc:     []string{"*"},
		StyleSrc:   []string{"*"},
		ConnectSrc: []string{fmt.Sprintf("%s:9997", os.Getenv("SG_TRACEGUIDE_SERVICE_HOST"))},
		ReportURI:  "/.csp-report",
	},
}

func init() {
	if UseWebpackDevServer {
		cspConfig.PolicyReportOnly.ScriptSrc = append(cspConfig.PolicyReportOnly.ScriptSrc, "localhost:8080")
		cspConfig.PolicyReportOnly.FontSrc = append(cspConfig.PolicyReportOnly.FontSrc, "localhost:8080")
		cspConfig.PolicyReportOnly.ConnectSrc = append(cspConfig.PolicyReportOnly.ConnectSrc, "localhost:3080", "localhost:8080", "ws://localhost:8080")
	}
}

func tmplReloadMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if appconf.Flags.ReloadAssets {
		tmpl.Load()
	}
	next(w, r)
}
