package handlerutil

import (
	"context"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// HandlerWithErrorReturn wraps a http.HandlerFunc-like func that also
// returns an error.  If the error is nil, this wrapper is a no-op. If
// the error is non-nil, it attempts to determine the HTTP status code
// equivalent of the returned error (if non-nil) and set that as the
// HTTP status.
//
// Error must never panic. If it has to execute something that may panic
// (for example, call out into an external code), then it must use recover
// to catch potential panics. If Error panics, the panic will propagate upstream.
type HandlerWithErrorReturn struct {
	Handler func(http.ResponseWriter, *http.Request) error       // the underlying handler
	Error   func(http.ResponseWriter, *http.Request, int, error) // called to send an error response (e.g., an error page), it must not panic

	PretendPanic bool
}

func (h HandlerWithErrorReturn) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle when h.Handler panics.
	defer func() {
		if e := recover(); e != nil {
			// ErrAbortHandler is a sentinal error which is used to stop an
			// http handler but not report the error. In practice we have only
			// seen this used by httputil.ReverseProxy when the server goes
			// down.
			if e == http.ErrAbortHandler {
				return
			}

			log15.Error("panic in HandlerWithErrorReturn.Handler", "error", e)
			stack := make([]byte, 1024*1024)
			n := runtime.Stack(stack, false)
			stack = stack[:n]
			_, _ = io.WriteString(os.Stderr, "\nstack trace:\n")
			_, _ = os.Stderr.Write(stack)

			err := errors.Errorf("panic: %v\n\nstack trace:\n%s", e, stack)
			status := http.StatusInternalServerError
			h.Error(w, r, status, err) // No need to handle a possible panic in h.Error because it's required not to panic.
		}
	}()

	err := h.Handler(w, r)
	if err != nil {
		status := httpErrCode(r, err)
		h.Error(w, r, status, err)
	}
}

// httpErrCode maps an error to a status code. If the client canceled the
// request we return the non-standard "499 Client Closed Request" used by
// nginx.
func httpErrCode(r *http.Request, err error) int {
	// If we failed due to ErrCanceled, it may be due to the client closing
	// the connection. If that is the case, return 499. We do not just check
	// if the client closed the connection, in case we failed due to another
	// reason leading to the client closing the connection.
	if errors.Is(err, context.Canceled) && errors.Is(r.Context().Err(), context.Canceled) {
		return 499
	}
	return errcode.HTTP(err)
}
