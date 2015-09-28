package internal

import (
	"fmt"
	"log"
	"net/http"
	"reflect"

	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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
		log.Printf("Redirecting to login from %s (got Unauthenticated gRPC code: %v).", req.URL, err)
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
		log.Printf("%s %s %d: %s", req.Method, req.URL.RequestURI(), status, err.Error())
	}

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
		log.Printf("Error during execution of error template: %s.", err2)
	}
}
