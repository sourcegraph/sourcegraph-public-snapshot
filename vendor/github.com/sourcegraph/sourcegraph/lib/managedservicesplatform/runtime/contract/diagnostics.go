package contract

import (
	"context"
	"crypto/subtle"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/trace"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime/internal/opentelemetry"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type diagnosticsContract struct {
	// DiagnosticsSecret can be used to authenticate diagnostics requests.
	//
	// 🚨 SECURITY: Do NOT use to authenticate sensitive data access. This should
	// only be used for governing access to diagnostic information (and should
	// still be treated with great care like any other application secrets).
	DiagnosticsSecret *string

	OpenTelemetry opentelemetry.Config
	sentryDSN     *string

	// Only MSP jobs with a schedule configured will have these values.
	cronSchedule *string
	cronDeadline *time.Duration

	// copies of higher-level configuration
	internal internalContract
	msp      bool
}

func loadDiagnosticsContract(
	logger log.Logger,
	env *Env,
	defaultGCPProjectID string,
	internal internalContract,
	msp bool,
) diagnosticsContract {
	c := diagnosticsContract{
		DiagnosticsSecret: env.GetOptional("DIAGNOSTICS_SECRET", "secret used to authenticate diagnostics requests"),
		OpenTelemetry: opentelemetry.Config{
			GCPProjectID: pointers.Deref(
				env.GetOptional("OTEL_GCP_PROJECT_ID", "GCP project ID for OpenTelemetry export"),
				defaultGCPProjectID),
			OtelSDKDisabled: env.GetBool("OTEL_SDK_DISABLED", "false", "disable OpenTelemetry SDK"),
		},
		sentryDSN: env.GetOptional("SENTRY_DSN", "Sentry error reporting DSN"),

		// Only MSP jobs with a schedule configured will have these values.
		cronSchedule: env.GetOptional("JOB_EXECUTION_CRON_SCHEDULE", "Jobs: expected cron schedule for job executions"),
		cronDeadline: env.GetOptionalInterval("JOB_EXECUTION_DEADLINE", "Jobs: maximum duration to wait for job executions"),

		internal: internal,
		msp:      msp,
	}
	if c.DiagnosticsSecret == nil {
		// We don't recommend using this for sensitive data access, so we just
		// log instead of erroring out entirely.
		message := "DIAGNOSTICS_SECRET not set, diagnostics handlers will not have any authorization checks"
		if c.msp {
			logger.Error(message) // error for visibility in production
		} else {
			logger.Warn(message)
		}
	}
	return c
}

type HandlerRegisterer interface {
	Handle(pattern string, handler http.Handler)
}

type ServiceState interface {
	// Healthy should return nil if the service is healthy, or an error with
	// detailed diagnostics if the service is not healthy. In general:
	//
	// - A healthy state indicates that the service is ready to serve traffic
	//   and do work.
	// - An unhealthy state indicates that the previous revision should continue
	//   to serve traffic.
	//
	// Healthy should be implemented with the above considerations in mind.
	//
	// The query parameter provides the URL query parameters the healtcheck was
	// called with, to implement different "degrees" of healtchecks that can be
	// used by a human operator. The default MSP healthchecks are called without
	// any query parameters, and should be implemented such that they can
	// evaluate quickly.
	//
	// Healthy is only called if the correct service secret is provided.
	Healthy(ctx context.Context, query url.Values) error
}

// RegisterDiagnosticsHandlers registers MSP-standard debug handlers on '/-/...',
// and should be called during service initialization with the service's primary
// endpoint.
//
// ServiceState is a standardized reporter for the state of the service.
func (c diagnosticsContract) RegisterDiagnosticsHandlers(r HandlerRegisterer, state ServiceState) {
	diagnosticsLogger := c.internal.logger.Scoped("diagnostics")

	// Only enable Prometheus metrics endpoint if we are not in a MSP environment,
	// i.e. in local dev. Prometheus can then be optionally spun up to collect
	// a locally running service's metrics.
	if !c.msp {
		// Prometheus standard endpoint is '/metrics', we use the same for
		// convenience.
		r.Handle("/metrics", promhttp.Handler())
		// Warn because this should only be enabled in dev, in production we
		// push metrics via OpenTelemetry to GCP instead.
		diagnosticsLogger.Warn("enabled Prometheus metrics endpoint at '/metrics'")
	}

	// Simple auth-less version reporter
	r.Handle("/-/version", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(c.internal.service.Version()))
	}))

	// Authenticated healthcheck
	r.Handle("/-/healthz", c.DiagnosticsAuthMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := opentelemetry.TracedLogger(r.Context(),
				diagnosticsLogger.Scoped("healthz"))

			if err := state.Healthy(r.Context(), r.URL.Query()); err != nil {
				logger.Warn("service reported not healthy",
					log.String("query", r.URL.Query().Encode()),
					log.Error(err))

				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("healthz: " + err.Error()))
				return
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("healthz: ok"))
		})))
}

// DiagnosticsAuthMiddleware uses DIAGNOSTICS_SECRET to authenticate requests to
// next. It is used for debug endpoints that require some degree of simple
// authentication as a safeguard.
//
// 🚨 SECURITY: Do NOT use this to authenticate sensitive data access. This should
// only be used for governing access to diagnostic information (and should
// still be treated with great care like any other application secrets).
func (c diagnosticsContract) DiagnosticsAuthMiddleware(next http.Handler) http.Handler {
	hasDiagnosticsSecret := func(w http.ResponseWriter, r *http.Request) (yes bool) {
		if c.DiagnosticsSecret == nil {
			return true
		}

		token, err := extractBearer(r.Header)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
			return false
		}

		if subtle.ConstantTimeCompare([]byte(token), []byte(*c.DiagnosticsSecret)) == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("unauthorized"))
			return false
		}
		return true
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !hasDiagnosticsSecret(w, r) {
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ConfigureSentry enables Sentry error log reporting for
// github.com/sourcegraph/log. This does not need to be used if you are already
// using MSP runtime initialization.
//
// The logging library MUST have been initialized with the Sentry sink:
//
//	liblog := log.Init(res, log.NewSentrySink())
//	defer liblog.Sync()
//
// The returned liblog is *log.PostInitCallbacks that is accepted by this method.
// Configuration updates are applied to all loggers, even if they are already
// initialized.
func (c diagnosticsContract) ConfigureSentry(liblog *log.PostInitCallbacks) bool {
	var sentryEnabled bool
	if c.useSentry() {
		liblog.Update(func() log.SinksConfig {
			sentryEnabled = true
			return log.SinksConfig{
				Sentry: &log.SentrySink{
					ClientOptions: sentry.ClientOptions{
						Dsn:         *c.sentryDSN,
						Environment: c.internal.environmentID,
					},
				},
			}
		})()
	}
	return sentryEnabled
}

// JobExecutionCheckIn checks in to indicate a job execution is starting, and
// returns a callback that should be called when the job execution is complete.
// This is only valid for MSP cron jobs.
//
// Direct usage example:
//
//	func work(ctx context.Context) (err error) {
//		done, err := c.Diagnostics.JobExecutionCheckIn(ctx)
//		if err != nil { /* failed to register check-in */ }
//		err = ... do work ...
//		done(err)
//		return
//	}
//
// Note the use of named returns in order to correctly capture the final error.
// Various contract environment variables MUST be set for Sentry monitor check-ins
// to be enabled, otherwise this method will only render log entries - required
// variables are set by MSP infrastructure.
func (c diagnosticsContract) JobExecutionCheckIn(ctx context.Context) (string, func(err error), error) {
	// All values must be set by MSP infrastructure
	useSentryCronMonitor := c.useSentry() &&
		c.cronSchedule != nil &&
		c.cronDeadline != nil

	executionID := uuid.Must(uuid.NewV7()).String()
	logger := opentelemetry.TracedLogger(ctx, c.internal.logger).
		Scoped("execution.checkin").
		With(
			log.String("executionID", executionID),
			log.Bool("useSentryCronMonitor", useSentryCronMonitor),
			log.Stringp("cronSchedule", c.cronSchedule),
			log.Durationp("cronDeadline", c.cronDeadline),
		)

	start := time.Now()
	logger.Info("job execution starting")

	logCompletion := func(err error) {
		d := log.Duration("duration", time.Since(start))
		if err != nil {
			logger.Error("job execution failed", log.Error(err), d)
		} else {
			logger.Info("job execution succeeded", d)
		}
	}

	if !useSentryCronMonitor {
		return executionID, logCompletion, nil
	}

	// Set up Sentry client with some metadata, similar to sourcegraph/log
	// https://github.com/sourcegraph/log/blob/main/internal/sinkcores/sentrycore/worker.go#L96
	sentryClient, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:         *c.sentryDSN,
		Release:     c.internal.service.Version(),
		Environment: c.internal.environmentID,
		Tags: map[string]string{
			"scope":                    "job.execution.checkin",
			"resource.service.name":    c.internal.service.Name(),
			"resource.service.version": c.internal.service.Version(),
		},
	})
	if err != nil {
		return executionID, logCompletion, errors.Wrap(err, "initialize Sentry client from configured DSN")
	}

	monitor := fmt.Sprintf("%s-%s", c.internal.service.Name(), c.internal.environmentID)
	monitorConfig := &sentry.MonitorConfig{
		Schedule:   sentry.CrontabSchedule(*c.cronSchedule),
		MaxRuntime: int64(math.Ceil(c.cronDeadline.Minutes())),
		// Half the deadline is used as a margin to allow for a grace period
		// for execution to report they are active
		CheckInMargin: int64(math.Ceil(c.cronDeadline.Minutes() / 2)),
		// When running in MSP, the timezone is always set to UTC.
		Timezone: "Etc/UTC",
	}

	// Also reflect trace into Sentry's specialized trace context
	// https://develop.sentry.dev/sdk/event-payloads/contexts/#trace-context
	// Only the trace context is supported right now, it's not possible to
	// attach error details - it's expected that an error will be reported
	// separately with a trace from logger.
	// https://develop.sentry.dev/sdk/check-ins/
	scope := sentry.NewScope()
	if span := trace.SpanContextFromContext(ctx); span.HasTraceID() {
		scope.SetContext("trace", sentry.Context{
			"trace_id": span.TraceID().String(),
			"span_id":  span.SpanID().String(),
		})
	}

	ev := sentryClient.CaptureCheckIn(
		&sentry.CheckIn{
			ID:          sentry.EventID(executionID),
			MonitorSlug: monitor,
			Status:      sentry.CheckInStatusInProgress,
		},
		monitorConfig,
		scope)
	return executionID, func(err error) {
		defer sentryClient.Flush(time.Second)

		status := sentry.CheckInStatusOK
		if err != nil {
			status = sentry.CheckInStatusError
		}

		logCompletion(err)
		_ = sentryClient.CaptureCheckIn(
			&sentry.CheckIn{
				ID:          *ev,
				MonitorSlug: monitor,
				Status:      status,
				Duration:    time.Since(start),
			},
			monitorConfig,
			scope)
	}, nil
}

func (c diagnosticsContract) useSentry() bool {
	return c.sentryDSN != nil
}
