package ui2

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"runtime/debug"
	"strconv"
	"strings"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/NYTimes/gziphandler"
	"github.com/gorilla/mux"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/randstring"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
)

const (
	routeHome                 = "home"
	routeSearch               = "search"
	routeComment              = "comment"
	routeRepo                 = "repo"
	routeTree                 = "tree"
	routeBlob                 = "blob"
	routeSignIn               = "sign-in"
	routeSettings             = "settings"
	routeSettingsAcceptInvite = "settings.accept-invite"
	routeSettingsEditorAuth   = "settings.editor-auth"
	routeSettingsTeamsNew     = "settings.teams.new"
	routeSettingsTeamsTeam    = "settings.teams.team"
	routePasswordReset        = "password-reset"

	routeAboutSubdomain = "about-subdomain"
	aboutRedirectScheme = "https"
	aboutRedirectHost   = "about.sourcegraph.com"

	// Legacy redirects
	routeLegacyLogin                   = "login"
	routeLegacyCareers                 = "careers"
	routeLegacyDefLanding              = "page.def.landing"
	routeLegacyOldRouteDefLanding      = "page.def.landing.old"
	routeLegacyRepoLanding             = "page.repo.landing"
	routeLegacyDefRedirectToDefLanding = "page.def.redirect"
	routeLegacyAcceptInvite            = "legacy.accept-invite"
	routeLegacySettingsTeam            = "legacy.settings.team"
	routeLegacyEditorAuth              = "legacy.editor-auth"
)

// aboutPaths is a list of paths that should redirect from sourcegraph.com/$PATH
// to about.sourcegraph.com/$PATH.
var aboutPaths = []string{
	"about",
	"plan",
	"contact",
	"docs",
	"enterprise",
	"pricing",
	"privacy",
	"security",
	"terms",
	"jobs",
	"beta",
}

// Router returns the router that serves pages for our web app.
func Router() *mux.Router {
	return router
}

var (
	router        *mux.Router
	mockServeRepo func(w http.ResponseWriter, r *http.Request)
)

func newRouter() *mux.Router {
	r := mux.NewRouter()
	r.StrictSlash(true)

	// Top-level routes.
	r.Path("/").Methods("GET").Name(routeHome)
	r.Path("/search").Methods("GET").Name(routeSearch)
	r.Path("/c/{ULID}").Methods("GET").Name(routeComment)
	r.Path("/sign-in").Methods("GET").Name(routeSignIn)
	r.Path("/settings").Methods("GET").Name(routeSettings)
	r.Path("/settings/accept-invite").Methods("GET").Name(routeSettingsAcceptInvite)
	r.Path("/settings/editor-auth").Methods("GET").Name(routeSettingsEditorAuth)
	r.Path("/settings/teams/new").Methods("GET").Name(routeSettingsTeamsNew)
	r.Path("/settings/teams/{team}").Methods("GET").Name(routeSettingsTeamsTeam)
	r.Path("/password-reset").Methods("GET").Name(routePasswordReset)
	r.Path("/{Path:(?:" + strings.Join(aboutPaths, "|") + ")}").Methods("GET").Name(routeAboutSubdomain)

	// Legacy redirects
	r.Path("/login").Methods("GET").Name(routeLegacyLogin)
	r.Path("/careers").Methods("GET").Name(routeLegacyCareers)
	r.Path("/accept-invite").Methods("GET").Name(routeLegacyAcceptInvite)
	r.Path("/settings/team/{team}").Methods("GET").Name(routeLegacySettingsTeam)
	r.Path("/editor-auth").Methods("GET").Name(routeLegacyEditorAuth)

	// repo
	repoRevPath := "/" + routevar.Repo + routevar.RepoRevSuffix
	r.Path(repoRevPath).Methods("GET").Name(routeRepo)

	// tree
	repoRev := r.PathPrefix(repoRevPath + "/" + routevar.RepoPathDelim).Subrouter()
	repoRev.Path("/tree{Path:.*}").Methods("GET").Name(routeTree)

	// blob
	repoRev.Path("/blob{Path:.*}").Methods("GET").Name(routeBlob)

	// legacy redirects
	repo := r.PathPrefix(repoRevPath + "/" + routevar.RepoPathDelim).Subrouter()
	repo.Path("/info").Methods("GET").Name(routeLegacyRepoLanding)
	repoRev.Path("/{dummy:def|refs}/" + routevar.Def).Methods("GET").Name(routeLegacyDefRedirectToDefLanding)
	repoRev.Path("/info/" + routevar.Def).Methods("GET").Name(routeLegacyDefLanding)
	repoRev.Path("/land/" + routevar.Def).Methods("GET").Name(routeLegacyOldRouteDefLanding)
	return r
}

func init() {
	// basic pages with static titles
	router = newRouter()
	router.Get(routeHome).Handler(handler(serveHome))
	router.Get(routeSignIn).Handler(handler(serveBasicPageString("Sign in or sign up - Sourcegraph")))
	router.Get(routeSettings).Handler(handler(serveBasicPageString("Profile - Sourcegraph")))
	router.Get(routeSettingsAcceptInvite).Handler(handler(serveBasicPageWithEmailVerification(func(c *Common, r *http.Request) string {
		return "Accept invite - Sourcegraph"
	})))
	router.Get(routeSettingsEditorAuth).Handler(handler(serveEditorAuthWithEditorBetaRegistration))
	router.Get(routeSettingsTeamsNew).Handler(handler(serveBasicPageString("New team - Sourcegraph")))
	router.Get(routeSettingsTeamsTeam).Handler(handler(serveBasicPage(func(c *Common, r *http.Request) string {
		// e.g. "myteam - Sourcegraph"
		return fmt.Sprintf("%s - Sourcegraph", mux.Vars(r)["team"])
	})))
	router.Get(routePasswordReset).Handler(handler(serveBasicPageString("Reset password - Sourcegraph")))

	// Legacy redirects
	if !envvar.DeploymentOnPrem() {
		router.Get(routeLegacyLogin).Handler(staticRedirectHandler("/sign-in", http.StatusMovedPermanently))
		router.Get(routeLegacyCareers).Handler(staticRedirectHandler("https://about.sourcegraph.com/jobs", http.StatusMovedPermanently))
		router.Get(routeLegacyOldRouteDefLanding).Handler(http.HandlerFunc(serveOldRouteDefLanding))
		router.Get(routeLegacyDefRedirectToDefLanding).Handler(http.HandlerFunc(serveDefRedirectToDefLanding))
		router.Get(routeLegacyDefLanding).Handler(handler(serveDefLanding))
		router.Get(routeLegacyRepoLanding).Handler(handler(serveRepoLanding))
		router.Get(routeLegacyAcceptInvite).Handler(staticRedirectHandler("/settings/accept-invite", http.StatusMovedPermanently))
		router.Get(routeLegacySettingsTeam).Handler(handler(serveLegacySettingsTeam))
		router.Get(routeLegacyEditorAuth).Handler(staticRedirectHandler("/settings/editor-auth", http.StatusMovedPermanently))
	}

	// search
	router.Get(routeSearch).Handler(handler(serveBasicPage(func(c *Common, r *http.Request) string {
		shortQuery := limitString(r.URL.Query().Get("q"), 25, true)
		if shortQuery == "" {
			return "Sourcegraph" // no query, on search homepage
		}
		// e.g. "myquery - Sourcegraph"
		return fmt.Sprintf("%s - Sourcegraph", shortQuery)
	})))

	if !envvar.DeploymentOnPrem() {
		// about subdomain
		router.Get(routeAboutSubdomain).Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.URL.Scheme = aboutRedirectScheme
			r.URL.User = nil
			r.URL.Host = aboutRedirectHost
			r.URL.Path = mux.Vars(r)["Path"]
			http.Redirect(w, r, r.URL.String(), http.StatusTemporaryRedirect)
		}))
	}

	// comment
	router.Get(routeComment).Handler(handler(serveComment))

	// repo
	serveRepoHandler := handler(serveRepoOrBlob(routeRepo, func(c *Common, r *http.Request) string {
		// e.g. "gorilla/mux - Sourcegraph"
		return fmt.Sprintf("%s - Sourcegraph", repoShortName(c.Repo.URI))
	}))
	router.Get(routeRepo).Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Debug mode: register the __errorTest handler.
		if handlerutil.DebugMode && r.URL.Path == "/__errorTest" {
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
		return fmt.Sprintf("%s - %s - Sourcegraph", dirName, repoShortName(c.Repo.URI))
	})))

	// blob
	router.Get(routeBlob).Handler(handler(serveRepoOrBlob(routeBlob, func(c *Common, r *http.Request) string {
		// e.g. "mux.go - gorilla/mux - Sourcegraph"
		fileName := path.Base(mux.Vars(r)["Path"])
		return fmt.Sprintf("%s - %s - Sourcegraph", fileName, repoShortName(c.Repo.URI))
	})))

	// All other routes that are not found.
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveError(w, r, errors.New("route not found"), http.StatusNotFound)
		return
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
	return traceutil.TraceRoute(gziphandler.GzipHandler(h))
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
	serveErrorNoDebug(w, r, err, statusCode, false)
}

type pageError struct {
	StatusCode int    `json:"statusCode"`
	StatusText string `json:"statusText"`
	Error      string `json:"error"`
	ErrorID    string `json:"errorID"`
}

// serveErrorNoDebug should not be called by anyone except serveErrorTest.
func serveErrorNoDebug(w http.ResponseWriter, r *http.Request, err error, statusCode int, nodebug bool) {
	w.WriteHeader(statusCode)
	errorID := randstring.NewLen(6)

	// Determine span URl and log the error.
	var spanURL string
	if span := opentracing.SpanFromContext(r.Context()); span != nil {
		ext.Error.Set(span, true)
		span.SetTag("err", err)
		span.SetTag("error-id", errorID)
		spanURL = traceutil.SpanURL(span)
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
	if handlerutil.DebugMode && !nodebug {
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
	if !handlerutil.DebugMode {
		w.WriteHeader(http.StatusNotFound)
		return nil
	}
	q := r.URL.Query()
	nodebug := q.Get("nodebug") == "true"
	errorText := q.Get("error")
	statusCode, _ := strconv.Atoi(q.Get("status"))
	serveErrorNoDebug(w, r, errors.New(errorText), statusCode, nodebug)
	return nil
}
