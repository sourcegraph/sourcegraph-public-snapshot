package httpapi

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/events"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

type Config struct {
	AnthropicAccessToken string
	OpenAIAccessToken    string
	OpenAIOrgID          string
}

func NewHandler(logger log.Logger, eventLogger events.Logger, config *Config) http.Handler {
	r := mux.NewRouter()

	// For cluster liveness and readiness probes
	healthzLogger := logger.Scoped("healthz", "healthz checks")
	r.HandleFunc("/-/healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := healthz(r.Context()); err != nil {
			healthzLogger.Error("check failed", log.Error(err))

			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("healthz: " + err.Error()))
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("healthz: ok"))
	})

	r.HandleFunc("/-/__version", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(version.Version()))
	})

	// V1 service routes
	v1router := r.PathPrefix("/v1").Subrouter()
	if config.AnthropicAccessToken != "" {
		v1router.Handle("/completions/anthropic", newAnthropicHandler(logger, eventLogger, config.AnthropicAccessToken)).Methods(http.MethodPost)
	}
	if config.OpenAIAccessToken != "" {
		v1router.Handle("/completions/openai", newOpenAIHandler(logger, eventLogger, config.OpenAIAccessToken, config.OpenAIOrgID)).Methods(http.MethodPost)
	}

	return r
}
