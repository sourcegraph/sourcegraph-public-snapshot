package server

import (
	"fmt"
	"net/http"

	"github.com/derision-test/glock"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

// NewServer returns an HTTP job queue server.
func NewServer(options Options, observationContext *observation.Context) goroutine.BackgroundRoutine {
	addr := fmt.Sprintf(":%d", options.Port)
	handler := newHandlerWithMetrics(options, glock.NewRealClock(), observationContext)
	httpHandler := ot.Middleware(httpserver.NewHandler(handler.setupRoutes))
	server := httpserver.NewFromAddr(addr, &http.Server{Handler: httpHandler})
	return server
}
