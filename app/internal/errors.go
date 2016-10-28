package internal

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/randstring"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
)

// ErrorHandler is a func that renders a custom error page for the
// given error (which was returned by an app handler).
type ErrorHandler func(http.ResponseWriter, *http.Request, error) error

var errorHandlers = map[reflect.Type]ErrorHandler{}

// RegisterErrorHandlerForType registers an error handler that
// HandleError delegates to, for all errors (returned by app handlers)
// of the same type as errorVal.
//
// For example, to delegate to handleFooError for all errors returned
// by app handlers of type "*foo", call
// RegisterErrorHandlerForType(&foo{}, handleFooError).
//
// These error handlers are matched BEFORE those registered with
// RegisterErrorHandler are matched.
//
// It should only be called at init time. It panics if there is
// already a handler registered for the type.
func RegisterErrorHandlerForType(errorVal interface{}, handler ErrorHandler) {
	t := reflect.TypeOf(errorVal)
	if _, present := errorHandlers[t]; present {
		return
	}
	errorHandlers[t] = handler
}

// UnauthorizedErrorHandler is the error handler that is called when
// an app handler returns an "unauthenticated" error (i.e.,
// legacyerr.ErrCode(err) == legacyerr.Unauthenticated).
//
// Currently it is set at init time by package authutil; this is
// necessary to avoid an import cycle.
var UnauthorizedErrorHandler ErrorHandler

// HandleError handles err, delegating to an error handler func
// previously registered with RegisterErrorHandler* if present, and
// otherwise displaying the standard error page.
func HandleError(resp http.ResponseWriter, req *http.Request, status int, err error) {
	origErr := err
	errorID := randstring.NewLen(6)

	if errH, present := errorHandlers[reflect.TypeOf(err)]; present {
		err = errH(resp, req, err)
		if err == nil {
			return
		}
	} else if UnauthorizedErrorHandler != nil && (legacyerr.ErrCode(err) == legacyerr.Unauthenticated || status == http.StatusUnauthorized) {
		log15.Debug("redirecting to login", "from", req.URL, "error", err, "error_id", errorID)
		err = UnauthorizedErrorHandler(resp, req, err)
		if err == nil {
			return
		}
	} else {
		// err is the error that occurred during an error handler, but
		// this if-branch is when there is no error handler, so set it
		// to nil.
		err = nil
	}

	if err == nil {
		err = origErr
	} else {
		// An error occurred during execution of the error handler.
		err = fmt.Errorf("during execution of error handler: %s (original error: %s)", err, origErr)
	}

	if status < 200 || status >= 400 {
		var spanURL string
		if span := opentracing.SpanFromContext(req.Context()); span != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err)
			spanURL = traceutil.SpanURL(span)
		}
		log15.Error("App HTTP handler error response", "method", req.Method, "request_uri", req.URL.RequestURI(), "status_code", status, "error", err, "error_id", errorID, "trace", spanURL)
	}

	// Handle panic during execution of error template.
	defer func() {
		if e := recover(); e != nil {
			log15.Error("panic during execution of error template", "error", e, "error_id", errorID, "func_name", "HandleError", "tmpl_name", "error/error.html")
			err := fmt.Errorf("panic during execution of error template (error id %v): %v", errorID, e)
			if !handlerutil.DebugMode {
				err = errPublicFacingErrorMessage
			}
			http.Error(resp, err.Error(), http.StatusInternalServerError)
		}
	}()

	// Display internal grpc error descriptions with full text (so it's not escaped).
	if legacyerr.ErrCode(err) == legacyerr.Internal {
		err = fmt.Errorf("internal error:\n\n%s", legacyerr.ErrorDesc(err))
	}

	errHeader := http.Header{"cache-control": []string{"no-cache"}}
	err2 := tmpl.Exec(req, resp, "error/error.html", status, errHeader, &struct {
		Meta       map[string]interface{} // placeholder, to make tmpl.Exec() happy
		StatusCode int
		Status     string
		Err        error
		tmpl.Common
	}{
		Meta:       make(map[string]interface{}),
		StatusCode: status,
		Status:     http.StatusText(status),
		Err:        err,
		Common:     tmpl.Common{ErrorID: errorID},
	})
	if err2 != nil {
		log15.Error("error during execution of error template", "error", err2, "error_id", errorID)
		err := fmt.Errorf("error during execution of error template (error id %v): %v", errorID, err2)
		if !handlerutil.DebugMode {
			err = errPublicFacingErrorMessage
		}
		http.Error(resp, err.Error(), http.StatusInternalServerError)
	}
}

// errPublicFacingErrorMessage is the public facing error message to display
// when not in debug mode (to hide potentially sensitive information in original
// error message). Normally, an HTML template is used to do this, but if rendering that fails,
// this is the plain-text backup.
var errPublicFacingErrorMessage = errors.New(`Sorry, there’s been a problem.

If this issue persists, please email us at support@sourcegraph.com.`)
