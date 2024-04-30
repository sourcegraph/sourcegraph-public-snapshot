package shared

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/sourcegraph/log"
	sglogr "github.com/sourcegraph/log/logr"

	"github.com/sourcegraph/sourcegraph/internal/appliance"
	pb "github.com/sourcegraph/sourcegraph/internal/appliance/v1"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/service"
)

var onlyOneSignalHandler = make(chan struct{})

func Start(ctx context.Context, observationCtx *observation.Context, ready service.ReadyFunc, config *Config) error {
	logger := observationCtx.Logger
	logr := sglogr.New(logger)

	ctrl.SetLogger(logr)

	k8sClient, err := client.New(config.k8sConfig, client.Options{})
	if err != nil {
		logger.Error("unable to create kubernetes client", log.Error(err))
		return err
	}

	app := appliance.NewAppliance(k8sClient)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Logger: logr,
		Metrics: metricsserver.Options{
			BindAddress:   config.metrics.addr,
			SecureServing: config.metrics.secure,
		},
	})
	if err != nil {
		logger.Error("unable to start manager", log.Error(err))
		return err
	}

	if err = (&appliance.Reconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		logger.Error("unable to create the appliance controller", log.Error(err))
		return err
	}

	// Mark health server as ready
	ready()

	listener, err := net.Listen("tcp", config.grpc.addr)
	if err != nil {
		logger.Error("unable to create tcp listener", log.Error(err))
		return err
	}

	grpcServer := makeGRPCServer(logger, app)

	g, ctx := errgroup.WithContext(ctx)
	ctx = shutdownOnSignal(ctx)

	g.Go(func() error {
		logger.Info("gRPC server listening", log.String("address", listener.Addr().String()))
		if err := grpcServer.Serve(listener); err != nil {
			logger.Error("problem running gRPC server", log.Error(err))
			return err
		}
		return nil
	})

	g.Go(func() error {
		logger.Info("starting manager")
		if err := mgr.Start(ctx); err != nil {
			logger.Error("problem running manager", log.Error(err))
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		grpcServer.GracefulStop()
		logger.Info("shutting down gRPC server gracefully")
		return ctx.Err()
	})

	return g.Wait()
}

func makeGRPCServer(logger log.Logger, server pb.ApplianceServiceServer) *grpc.Server {
	grpcServer := defaults.NewServer(logger)
	pb.RegisterApplianceServiceServer(grpcServer, server)

	return grpcServer
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
