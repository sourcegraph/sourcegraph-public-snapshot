package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/justinas/nosurf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/conf"
	httpapiauth "sourcegraph.com/sourcegraph/sourcegraph/httpapi/auth"
	apirouter "sourcegraph.com/sourcegraph/sourcegraph/httpapi/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/csp"
	"sourcegraph.com/sourcegraph/sourcegraph/util/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/util/eventsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
)

// NewHandler returns a new API handler that uses the provided API
// router, which must have been created by httpapi/router.New, or
// creates a new one if nil.
func NewHandler(m *mux.Router) http.Handler {
	if m == nil {
		m = apirouter.New(nil)
	}
	m.StrictSlash(true)

	// SECURITY NOTE: The HTTP API should not accept cookies as
	// authentication. Doing so would open it up to CSRF
	// attacks.
	var mw []handlerutil.Middleware
	mw = append(mw, httpapiauth.PasswordMiddleware, httpapiauth.OAuth2AccessTokenMiddleware)
	mw = append(mw, eventsutil.AgentMiddleware)

	if conf.GetenvBool("SG_USE_CSP") {
		// Set the CSP handler. Determine the report URI by seeing what
		// path prefix m currently has (if it was just /.csp-report, then
		// it'd never match, since this handler is usually mounted at
		// /.api/).
		reportURI, err := m.Path(cspConfig.Policy.ReportURI).URLPath()
		if err != nil {
			panic(err.Error())
		}
		cspConfig.Policy.ReportURI = reportURI.String()
		cspHandler := csp.NewHandler(cspConfig)
		mw = append(mw, cspHandler.Middleware)
	}

	// Set handlers for the installed routes.
	m.Get(apirouter.Signup).Handler(nosurf.New(grpcErrorHandler(serveSignup)))
	m.Get(apirouter.Login).Handler(nosurf.New(grpcErrorHandler(serveLogin)))
	m.Get(apirouter.Logout).Handler(nosurf.New(handler(serveLogout)))
	m.Get(apirouter.ForgotPassword).Handler(nosurf.New(grpcErrorHandler(serveForgotPassword)))
	m.Get(apirouter.ResetPassword).Handler(nosurf.New(grpcErrorHandler(servePasswordReset)))

	m.Get(apirouter.Annotations).Handler(handler(serveAnnotations))
	m.Get(apirouter.AuthInfo).Handler(handler(serveAuthInfo))
	m.Get(apirouter.BlackHole).Handler(handler(serveBlackHole))
	m.Get(apirouter.Builds).Handler(handler(serveBuilds))
	m.Get(apirouter.BuildTaskLog).Handler(handler(serveBuildTaskLog))
	m.Get(apirouter.Coverage).Handler(handler(serveCoverage))
	m.Get(apirouter.Def).Handler(handler(serveDef))
	m.Get(apirouter.DefAuthors).Handler(handler(serveDefAuthors))
	m.Get(apirouter.DefRefs).Handler(handler(serveDefRefs))
	m.Get(apirouter.DefRefLocations).Handler(handler(serveDefRefLocations))
	m.Get(apirouter.Defs).Handler(handler(serveDefs))
	m.Get(apirouter.GlobalSearch).Handler(handler(serveGlobalSearch))
	m.Get(apirouter.Repo).Handler(handler(serveRepo))
	m.Get(apirouter.RepoResolve).Handler(handler(serveRepoResolve))
	m.Get(apirouter.RepoInventory).Handler(handler(serveRepoInventory))
	m.Get(apirouter.RepoCreate).Handler(handler(serveRepoCreate))
	m.Get(apirouter.RepoBranches).Handler(handler(serveRepoBranches))
	m.Get(apirouter.RepoCommits).Handler(handler(serveRepoCommits))
	m.Get(apirouter.RepoTree).Handler(handler(serveRepoTree))
	m.Get(apirouter.RepoTreeList).Handler(handler(serveRepoTreeList))
	m.Get(apirouter.RepoTreeSearch).Handler(handler(serveRepoTreeSearch))
	m.Get(apirouter.RepoBuild).Handler(handler(serveRepoBuild))
	m.Get(apirouter.RepoBuilds).Handler(handler(serveRepoBuilds))
	m.Get(apirouter.RepoBuildTasks).Handler(handler(serveBuildTasks))
	m.Get(apirouter.RepoBuildsCreate).Handler(handler(serveRepoBuildsCreate))
	m.Get(apirouter.RepoRefresh).Handler(handler(serveRepoRefresh))
	m.Get(apirouter.RepoResolveRev).Handler(handler(serveRepoResolveRev))
	m.Get(apirouter.RepoTags).Handler(handler(serveRepoTags))
	m.Get(apirouter.Repos).Handler(handler(serveRepos))
	m.Get(apirouter.RemoteRepos).Handler(handler(serveRemoteRepos))
	m.Get(apirouter.SrclibImport).Handler(handler(serveSrclibImport))
	m.Get(apirouter.SrclibDataVer).Handler(handler(serveSrclibDataVersion))
	m.Get(apirouter.User).Handler(handler(serveUser))
	m.Get(apirouter.UserEmails).Handler(handler(serveUserEmails))
	m.Get(apirouter.InternalAppdashRecordSpan).Handler(handler(serveInternalAppdashRecordSpan))

	m.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("API no route: %s %s from %s", r.Method, r.URL, r.Referer())
		http.Error(w, "no route", http.StatusNotFound)
	})

	return handlerutil.WithMiddleware(m, mw...)
}

// handler is a wrapper func for API handlers.
func handler(h func(http.ResponseWriter, *http.Request) error) http.Handler {
	return handlerutil.HandlerWithErrorReturn{
		Handler: func(w http.ResponseWriter, r *http.Request) error {
			w.Header().Set("Content-Type", "application/json")
			return h(w, r)
		},
		Error: handleError,
	}
}

// grpcErrorHandler is a wrapper func for API handlers that gives special
// treatment to gRPC errors using handleErrorWithGRPC
func grpcErrorHandler(h func(http.ResponseWriter, *http.Request) error) http.Handler {
	return handlerutil.HandlerWithErrorReturn{
		Handler: h,
		Error:   handleErrorWithGRPC,
	}
}

// cspConfig is the Content Security Policy config for API handlers.
var cspConfig = csp.Config{
	// Strict because API responses should never be treated as page
	// content.
	Policy: &csp.Policy{DefaultSrc: []string{"'none'"}, ReportURI: "/.csp-report"},
}

var schemaDecoder = schema.NewDecoder()

func init() {
	schemaDecoder.IgnoreUnknownKeys(true)

	// Register a converter for unix timestamp strings -> time.Time values
	// (needed for Appdash PageLoadEvent type).
	schemaDecoder.RegisterConverter(time.Time{}, func(s string) reflect.Value {
		ms, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return reflect.ValueOf(err)
		}
		return reflect.ValueOf(time.Unix(0, ms*int64(time.Millisecond)))
	})
}

func handleError(w http.ResponseWriter, r *http.Request, status int, err error) {
	// Handle custom errors
	if ee, ok := err.(*handlerutil.URLMovedError); ok {
		err := handlerutil.RedirectToNewRepoURI(w, r, ee.NewURL)
		if err != nil {
			log15.Error("error redirecting to new URI", "err", err, "new_url", ee.NewURL)
		}
		return
	}

	// Never cache error responses.
	w.Header().Set("cache-control", "no-cache, max-age=0")

	errBody := err.Error()

	var displayErrBody string
	if handlerutil.DebugMode(r) {
		// Only display error message to admins or locally, since it
		// can contain sensitive info (like API keys in net/http error
		// messages).
		displayErrBody = string(errBody)
	}
	http.Error(w, displayErrBody, status)
	if status < 200 || status >= 500 {
		log15.Error("API HTTP handler error response", "method", r.Method, "request_uri", r.URL.RequestURI(), "status_code", status, "error", err)
	}
}

// handleErrorWithGRPC is the error handler put on user-form APIs like login and signup. It
// packages gRPC errors so the errors can be parsed in the frontend and displayed to users.
func handleErrorWithGRPC(w http.ResponseWriter, r *http.Request, status int, err error) {
	// Handle custom errors
	if ee, ok := err.(*handlerutil.URLMovedError); ok {
		err := handlerutil.RedirectToNewRepoURI(w, r, ee.NewURL)
		if err != nil {
			log15.Error("error redirecting to new URI", "err", err, "new_url", ee.NewURL)
		}
		return
	}

	// Never cache error responses.
	w.Header().Set("cache-control", "no-cache, max-age=0")

	if code := grpc.Code(err); code != codes.Unknown {

		type errorMessage struct {
			Code    codes.Code `json:"code"`
			Message string     `json:"message"`
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(errcode.HTTP(err))
		json.NewEncoder(w).Encode(errorMessage{Code: code, Message: grpc.ErrorDesc(err)})
		return
	}

	errBody := err.Error()

	var displayErrBody string
	if handlerutil.DebugMode(r) {
		// Only display error message to admins or locally, since it
		// can contain sensitive info (like API keys in net/http error
		// messages).
		displayErrBody = string(errBody)
	}
	http.Error(w, displayErrBody, status)
	if status < 200 || status >= 500 {
		log15.Error("API HTTP handler error response", "method", r.Method, "request_uri", r.URL.RequestURI(), "status_code", status, "error", err)
	}
}
