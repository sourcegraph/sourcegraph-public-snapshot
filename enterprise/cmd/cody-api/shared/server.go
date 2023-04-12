package shared

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-api/internal/streaming"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/instrumentation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

const addr = ":4100"

func makeServer() goroutine.BackgroundRoutine {
	logger := log.Scoped("cody-api", "Cody API server")

	handler := makeHandler(logger)
	handler = handlePanic(logger, handler)
	handler = trace.HTTPMiddleware(logger, handler, conf.DefaultClient())
	handler = instrumentation.HTTPMiddleware("", handler)
	handler = actor.HTTPMiddleware(logger, handler)

	return httpserver.NewFromAddr(addr, &http.Server{
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Handler:      handler,
	})
}

func handlePanic(logger log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				err := fmt.Sprintf("%v", rec)
				http.Error(w, fmt.Sprintf("%v", rec), http.StatusInternalServerError)
				logger.Error("recovered from panic", log.String("err", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func makeHandler(logger log.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/healthz", handleHealthCheck(logger))
	mux.Handle("/completions/final", streaming.NewCompletionsFinalHandler(logger))
	mux.Handle("/completions/stream", streaming.NewCompletionsStreamHandler(logger))

	return mux
}

func handleHealthCheck(l log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write([]byte("OK")); err != nil {
			l.Error("failed to write healthcheck response", log.Error(err))
		}
	})
}
