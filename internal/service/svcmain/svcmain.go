// Package svcmain runs one or more services.
package svcmain

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/hostname"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/profiler"
	"github.com/sourcegraph/sourcegraph/internal/service"
	"github.com/sourcegraph/sourcegraph/internal/singleprogram"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

type Config struct {
	AfterConfigure func() // run after all services' Configure hooks are called
}

// Main is called from the `main` function of the `sourcegraph-oss` and `sourcegraph` commands.
func Main(services []service.Service, config Config) {
	singleprogram.Init()
	run(services, config)
}

// DeprecatedSingleServiceMain is called from the `main` function of a command to start a single
// service (such as frontend or gitserver).
//
// DEPRECATED: Building per-service commands (i.e., a separate binary for frontend, gitserver, etc.)
// is deprecated.
func DeprecatedSingleServiceMain(svc service.Service, config Config) {
	run([]service.Service{svc}, config)
}

func run(services []service.Service, config Config) {
	// Initialize log15. Even though it's deprecated, it's still fairly widely used.
	logging.Init() //nolint:staticcheck // Deprecated, but logs unmigrated to sourcegraph/log look really bad without this.

	liblog := log.Init(log.Resource{
		Name:       env.MyName,
		Version:    version.Version(),
		InstanceID: hostname.Get(),
	},
		// Experimental: DevX is observing how sampling affects the errors signal.
		log.NewSentrySinkWith(
			log.SentrySink{
				ClientOptions: sentry.ClientOptions{SampleRate: 0.2},
			},
		),
	)
	defer liblog.Sync()

	conf.Init()
	go conf.Watch(liblog.Update(conf.GetLogSinks))

	tracer.Init(log.Scoped("tracer", "internal tracer package"), conf.DefaultClient())
	profiler.Init()

	logger := log.Scoped("sourcegraph", "Sourcegraph")
	obctx := observation.NewContext(logger)
	ctx := context.Background()

	// Run the services' Configure funcs before env vars are locked.
	serviceConfigs := make([]env.Config, len(services))
	for i, s := range services {
		serviceConfigs[i] = s.Configure()
	}

	// Validate each service's configuration.
	for i, c := range serviceConfigs {
		if c == nil {
			continue
		}
		if err := c.Validate(); err != nil {
			logger.Fatal("invalid configuration", log.String("service", services[i].Name()), log.Error(err))
		}
	}

	env.Lock()
	env.HandleHelpFlag()

	if config.AfterConfigure != nil {
		config.AfterConfigure()
	}

	// Debug server. TODO(sqs): make it ready when all the constituent services are ready, not immediately.
	ready := make(chan struct{})
	close(ready)
	go debugserver.NewServerRoutine(ready).Start()

	// Start the services.
	for i := range services {
		service := services[i]
		serviceConfig := serviceConfigs[i]
		// TODO(sqs): use goroutine package / MonitorBackgroundRoutines, and let each service Start return a Stop as well
		go func() {
			obctx := observation.ScopedContext("", service.Name(), "", obctx)
			err := service.Start(ctx, obctx, serviceConfig)
			if err != nil {
				logger.Fatal("failed to start service", log.String("service", service.Name()), log.Error(err))
			}
		}()
	}

	select {}
}
