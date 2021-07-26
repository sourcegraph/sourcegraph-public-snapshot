package server

import (
	"fmt"
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

// ServerOptions captures the options required for setting up an executor queue
// server.
type ServerOptions struct {
	Port int
}

// NewServer returns an HTTP job queue server.
func NewServer(options ServerOptions, queueOptions map[string]QueueOptions) goroutine.BackgroundRoutine {
	addr := fmt.Sprintf(":%d", options.Port)
	router := setupRoutes(queueOptions)
	httpHandler := ot.Middleware(httpserver.NewHandler(router))
	server := httpserver.NewFromAddr(addr, &http.Server{Handler: httpHandler})
	return server
}
