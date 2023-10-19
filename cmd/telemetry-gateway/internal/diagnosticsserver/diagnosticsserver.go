package diagnosticsserver

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/authbearer"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

// NewDiagnosticsHandler creates a handler for diagnostic endpoints typically served
// on "/-/..." paths. It should be placed before any authentication middleware, since
// we do a simple auth on a static secret instead that is uniquely generated per
// deployment.
func NewDiagnosticsHandler(
	baseLogger log.Logger,
	secret string,
	healthCheck func(context.Context) error,
) http.Handler {
	baseLogger = baseLogger.Scoped("diagnostics")

	hasValidSecret := func(w http.ResponseWriter, r *http.Request) (yes bool) {
		token, err := authbearer.ExtractBearer(r.Header)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": err.Error(),
			})
			return false
		}

		if token != secret {
			w.WriteHeader(http.StatusUnauthorized)
			return false
		}
		return true
	}

	mux := http.NewServeMux()

	// For sanity-checking what's live. Intentionally doesn't require the
	// secret for convenience, and it's a mostly harmless endpoint.
	mux.HandleFunc("/-/version", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(version.Version()))
	})

	mux.HandleFunc("/-/healthz", func(w http.ResponseWriter, r *http.Request) {
		logger := trace.Logger(r.Context(), baseLogger)
		if !hasValidSecret(w, r) {
			return
		}

		if err := healthCheck(r.Context()); err != nil {
			logger.Error("check failed", log.Error(err))

			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("healthz: " + err.Error()))
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("healthz: ok"))
	})

	return mux
}
