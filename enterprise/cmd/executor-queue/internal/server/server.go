package server

import (
	"fmt"
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

// NewServer returns an HTTP job queue server.
func NewServer(options Options, queueOptions map[string]QueueOptions, observationContext *observation.Context) goroutine.BackgroundRoutine {
	addr := fmt.Sprintf(":%d", options.Port)
	httpHandler := ot.Middleware(httpserver.NewHandler(setupRoutes(options, queueOptions, observationContext)))
	server := httpserver.NewFromAddr(addr, &http.Server{Handler: httpHandler})
	return server
}
