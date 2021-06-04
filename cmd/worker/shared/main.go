package shared

import (
	"context"
	"net/http"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
)

const addr = ":3189"

func Main() {
	env.Lock()
	env.HandleHelpFlag()
	logging.Init()
	tracer.Init()
	trace.Init(true)

	// Start debug server
	ready := make(chan struct{})
	go debugserver.NewServerRoutine(ready).Start()

	var allRoutines []goroutine.BackgroundRoutine

	// Initialize health server
	server := httpserver.NewFromAddr(addr, &http.Server{
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Handler:      httpserver.NewHandler(nil),
	})
	allRoutines = append(allRoutines, server)

	close(ready)
	goroutine.MonitorBackgroundRoutines(context.Background(), allRoutines...)
}
