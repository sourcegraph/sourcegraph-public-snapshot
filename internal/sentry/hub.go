package sentry

import (
	"net/http"
	"runtime/debug"

	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/sentry-go"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/version"
)

type Hub struct {
	*sentry.Hub
}

func NewWithDsn(dsn string) (*Hub, error) {
	c, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:        dsn,
		Debug:      sentryDebug,
		ServerName: "", // Sentry client will gather the server name when leave empty
		Release:    version.Version(),
	})
	if err != nil {
		return nil, err
	}

	h := &Hub{
		Hub: sentry.NewHub(c, sentry.NewScope()),
	}

	return h, nil
}

func (h *Hub) captureError(err error, level sentry.Level, tags map[string]string) {
	event, extraDetails := errors.BuildSentryReport(err)

	// Sentry uses the Type of the first exception as the issue title. By default,
	// "github.com/cockroachdb/errors" uses "<filename>:<lineno> (<functionname>)"
	// which is very sensitive to move up/down lines. Using the original error
	// string would be much more readable. We are also not losing location
	// information because that is also encoded in the stack trace.
	if len(event.Exception) > 0 {
		event.Exception[0].Type = errors.Cause(err).Error()
	}

	h.WithScope(func(scope *sentry.Scope) {
		scope.SetExtras(extraDetails)
		scope.SetTags(tags)
		scope.SetLevel(level)
		h.CaptureEvent(event)
	})
}

// CaptureError adds the given error to the default Sentry client delivery queue
// for reporting.
func (h *Hub) CaptureError(err error, tags map[string]string) {
	h.captureError(err, sentry.LevelError, tags)
}

// CapturePanic does same thing as CaptureError, and adds additional tags to
// mark the report as "fatal" level.
func (h *Hub) CapturePanic(err error, tags map[string]string) {
	h.captureError(err, sentry.LevelFatal, tags)
}

func (h *Hub) Recoverer(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				err := errors.Errorf("handler panic: %v", errors.Safe(r))
				h.CapturePanic(err, nil)

				log15.Error("recovered from panic", "error", err)
				debug.PrintStack()

				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		handler.ServeHTTP(w, r)
	})
}
