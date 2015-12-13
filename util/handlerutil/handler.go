package handlerutil

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"sync"

	"github.com/gorilla/schema"
	"github.com/resonancelabs/go-pub/instrument"
	"github.com/resonancelabs/go-pub/instrument/httpwrapper"

	"github.com/rogpeppe/rog-go/parallel"

	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
	"src.sourcegraph.com/sourcegraph/util/traceutil"
)

var (
	schemaDecoder = schema.NewDecoder()
	once          sync.Once
)

func init() {
	once.Do(func() {
		schemaDecoder.IgnoreUnknownKeys(true)
	})
}

// Handler is the outermost http.Handler wrapper per route.
func Handler(h HandlerWithErrorReturn) http.Handler {
	return WithMiddleware(h, httpwrapper.MakeMiddleware(httpwrapperConfig))
}

var httpwrapperConfig = &httpwrapper.ServerConfig{
	WithActiveSpanFunc: func(r *http.Request, span instrument.ActiveSpan) {
		span.SetName(fmt.Sprintf("http/%s", httpctx.RouteName(r)))
		span.Log(instrument.EventName("cr/span_attributes").Payload(map[string]string{
			"route_path": r.URL.Path,
		}))

		spanID := traceutil.SpanID(r)
		span.Log(instrument.EventName("appdash_span_id").Payload(spanID))
		span.AddTraceJoinId("appdash_trace_id", spanID.Trace)
	},
}

// HandlerWithErrorReturn wraps a http.HandlerFunc-like func that also
// returns an error.  If the error is nil, this wrapper is a no-op. If
// the error is non-nil, it attempts to determine the HTTP status code
// equivalent of the returned error (if non-nil) and set that as the
// HTTP status. If a non-nil error is returned
type HandlerWithErrorReturn struct {
	Handler func(http.ResponseWriter, *http.Request) error       // the underlying handler
	Error   func(http.ResponseWriter, *http.Request, int, error) // called to send an error response (e.g., an error page)
}

func (h HandlerWithErrorReturn) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error

	defer func() {
		if rv := recover(); rv != nil {
			err = fmt.Errorf("panic: %v", rv)
			log.Println(string(debug.Stack()))
			status := http.StatusInternalServerError
			reportError(r, status, err, true)
			h.Error(w, r, status, err)
		}
	}()

	err = collapseMultipleErrors(h.Handler(w, r))
	if err != nil {
		status := errcode.HTTP(err)
		reportError(r, status, err, false)
		h.Error(w, r, status, err)
	}
}

// collapseMultipleErrors returns the first err if err is a
// parallel.Errors list of length 1. Otherwise it returns err
// unchanged. This lets us return the proper HTTP status code for
// single errors.
func collapseMultipleErrors(err error) error {
	if errs, ok := err.(parallel.Errors); ok && len(errs) == 1 {
		return errs[0]
	}
	return err
}
