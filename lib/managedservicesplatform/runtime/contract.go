package runtime

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime/internal/opentelemetry"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// Contract loads standardized MSP-provisioned (Managed Services Platform)
// configuration.
type Contract struct {
	// MSP indicates if we are running in a live Managed Services Platform
	// environment. In local development, this should generally be false.
	MSP bool
	// Port is the port the service must listen on.
	Port int
	// ExternalDNSName is the DNS name the service uses, if one is configured.
	ExternalDNSName *string

	// RedisEndpoint is the full connection string of a MSP Redis instance if
	// provisioned, including any prerequisite authentication.
	RedisEndpoint *string

	// PostgreSQL has helpers and configuration for MSP PostgreSQL instances.
	PostgreSQL postgreSQLContract

	// BigQuery has embedded helpers and configuration for MSP-provisioned
	// BigQuery datasets and tables.
	BigQuery bigQueryContract

	// internal configuration for MSP internals that are not exposed to service
	// developers.
	internal internalContract
}

type internalContract struct {
	logger log.Logger
	// service is a reference to the service that is being configured.
	service ServiceMetadata

	diagnosticsSecret *string
	opentelemetry     opentelemetry.Config
	sentryDSN         *string
}

func newContract(logger log.Logger, env *Env, service ServiceMetadata) Contract {
	defaultGCPProjectID := pointers.Deref(env.GetOptional("GOOGLE_CLOUD_PROJECT", "GCP project ID"), "")

	return Contract{
		MSP:             env.GetBool("MSP", "false", "indicates if we are running in a MSP environment"),
		Port:            env.GetInt("PORT", "", "service port"),
		ExternalDNSName: env.GetOptional("EXTERNAL_DNS_NAME", "external DNS name provisioned for the service"),
		RedisEndpoint:   env.GetOptional("REDIS_ENDPOINT", "full Redis address, including any prerequisite authentication"),

		PostgreSQL: loadPostgreSQLContract(env),
		BigQuery:   loadBigQueryContract(env),

		internal: internalContract{
			logger:            logger,
			service:           service,
			diagnosticsSecret: env.GetOptional("DIAGNOSTICS_SECRET", "secret used to authenticate diagnostics requests"),
			opentelemetry: opentelemetry.Config{
				GCPProjectID: pointers.Deref(
					env.GetOptional("OTEL_GCP_PROJECT_ID", "GCP project ID for OpenTelemetry export"),
					defaultGCPProjectID),
			},
			sentryDSN: env.GetOptional("SENTRY_DSN", "Sentry error reporting DSN"),
		},
	}
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
func (c Contract) RegisterDiagnosticsHandlers(r HandlerRegisterer, state ServiceState) {
	// Only enable Prometheus metrics endpoint if we are not in a MSP environment,
	// i.e. in local dev.
	if !c.MSP {
		// Prometheus standard endpoint is '/metrics', we use the same for
		// convenience.
		r.Handle("/metrics", promhttp.Handler())
		// Warn because this should only be enabled in dev
		c.internal.logger.Warn("enabled Prometheus metrics endpoint at '/metrics'")
	}

	// Simple auth-less version reporter
	r.Handle("/-/version", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(c.internal.service.Version()))
	}))

	// Authenticated healthcheck
	r.Handle("/-/healthz", c.DiagnosticsAuthMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := opentelemetry.TracedLogger(r.Context(), c.internal.logger)

			if err := state.Healthy(r.Context(), r.URL.Query()); err != nil {
				logger.Error("service not healthy", log.Error(err))

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
func (c Contract) DiagnosticsAuthMiddleware(next http.Handler) http.Handler {
	hasDiagnosticsSecret := func(w http.ResponseWriter, r *http.Request) (yes bool) {
		if c.internal.diagnosticsSecret == nil {
			return true
		}

		token, err := extractBearer(r.Header)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
			return false
		}

		if token != *c.internal.diagnosticsSecret {
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

func extractBearer(h http.Header) (string, error) {
	var token string

	if authHeader := h.Get("Authorization"); authHeader != "" {
		typ := strings.SplitN(authHeader, " ", 2)
		if len(typ) != 2 {
			return "", errors.New("token type missing in Authorization header")
		}
		if strings.ToLower(typ[0]) != "bearer" {
			return "", errors.Newf("invalid token type %s", typ[0])
		}

		token = typ[1]
	}

	return token, nil
}
