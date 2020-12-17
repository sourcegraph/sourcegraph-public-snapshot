package sentry

import (
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"

	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/sentry-go"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

var sentryDebug, _ = strconv.ParseBool(env.Get("SENTRY_DEBUG", "false", "print debug messages for Sentry"))

// Init initializes the default Sentry client that uses SENTRY_DSN_BACKEND
// environment variable as the DSN. It then watches site configuration for any
// subsequent changes. SENTRY_DEBUG can be set as a boolean to print debug
// messages.
func Init() {
	initClient := func(dsn string) error {
		if dsn == "" {
			return nil
		}

		err := sentry.Init(sentry.ClientOptions{
			Dsn:              dsn,
			Debug:            sentryDebug,
			AttachStacktrace: false,
			SampleRate:       0,
			IgnoreErrors:     nil,
			BeforeBreadcrumb: nil,
			Integrations:     nil,
			DebugWriter:      nil,
			Transport:        nil,
			ServerName:       "",
			Release:          version.Version(),
			Dist:             "",
			Environment:      "",
			MaxBreadcrumbs:   0,
			HTTPClient:       nil,
			HTTPTransport:    nil,
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
		conf.Watch(func() {
			if conf.Get().Log == nil {
				return
			}

			if conf.Get().Log.Sentry == nil {
				return
			}

			// An empty dsn value is ignored: not an error.
			if err := initClient(conf.Get().Log.Sentry.Dsn); err != nil {
				log15.Error("sentry.initClient", "error", err)
			}
		})
	}()
}

// CaptureError adds the given error to the default Sentry client delivery queue
// for reporting.
func CaptureError(err error, tags map[string]string) {
	event, extraDetails := errors.BuildSentryReport(err)

	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetExtras(extraDetails)
		scope.SetTags(tags)
		sentry.CaptureEvent(event)
	})
}

// CapturePanic does same thing as CaptureError, and adds additional tags to
// mark the report as "fatal" level.
func CapturePanic(err error, tags map[string]string) {
	event, extraDetails := errors.BuildSentryReport(err)

	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetExtras(extraDetails)
		scope.SetTags(tags)
		scope.SetLevel(sentry.LevelFatal)
		sentry.CaptureEvent(event)
	})
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
				debug.PrintStack()

				err := errors.Errorf("%v", r)
				CapturePanic(err, nil)

				log15.Error("recovered from panic", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		handler.ServeHTTP(w, r)
	})
}
