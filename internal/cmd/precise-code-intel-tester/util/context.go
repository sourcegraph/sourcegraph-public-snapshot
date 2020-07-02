package util

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// SignalSensitiveContext returns a background context that is canceled after receiving an
// interrupt or terminate signal. A second signal will abort the program. This function returns
// the context and a function that should be  deferred by the caller to clean up internal channels.
func SignalSensitiveContext() (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.Background())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		i := 0
		for range signals {
			cancel()

			if i > 0 {
				os.Exit(1)
			}

			i++
		}
	}()

	return ctx, func() {
		cancel()
		signal.Reset(syscall.SIGINT, syscall.SIGTERM)
		close(signals)
	}
}
