package ui2

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/NYTimes/gziphandler"
	"github.com/gorilla/mux"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/randstring"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
)

const (
	routeHome       = "home"
	routeRepoOrRoot = "repo-or-root"
	routeTree       = "tree"
	routeBlob       = "blob"

	aboutRedirectURL = "https://about.sourcegraph.com/" // must end with trailing slash
)

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

	// home
	r.Path("/").Methods("GET").Name(routeHome)

	// repo-or-root
	repoRevPath := "/" + routevar.Repo + routevar.RepoRevSuffix
	r.Path(repoRevPath).Methods("GET").Name(routeRepoOrRoot)

	// tree
	repoRev := r.PathPrefix(repoRevPath + "/" + routevar.RepoPathDelim).Subrouter()
	repoRev.Path("/tree{Path:.*}").Methods("GET").Name(routeTree)

	// blob
	repoRev.Path("/blob{Path:.*}").Methods("GET").Name(routeBlob)
	return r
}

func init() {
	router = newRouter()
	router.Get(routeHome).Handler(handler(serveHome))
	serveRepoHandler := handler(serveRepo)
	router.Get(routeRepoOrRoot).Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Debug mode: register the __errorTest handler.
		if handlerutil.DebugMode && r.URL.Path == "/__errorTest" {
			handler(serveErrorTest).ServeHTTP(w, r)
			return
		}

		_, err := handlerutil.GetRepo(r.Context(), mux.Vars(r))
		if legacyerr.ErrCode(err) == legacyerr.NotFound {
			// Repository not found, so redirect the request to about.sourcegraph.com
			// instead. This case does NOT trigger for repos that are cloning.
			http.Redirect(w, r, aboutRedirectURL+mux.Vars(r)["Repo"], http.StatusTemporaryRedirect)
			return
		}
		if mockServeRepo != nil {
			mockServeRepo(w, r)
			return
		}
		serveRepoHandler.ServeHTTP(w, r)
	}))
	router.Get(routeTree).Handler(handler(serveTree))
	router.Get(routeBlob).Handler(handler(serveBlob))
}

// urlTo returns a url to the named route.
func urlTo(routeName string, params ...string) *url.URL {
	route := Router().Get(routeName)
	if route == nil {
		panic(fmt.Sprintf("no such route %q (params: %v)", routeName, params))
	}
	u, err := route.URL(params...)
	if err != nil {
		panic(fmt.Sprintf("failed to compose URL to %q (params: %v)", routeName, params))
	}
	return u
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
	if handlerutil.DebugMode {
		errorIfDebug = err.Error()
	}
	err2 := renderTemplate(w, "error.html", &struct {
		StatusCode                 int
		StatusText, Error, ErrorID string
	}{
		StatusCode: statusCode,
		StatusText: http.StatusText(statusCode),
		Error:      errorIfDebug,
		ErrorID:    errorID,
	})
	if err2 != nil {
		log15.Error("ui: error while serving error template", "error", err2)
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
