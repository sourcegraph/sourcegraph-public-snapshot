package ui

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/NYTimes/gziphandler"
	"github.com/gorilla/mux"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	uirouter "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/ui/router"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/randstring"
	"github.com/sourcegraph/sourcegraph/pkg/routevar"
	"github.com/sourcegraph/sourcegraph/pkg/trace"
)

const (
	routeHome           = "home"
	routeStart          = "start"
	routeSearch         = "search"
	routeSearchBadge    = "search-badge"
	routeSearchSearches = "search-searches"
	routeOpen           = "open"
	routeRepo           = "repo"
	routeRepoSettings   = "repo-settings"
	routeRepoCommit     = "repo-commit"
	routeRepoBranches   = "repo-branches"
	routeRepoCommits    = "repo-commits"
	routeRepoTags       = "repo-tags"
	routeRepoCompare    = "repo-compare"
	routeRepoStats      = "repo-stats"
	routeRepoGraph      = "repo-graph"
	routeChecks         = "checks"
	routeCodemods       = "codemods"
	routeChanges        = "changes"
	routeThreads        = "threads"
	routeProject        = "project"
	routeTree           = "tree"
	routeBlob           = "blob"
	routeRaw            = "raw"
	routeOrganizations  = "org"
	routeSettings       = "settings"
	routeSiteAdmin      = "site-admin"
	routeDiscussions    = "discussions"
	routeAPIConsole     = "api-console"
	routeSearchScope    = "scope"
	routeUser           = "user"
	routeUserSettings   = "user-settings"
	routeUserRedirect   = "user-redirect"
	routeAboutSubdomain = "about-subdomain"
	aboutRedirectScheme = "https"
	aboutRedirectHost   = "about.sourcegraph.com"
	routeSurvey         = "survey"
	routeSurveyScore    = "survey-score"
	routeRegistry       = "registry"
	routeExtensions     = "extensions"
	routeHelp           = "help"
	routeExplore        = "explore"
	routeWelcome        = "welcome"
	routeSnippets       = "snippets"
	routeSubscriptions  = "subscriptions"

	// Legacy redirects
	routeLegacyLogin                   = "login"
	routeLegacyCareers                 = "careers"
	routeLegacyDefLanding              = "page.def.landing"
	routeLegacyOldRouteDefLanding      = "page.def.landing.old"
	routeLegacyRepoLanding             = "page.repo.landing"
	routeLegacyDefRedirectToDefLanding = "page.def.redirect"
	routeLegacyEditorAuth              = "legacy.editor-auth"
	routeLegacyEditorAuth2             = "legacy.editor-auth2"
	routeLegacySearchQueries           = "search-queries"
)

// aboutRedirects contains map entries, each of which indicates that
// sourcegraph.com/$KEY should redirect to about.sourcegraph.com/$VALUE.
var aboutRedirects = map[string]string{
	"about":      "about",
	"plan":       "plan",
	"contact":    "contact",
	"docs":       "docs",
	"enterprise": "enterprise",
	"pricing":    "pricing",
	"privacy":    "privacy",
	"security":   "security",
	"terms":      "terms",
	"jobs":       "jobs",
	"beta":       "beta",
	"server":     "products/server",
}

// Router returns the router that serves pages for our web app.
func Router() *mux.Router {
	return uirouter.Router
}

var mockServeRepo func(w http.ResponseWriter, r *http.Request)

func newRouter() *mux.Router {
	r := mux.NewRouter()
	r.StrictSlash(true)

	r.Path("/settings/editor-auth").Methods("GET").Name(routeLegacyEditorAuth2)

	// Top-level routes.
	r.Path("/").Methods("GET").Name(routeHome)
	r.Path("/start").Methods("GET").Name(routeStart)
	r.PathPrefix("/welcome").Methods("GET").Name(routeWelcome)
	r.Path("/search").Methods("GET").Name(routeSearch)
	r.Path("/search/badge").Methods("GET").Name(routeSearchBadge)
	r.Path("/search/searches").Methods("GET").Name(routeSearchSearches)
	r.Path("/open").Methods("GET").Name(routeOpen)
	r.Path("/sign-in").Methods("GET").Name(uirouter.RouteSignIn)
	r.Path("/sign-up").Methods("GET").Name(uirouter.RouteSignUp)
	r.PathPrefix("/organizations").Methods("GET").Name(routeOrganizations)
	r.PathPrefix("/settings").Methods("GET").Name(routeSettings)
	r.PathPrefix("/site-admin").Methods("GET").Name(routeSiteAdmin)
	r.Path("/password-reset").Methods("GET").Name(uirouter.RoutePasswordReset)
	r.Path("/discussions").Methods("GET").Name(routeDiscussions)
	r.Path("/api/console").Methods("GET").Name(routeAPIConsole)
	r.Path("/{Path:(?:" + strings.Join(mapKeys(aboutRedirects), "|") + ")}").Methods("GET").Name(routeAboutSubdomain)
	r.Path("/search/scope/{scope}").Methods("GET").Name(routeSearchScope)
	r.PathPrefix("/users/{username}/settings").Methods("GET").Name(routeUserSettings)
	r.PathPrefix("/users/{username}").Methods("GET").Name(routeUser)
	r.PathPrefix("/user").Methods("GET").Name(routeUserRedirect)
	r.Path("/survey").Methods("GET").Name(routeSurvey)
	r.Path("/survey/{score}").Methods("GET").Name(routeSurveyScore)
	r.PathPrefix("/registry").Methods("GET").Name(routeRegistry)
	r.PathPrefix("/extensions").Methods("GET").Name(routeExtensions)
	r.PathPrefix("/help").Methods("GET").Name(routeHelp)
	r.PathPrefix("/explore").Methods("GET").Name(routeExplore)
	r.PathPrefix("/snippets").Methods("GET").Name(routeSnippets)
	r.PathPrefix("/subscriptions").Methods("GET").Name(routeSubscriptions)
	r.PathPrefix("/threads").Methods("GET").Name(routeThreads)
	r.PathPrefix("/checks").Methods("GET").Name(routeChecks)
	r.PathPrefix("/codemods").Methods("GET").Name(routeCodemods)
	r.PathPrefix("/change").Methods("GET").Name(routeChanges)
	r.PathPrefix("/p").Methods("GET").Name(routeProject)

	// Legacy redirects
	r.Path("/login").Methods("GET").Name(routeLegacyLogin)
	r.Path("/careers").Methods("GET").Name(routeLegacyCareers)
	r.Path("/editor-auth").Methods("GET").Name(routeLegacyEditorAuth)
	r.Path("/search/queries").Methods("GET").Name(routeLegacySearchQueries)

	// repo
	repoRevPath := "/" + routevar.Repo + routevar.RepoRevSuffix
	r.Path(repoRevPath).Methods("GET").Name(routeRepo)

	// tree
	repoRev := r.PathPrefix(repoRevPath + "/" + routevar.RepoPathDelim).Subrouter()
	repoRev.Path("/tree{Path:.*}").Methods("GET").Name(routeTree)

	repoRev.PathPrefix("/commits").Methods("GET").Name(routeRepoCommits)
	repoRev.PathPrefix("/graph").Methods("GET").Name(routeRepoGraph)

	// blob
	repoRev.Path("/blob{Path:.*}").Methods("GET").Name(routeBlob)

	// raw
	repoRev.Path("/raw{Path:.*}").Methods("GET", "HEAD").Name(routeRaw)

	repo := r.PathPrefix(repoRevPath + "/" + routevar.RepoPathDelim).Subrouter()
	repo.PathPrefix("/settings").Methods("GET").Name(routeRepoSettings)
	repo.PathPrefix("/commit").Methods("GET").Name(routeRepoCommit)
	repo.PathPrefix("/branches").Methods("GET").Name(routeRepoBranches)
	repo.PathPrefix("/tags").Methods("GET").Name(routeRepoTags)
	repo.PathPrefix("/compare").Methods("GET").Name(routeRepoCompare)
	repo.PathPrefix("/stats").Methods("GET").Name(routeRepoStats)

	// legacy redirects
	repo.Path("/info").Methods("GET").Name(routeLegacyRepoLanding)
	repoRev.Path("/{dummy:def|refs}/" + routevar.Def).Methods("GET").Name(routeLegacyDefRedirectToDefLanding)
	repoRev.Path("/info/" + routevar.Def).Methods("GET").Name(routeLegacyDefLanding)
	repoRev.Path("/land/" + routevar.Def).Methods("GET").Name(routeLegacyOldRouteDefLanding)
	return r
}

func init() {
	initRouter()
}

func initRouter() {
	// basic pages with static titles
	router := newRouter()
	uirouter.Router = router // make accessible to other packages
	router.Get(routeHome).Handler(handler(serveHome))
	router.Get(routeStart).Handler(staticRedirectHandler("https://about.sourcegraph.com/", http.StatusMovedPermanently))
	router.Get(routeWelcome).Handler(staticRedirectHandler("https://about.sourcegraph.com/", http.StatusMovedPermanently))
	router.Get(routeThreads).Handler(handler(serveBasicPageString("Threads - Sourcegraph")))
	router.Get(uirouter.RouteSignIn).Handler(handler(serveSignIn))
	router.Get(uirouter.RouteSignUp).Handler(handler(serveBasicPageString("Sign up - Sourcegraph")))
	router.Get(routeOrganizations).Handler(handler(serveBasicPageString("Organization - Sourcegraph")))
	router.Get(routeSettings).Handler(handler(serveBasicPageString("Settings - Sourcegraph")))
	router.Get(routeSiteAdmin).Handler(handler(serveBasicPageString("Admin - Sourcegraph")))
	router.Get(uirouter.RoutePasswordReset).Handler(handler(serveBasicPageString("Reset password - Sourcegraph")))
	router.Get(routeDiscussions).Handler(handler(serveBasicPageString("Discussions - Sourcegraph")))
	router.Get(routeAPIConsole).Handler(handler(serveBasicPageString("API explorer - Sourcegraph")))
	router.Get(routeRepoSettings).Handler(handler(serveBasicPageString("Repository settings - Sourcegraph")))
	router.Get(routeRepoCommit).Handler(handler(serveBasicPageString("Commit - Sourcegraph")))
	router.Get(routeRepoBranches).Handler(handler(serveBasicPageString("Branches - Sourcegraph")))
	router.Get(routeRepoCommits).Handler(handler(serveBasicPageString("Commits - Sourcegraph")))
	router.Get(routeRepoTags).Handler(handler(serveBasicPageString("Tags - Sourcegraph")))
	router.Get(routeRepoCompare).Handler(handler(serveBasicPageString("Compare - Sourcegraph")))
	router.Get(routeRepoStats).Handler(handler(serveBasicPageString("Stats - Sourcegraph")))
	router.Get(routeRepoGraph).Handler(handler(serveBasicPageString("Repository graph - Sourcegraph")))
	router.Get(routeThreads).Handler(handler(serveBasicPageString("Threads - Sourcegraph")))
	router.Get(routeChecks).Handler(handler(serveBasicPageString("Checks - Sourcegraph")))
	router.Get(routeCodemods).Handler(handler(serveBasicPageString("Codemods - Sourcegraph")))
	router.Get(routeChanges).Handler(handler(serveBasicPageString("Changes - Sourcegraph")))
	router.Get(routeSearchScope).Handler(handler(serveBasicPageString("Search scope - Sourcegraph")))
	router.Get(routeSurvey).Handler(handler(serveBasicPageString("Survey - Sourcegraph")))
	router.Get(routeSurveyScore).Handler(handler(serveBasicPageString("Survey - Sourcegraph")))
	router.Get(routeRegistry).Handler(handler(serveBasicPageString("Registry - Sourcegraph")))
	router.Get(routeExtensions).Handler(handler(serveBasicPageString("Extensions - Sourcegraph")))
	router.Get(routeExplore).Handler(handler(serveBasicPageString("Explore - Sourcegraph")))
	router.Get(routeHelp).HandlerFunc(serveHelp)
	router.Get(routeSnippets).Handler(handler(serveBasicPageString("Snippets - Sourcegraph")))
	router.Get(routeSubscriptions).Handler(handler(serveBasicPageString("Subscriptions - Sourcegraph")))
	router.Get(routeProject).Handler(handler(serveBasicPageString("Projects - Sourcegraph")))

	router.Get(routeUserSettings).Handler(handler(serveBasicPageString("User settings - Sourcegraph")))
	router.Get(routeUserRedirect).Handler(handler(serveBasicPageString("User - Sourcegraph")))
	router.Get(routeUser).Handler(handler(serveBasicPage(func(c *Common, r *http.Request) string {
		return mux.Vars(r)["username"] + " - Sourcegraph"
	})))

	// Legacy redirects
	if envvar.SourcegraphDotComMode() {
		router.Get(routeLegacyLogin).Handler(staticRedirectHandler("/sign-in", http.StatusMovedPermanently))
		router.Get(routeLegacyCareers).Handler(staticRedirectHandler("https://about.sourcegraph.com/jobs", http.StatusMovedPermanently))
		router.Get(routeLegacyOldRouteDefLanding).Handler(http.HandlerFunc(serveOldRouteDefLanding))
		router.Get(routeLegacyDefRedirectToDefLanding).Handler(http.HandlerFunc(serveDefRedirectToDefLanding))
		router.Get(routeLegacyDefLanding).Handler(handler(serveDefLanding))
		router.Get(routeLegacyRepoLanding).Handler(handler(serveRepoLanding))
	}
	router.Get(routeLegacySearchQueries).Handler(staticRedirectHandler("/search/searches", http.StatusMovedPermanently))

	// search
	router.Get(routeSearch).Handler(handler(serveBasicPage(func(c *Common, r *http.Request) string {
		shortQuery := limitString(r.URL.Query().Get("q"), 25, true)
		if shortQuery == "" {
			return "Sourcegraph" // no query, on search homepage
		}
		// e.g. "myquery - Sourcegraph"
		return fmt.Sprintf("%s - Sourcegraph", shortQuery)
	})))

	// search badge
	router.Get(routeSearchBadge).Handler(searchBadgeHandler)

	// Saved searches
	router.Get(routeSearchSearches).Handler(handler(serveBasicPage(func(c *Common, r *http.Request) string {
		return "Saved searches - Sourcegraph"
	})))

	if envvar.SourcegraphDotComMode() {
		// about subdomain
		router.Get(routeAboutSubdomain).Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.URL.Scheme = aboutRedirectScheme
			r.URL.User = nil
			r.URL.Host = aboutRedirectHost
			r.URL.Path = "/" + aboutRedirects[mux.Vars(r)["Path"]]
			http.Redirect(w, r, r.URL.String(), http.StatusTemporaryRedirect)
		}))
	}

	// repo
	serveRepoHandler := handler(serveRepoOrBlob(routeRepo, func(c *Common, r *http.Request) string {
		// e.g. "gorilla/mux - Sourcegraph"
		return fmt.Sprintf("%s - Sourcegraph", repoShortName(c.Repo.Name))
	}))
	router.Get(routeRepo).Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Debug mode: register the __errorTest handler.
		if env.InsecureDev && r.URL.Path == "/__errorTest" {
			handler(serveErrorTest).ServeHTTP(w, r)
			return
		}

		if mockServeRepo != nil {
			mockServeRepo(w, r)
			return
		}
		serveRepoHandler.ServeHTTP(w, r)
	}))

	// tree
	router.Get(routeTree).Handler(handler(serveBasicPage(func(c *Common, r *http.Request) string {
		// e.g. "src - gorilla/mux - Sourcegraph"
		dirName := path.Base(mux.Vars(r)["Path"])
		return fmt.Sprintf("%s - %s - Sourcegraph", dirName, repoShortName(c.Repo.Name))
	})))

	// blob
	router.Get(routeBlob).Handler(handler(serveRepoOrBlob(routeBlob, func(c *Common, r *http.Request) string {
		// e.g. "mux.go - gorilla/mux - Sourcegraph"
		fileName := path.Base(mux.Vars(r)["Path"])
		return fmt.Sprintf("%s - %s - Sourcegraph", fileName, repoShortName(c.Repo.Name))
	})))

	// raw
	router.Get(routeRaw).Handler(handler(serveRaw))

	// All other routes that are not found.
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveError(w, r, errors.New("route not found"), http.StatusNotFound)
	})
}

// staticRedirectHandler returns an HTTP handler that redirects all requests to
// the specified url or relative path with the specified status code.
//
// The scheme, host, and path in the specified url override ones in the incoming
// request. For example:
//
// 	staticRedirectHandler("http://google.com") serving "https://sourcegraph.com/foobar?q=foo" -> "http://google.com/foobar?q=foo"
// 	staticRedirectHandler("/foo") serving "https://sourcegraph.com/bar?q=foo" -> "https://sourcegraph.com/foo?q=foo"
//
func staticRedirectHandler(u string, code int) http.Handler {
	target, err := url.Parse(u)
	if err != nil {
		// panic is OK here because staticRedirectHandler is called only inside
		// init / crash would be on server startup.
		panic(err)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if target.Scheme != "" {
			r.URL.Scheme = target.Scheme
		}
		if target.Host != "" {
			r.URL.Host = target.Host
		}
		if target.Path != "" {
			r.URL.Path = target.Path
		}
		http.Redirect(w, r, r.URL.String(), code)
	})
}

// limitString limits the given string to at most N characters, optionally
// adding an ellipsis (…) at the end.
func limitString(s string, n int, ellipsis bool) string {
	if len(s) < n {
		return s
	}
	if ellipsis {
		return s[:n-1] + "…"
	}
	return s[:n-1]
}

// handler wraps an HTTP handler that returns potential errors. If any error is
// returned, serveError is called.
//
// Clients that wish to return their own HTTP status code should use this from
// their handler:
//
// 	serveError(w, r, err, http.MyStatusCode)
//  return nil
//
func handler(f func(w http.ResponseWriter, r *http.Request) error) http.Handler {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				serveError(w, r, recoverError{recover: rec, stack: debug.Stack()}, http.StatusInternalServerError)
			}
		}()
		if err := f(w, r); err != nil {
			serveError(w, r, err, http.StatusInternalServerError)
		}
	})
	return trace.TraceRoute(gziphandler.GzipHandler(h))
}

type recoverError struct {
	recover interface{}
	stack   []byte
}

func (r recoverError) Error() string {
	return fmt.Sprintf("ui: recovered from panic: %v", r.recover)
}

// serveError serves the error template with the specified error message. It is
// assumed that the error message could accidentally contain sensitive data,
// and as such is only presented to the user in debug mode.
func serveError(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
	serveErrorNoDebug(w, r, err, statusCode, false, false)
}

// dangerouslyServeError is like serveError except it always shows the error to
// the user and as such, if it contains sensitive information, it can leak
// sensitive information.
//
// See https://github.com/sourcegraph/sourcegraph/issues/9453
func dangerouslyServeError(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
	serveErrorNoDebug(w, r, err, statusCode, false, true)
}

type pageError struct {
	StatusCode int    `json:"statusCode"`
	StatusText string `json:"statusText"`
	Error      string `json:"error"`
	ErrorID    string `json:"errorID"`
}

// serveErrorNoDebug should not be called by anyone except serveErrorTest.
func serveErrorNoDebug(w http.ResponseWriter, r *http.Request, err error, statusCode int, nodebug, forceServeError bool) {
	w.WriteHeader(statusCode)
	errorID := randstring.NewLen(6)

	// Determine span URl and log the error.
	var spanURL string
	if span := opentracing.SpanFromContext(r.Context()); span != nil {
		ext.Error.Set(span, true)
		span.SetTag("err", err)
		span.SetTag("error-id", errorID)
		spanURL = trace.SpanURL(span)
	}
	log15.Error("ui HTTP handler error response", "method", r.Method, "request_uri", r.URL.RequestURI(), "status_code", statusCode, "error", err, "error_id", errorID, "trace", spanURL)

	// In the case of recovering from a panic, we nicely include the stack
	// trace in the error that is shown on the page. Additionally, we log it
	// separately (since log15 prints the escaped sequence).
	if r, ok := err.(recoverError); ok {
		err = fmt.Errorf("ui: recovered from panic %v\n\n%s", r.recover, r.stack)
		log.Println(err)
	}

	var errorIfDebug string
	if forceServeError || (env.InsecureDev && !nodebug) {
		errorIfDebug = err.Error()
	}

	pageErrorContext := &pageError{
		StatusCode: statusCode,
		StatusText: http.StatusText(statusCode),
		Error:      errorIfDebug,
		ErrorID:    errorID,
	}

	// First try to render the error fancily: this relies on *Common
	// functionality that might always work (for example, if some services are
	// down rather than something that is primarily a user error).
	delete(mux.Vars(r), "Repo")
	var commonServeErr error
	title := fmt.Sprintf("%v %s - Sourcegraph", statusCode, http.StatusText(statusCode))
	common, commonErr := newCommon(w, r, title, func(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
		// Stub out serveError to newCommon so that it is not reentrant.
		commonServeErr = err
	})
	common.Error = pageErrorContext
	if commonErr == nil && commonServeErr == nil {
		if common == nil {
			return // request handled by newCommon
		}
		fancyErr := renderTemplate(w, "app.html", &struct {
			*Common
		}{
			Common: common,
		})
		if fancyErr != nil {
			log15.Error("ui: error while serving fancy error template", "error", fancyErr)
			// continue onto fallback below..
		} else {
			return
		}
	}

	// Fallback to ugly / reliable error template.
	stdErr := renderTemplate(w, "error.html", pageErrorContext)
	if stdErr != nil {
		log15.Error("ui: error while serving final error template", "error", stdErr)
	}
}

// serveErrorTest makes it easy to test styling/layout of the error template by
// visiting:
//
// 	http://localhost:3080/__errorTest?nodebug=true&error=theerror&status=500
//
// The `nodebug=true` parameter hides error messages (which is ALWAYS the case
// in production), `error` controls the error message text, and status controls
// the status code.
func serveErrorTest(w http.ResponseWriter, r *http.Request) error {
	if !env.InsecureDev {
		w.WriteHeader(http.StatusNotFound)
		return nil
	}
	q := r.URL.Query()
	nodebug := q.Get("nodebug") == "true"
	errorText := q.Get("error")
	statusCode, _ := strconv.Atoi(q.Get("status"))
	serveErrorNoDebug(w, r, errors.New(errorText), statusCode, nodebug, false)
	return nil
}

func mapKeys(m map[string]string) (keys []string) {
	keys = make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
