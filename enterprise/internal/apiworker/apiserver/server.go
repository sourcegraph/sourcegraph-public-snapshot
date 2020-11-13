package apiserver

import (
	"context"
	"fmt"

	"github.com/efritz/glock"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
)

// NewServer returns an HTTP job queue server.
func NewServer(options Options) (goroutine.BackgroundRoutine, error) {
	addr := fmt.Sprintf(":%d", options.Port)
	handler := newHandler(options, glock.NewRealClock())

	server, err := httpserver.NewFromAddr(addr, httpserver.NewHandler(handler.setupRoutes), httpserver.Options{})
	if err != nil {
		return nil, err
	}

	janitor := goroutine.NewPeriodicGoroutine(context.Background(), options.CleanupInterval, &handlerWrapper{handler})

	return goroutine.CombinedRoutine{server, janitor}, nil
}

type handlerWrapper struct{ handler *handler }

var _ goroutine.Handler = &handlerWrapper{}

func (hw *handlerWrapper) Handle(ctx context.Context) error { return hw.handler.cleanup(ctx) }
func (hw *handlerWrapper) HandleError(err error)            { log15.Error("Failed to requeue jobs", "err", err) }
func (hw *handlerWrapper) OnShutdown()                      { hw.handler.shutdown() }
