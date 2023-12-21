package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"strings"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/hook"
	"github.com/sourcegraph/log/output"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/httpapi/requestlogger"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/response"
	"github.com/sourcegraph/sourcegraph/internal/authbearer"
	"github.com/sourcegraph/sourcegraph/internal/instrumentation"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	sgtrace "github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewDiagnosticsHandler creates a handler for diagnostic endpoints typically served
// on "/-/..." paths. It should be placed before any authentication middleware, since
// we do a simple auth on a static secret instead that is uniquely generated per
// deployment.
func NewDiagnosticsHandler(baseLogger log.Logger, next http.Handler, secret string, sources *actor.Sources) http.Handler {
	baseLogger = baseLogger.Scoped("diagnostics")

	hasValidSecret := func(l log.Logger, w http.ResponseWriter, r *http.Request) (yes bool) {
		token, err := authbearer.ExtractBearer(r.Header)
		if err != nil {
			response.JSONError(l, w, http.StatusBadRequest, err)
			return false
		}

		if token != secret {
			w.WriteHeader(http.StatusUnauthorized)
			return false
		}
		return true
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		// For sanity-checking what's live. Intentionally doesn't require the
		// secret for convenience, and it's a mostly harmless endpoint.
		case "/-/__version":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(version.Version()))

		// For service liveness and readiness probes
		case "/-/healthz":
			logger := sgtrace.Logger(r.Context(), baseLogger)
			if !hasValidSecret(logger, w, r) {
				return
			}

			if err := healthz(r.Context()); err != nil {
				logger.Error("check failed", log.Error(err))

				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("healthz: " + err.Error()))
				return
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("healthz: ok"))

		// Escape hatch to sync all sources.
		case "/-/actor/sync-all-sources":
			logger := sgtrace.Logger(r.Context(), baseLogger)
			if !hasValidSecret(logger, w, r) {
				return
			}

			// Tee log output into "jq --slurp '.[].Body'"-compatible output
			// for ease of use
			var b bytes.Buffer
			logger = hook.Writer(logger, &b, log.LevelInfo, output.FormatJSON)

			if err := sources.SyncAll(r.Context(), logger); err != nil {
				response.JSONError(baseLogger, w, http.StatusInternalServerError, err)
				return
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(b.Bytes())

		// Unknown "/-/..." endpoint
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/-/") {
			instrumentation.HTTPMiddleware(
				"diagnostics",
				requestlogger.Middleware(baseLogger, handler),
				otelhttp.WithPublicEndpoint(),
			).ServeHTTP(w, r)
			return
		}

		// Next handler, we don't care about this request
		next.ServeHTTP(w, r)
	})
}

func healthz(ctx context.Context) error {
	// Check redis health
	rpool := redispool.Cache.Pool()
	rconn, err := rpool.GetContext(ctx)
	if err != nil {
		return errors.Wrap(err, "redis: failed to get conn")
	}
	defer rconn.Close()

	data, err := rconn.Do("PING")
	if err != nil {
		return errors.Wrap(err, "redis: failed to ping")
	}
	if data != "PONG" {
		return errors.New("redis: failed to ping: no pong received")
	}

	return nil
}
