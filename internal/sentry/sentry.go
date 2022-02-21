package sentry

import (
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"

	"github.com/cockroachdb/redact"
	"github.com/getsentry/sentry-go"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var sentryDebug, _ = strconv.ParseBool(env.Get("SENTRY_DEBUG", "false", "print debug messages for Sentry"))

// Init initializes the default Sentry client that uses SENTRY_DSN_BACKEND
// environment variable as the DSN. It then watches site configuration for any
// subsequent changes. SENTRY_DEBUG can be set as a boolean to print debug
// messages.
func Init(c conftypes.WatchableSiteConfig) {
	initClient := func(dsn string) error {
		if dsn == "" {
			return nil
		}

		err := sentry.Init(sentry.ClientOptions{
			Dsn:        dsn,
			Debug:      sentryDebug,
			ServerName: "", // Sentry client will gather the server name when leave empty
			Release:    version.Version(),
		})
		if err != nil {
			return err
		}

		sentry.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetTags(
				map[string]string{
					"service": filepath.Base(os.Args[0]),
				},
			)
			scope.SetLevel(sentry.LevelError)
		})
		return nil
	}

	err := initClient(os.Getenv("SENTRY_DSN_BACKEND"))
	if err != nil {
		log15.Error("sentry.initClient", "error", err)
	}

	go func() {
		c.Watch(func() {
			if c.SiteConfig().Log == nil {
				return
			}

			sentryConfig := c.SiteConfig().Log.Sentry
			if sentryConfig == nil {
				return
			}

			// Create a local variable to not mutate the original config object
			backendDSN := sentryConfig.BackendDSN

			// Fallback to default DSN if the backend DSN is not specified separately
			if backendDSN == "" {
				backendDSN = sentryConfig.Dsn
			}

			// An empty dsn value is ignored: not an error.
			if err := initClient(backendDSN); err != nil {
				log15.Error("sentry.initClient", "error", err)
			}
		})
	}()
}

func captureError(err error, level sentry.Level, tags map[string]string) {
	event, extraDetails := errors.BuildSentryReport(err)

	// Sentry uses the Type of the first exception as the issue title. By default,
	// "github.com/cockroachdb/errors" uses "<filename>:<lineno> (<functionname>)"
	// which is very sensitive to move up/down lines. Using the original error
	// string would be much more readable. We are also not losing location
	// information because that is also encoded in the stack trace.
	if len(event.Exception) > 0 {
		event.Exception[0].Type = errors.Cause(err).Error()
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetExtras(extraDetails)
		scope.SetTags(tags)
		scope.SetLevel(level)
		sentry.CaptureEvent(event)
	})
}

// CaptureError adds the given error to the default Sentry client delivery queue
// for reporting.
func CaptureError(err error, tags map[string]string) {
	captureError(err, sentry.LevelError, tags)
}

// CapturePanic does same thing as CaptureError, and adds additional tags to
// mark the report as "fatal" level.
func CapturePanic(err error, tags map[string]string) {
	captureError(err, sentry.LevelFatal, tags)
}

// Recovery handler to wrap the stdlib net/http Mux.
// Example:
//  mux := http.NewServeMux
//  ...
//	http.Handle("/", sentry.Recoverer(mux))
func Recoverer(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				err := errors.Errorf("handler panic: %v", redact.Safe(r))
				CapturePanic(err, nil)

				log15.Error("recovered from panic", "error", err)
				debug.PrintStack()

				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		handler.ServeHTTP(w, r)
	})
}
