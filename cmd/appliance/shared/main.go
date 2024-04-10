package shared

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/service"
)

var onlyOneSignalHandler = make(chan struct{})

func Start(ctx context.Context, observationCtx *observation.Context, ready service.ReadyFunc, config *Config) error {
	logger := observationCtx.Logger

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{})
	if err != nil {
		logger.Fatal("unable to start manager", log.Error(err))
		return err
	}

	// Mark health server as ready
	ready()

	logger.Info("starting manager")
	if err := mgr.Start(shutdownOnSignal(ctx)); err != nil {
		logger.Fatal("problem running manager", log.Error(err))
		return err
	}

	return nil
}

// shutdownOnSignal registers for SIGTERM and SIGINT. A context is returned
// which is canceled on one of these signals. If a second signal is caught, the program
// is terminated with exit code 1.
func shutdownOnSignal(ctx context.Context) context.Context {
	close(onlyOneSignalHandler)

	ctx, cancel := context.WithCancel(ctx)

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cancel() // first signal. Cancel context.
		<-c
		os.Exit(1) // second signal. Exit now.
	}()

	return ctx
}
