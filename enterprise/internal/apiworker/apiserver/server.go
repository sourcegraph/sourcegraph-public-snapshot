package apiserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/efritz/glock"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

// NewServer returns an HTTP job queue server.
func NewServer(options Options) goroutine.BackgroundRoutine {
	addr := fmt.Sprintf(":%d", options.Port)
	handler := newHandler(options, glock.NewRealClock())
	httpHandler := ot.Middleware(httpserver.NewHandler(handler.setupRoutes))
	server := httpserver.NewFromAddr(addr, &http.Server{Handler: httpHandler})
	janitor := goroutine.NewPeriodicGoroutine(context.Background(), options.CleanupInterval, &handlerWrapper{handler})
	return goroutine.CombinedRoutine{server, janitor}
}

type handlerWrapper struct{ handler *handler }

var _ goroutine.Handler = &handlerWrapper{}

func (hw *handlerWrapper) Handle(ctx context.Context) error { return hw.handler.cleanup(ctx) }
func (hw *handlerWrapper) HandleError(err error)            { log15.Error("Failed to requeue jobs", "err", err) }
func (hw *handlerWrapper) OnShutdown()                      { hw.handler.shutdown() }
