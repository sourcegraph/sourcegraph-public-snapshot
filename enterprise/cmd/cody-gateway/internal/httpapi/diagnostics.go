package httpapi

import (
	"context"
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/auth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/response"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	sgtrace "github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewDiagnosticsHandler creates a handler for diagnostic endpoints typically served
// on "/-/..." paths. It should be placed before any authentication middleware, since
// we do a simple auth on a static secret instead.
func NewDiagnosticsHandler(baseLogger log.Logger, next http.Handler, secret string) http.Handler {
	baseLogger = baseLogger.Scoped("diagnostics", "healthz checks")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		// For service liveness and readiness probes
		case "/-/healthz":
			logger := sgtrace.Logger(r.Context(), baseLogger)

			if secret != "" {
				token, err := auth.ExtractBearer(r.Header)
				if err != nil {
					response.JSONError(logger, w, http.StatusBadRequest, err)
					return
				}

				if token != secret {
					w.WriteHeader(http.StatusUnauthorized)
					_, _ = w.Write([]byte("healthz: unauthorized"))
					return
				}
			}

			if err := healthz(r.Context()); err != nil {
				logger.Error("check failed", log.Error(err))

				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("healthz: " + err.Error()))
				return
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("healthz: ok"))

		// For sanity-checking what's live
		case "/-/__version":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(version.Version()))

		// Next handler in the middleware
		default:
			next.ServeHTTP(w, r)
		}
	})
}

func healthz(ctx context.Context) error {
	// Check redis health
	rpool, ok := redispool.Cache.Pool()
	if !ok {
		return errors.New("redis: not available")
	}
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
