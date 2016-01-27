package internal

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"gopkg.in/inconshreveable/log15.v2"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
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
// grpc.Code(err) == codes.Unauthenticated).
//
// Currently it is set at init time by package authutil; this is
// necessary to avoid an import cycle.
var UnauthorizedErrorHandler ErrorHandler

// HandleError handles err, delegating to an error handler func
// previously registered with RegisterErrorHandler* if present, and
// otherwise displaying the standard error page.
func HandleError(resp http.ResponseWriter, req *http.Request, status int, err error) {
	origErr := err

	if errH, present := errorHandlers[reflect.TypeOf(err)]; present {
		err = errH(resp, req, err)
		if err == nil {
			return
		}
	} else if grpc.Code(err) == codes.Unauthenticated {
		log15.Debug("redirecting to login", "from", req.URL, "grpc_code", err)
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

	if status < 200 || status >= 500 {
		log15.Debug("app/internal.HandleError called with unsuccessful status code", "method", req.Method, "request_uri", req.URL.RequestURI(), "status_code", status, "error", err.Error())
	}

	// Handle potential panic during execution of error template, since it may call out to
	// code provided by an external platform app (e.g., via the GlobalApp.IconBadge callback).
	defer func() {
		if e := recover(); e != nil {
			log15.Error("panic during execution of error template", "error", e, "func_name", "HandleError", "tmpl_name", "error/error.html")
			err := fmt.Errorf("panic during execution of error template: %v", e)
			if !handlerutil.DebugMode(req) {
				err = errPublicFacingErrorMessage
			}
			http.Error(resp, err.Error(), http.StatusInternalServerError)
		}
	}()

	errHeader := http.Header{"cache-control": []string{"no-cache"}}
	err2 := tmpl.Exec(req, resp, "error/error.html", status, errHeader, &struct {
		StatusCode int
		Status     string
		Err        error
		tmpl.Common
	}{
		StatusCode: status,
		Status:     http.StatusText(status),
		Err:        err,
	})
	if err2 != nil {
		log15.Error("error during execution of error template", "error", err2)
		err := fmt.Errorf("error during execution of error template: %v", err2)
		if !handlerutil.DebugMode(req) {
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

If this issue persists, please post an issue in our tracker (https://src.sourcegraph.com/sourcegraph/.tracker/new)
with steps to reproduce and other useful context. Or, contact us by email (help@sourcegraph.com).`)
