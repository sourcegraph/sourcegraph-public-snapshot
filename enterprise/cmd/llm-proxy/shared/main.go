package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/instrumentation"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/service"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func Main(ctx context.Context, obctx *observation.Context, ready service.ReadyFunc, config *Config) error {
	handler := newHandler(obctx.Logger, config)
	handler = trace.HTTPMiddleware(obctx.Logger, handler, conf.DefaultClient())
	handler = instrumentation.HTTPMiddleware("", handler)
	handler = actor.HTTPMiddleware(obctx.Logger, handler)

	server := httpserver.NewFromAddr(config.Address, &http.Server{
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Handler:      handler,
	})

	// Mark health server as ready and go!
	ready()
	obctx.Logger.Info("service ready", log.String("address", config.Address))

	goroutine.MonitorBackgroundRoutines(ctx, server)

	return nil
}

func newHandler(logger log.Logger, config *Config) http.Handler {
	r := mux.NewRouter()
	s := r.PathPrefix("/v1").Subrouter()

	s.Handle("/completions/anthropic",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r, err := http.NewRequest(http.MethodPost, "https://api.anthropic.com/v1/complete", r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				err = json.NewEncoder(w).Encode(map[string]string{
					"error": fmt.Sprintf("failed to create request: %s", err),
				})
				if err != nil {
					logger.Error("failed to write response", log.Error(err))
				}
				return
			}

			// Mimic headers set by the official Anthropic client:
			// https://sourcegraph.com/github.com/anthropics/anthropic-sdk-typescript@493075d70f50f1568a276ed0cb177e297f5fef9f/-/blob/src/index.ts
			r.Header.Set("Cache-Control", "no-cache")
			r.Header.Set("Accept", "application/json")
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Client", "sourcegraph-llm-proxy/1.0")
			r.Header.Set("X-API-Key", config.Anthropic.AccessToken)

			resp, err := httpcli.ExternalDoer.Do(r)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				err = json.NewEncoder(w).Encode(map[string]string{
					"error": fmt.Sprintf("failed to make request to Anthropic: %s", err),
				})
				if err != nil {
					logger.Error("failed to write response", log.Error(err))
				}
				return
			}
			defer func() { _ = resp.Body.Close() }()

			_, _ = io.Copy(w, resp.Body)
		}),
	).Methods(http.MethodPost)
	return r
}
