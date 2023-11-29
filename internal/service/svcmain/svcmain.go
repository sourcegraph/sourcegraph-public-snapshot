// Package svcmain runs one or more services.
package svcmain

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/getsentry/sentry-go"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/output"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/hostname"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/profiler"
	sgservice "github.com/sourcegraph/sourcegraph/internal/service"
	"github.com/sourcegraph/sourcegraph/internal/singleprogram"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

// Main is called from the `main` function of `cmd/sourcegraph`.
//
// args is the commandline arguments (usually os.Args).
func Main(services []sgservice.Service, args []string) {
	// Unlike other sourcegraph binaries we expect to be run
	// by a user instead of deployed to a cloud. So adjust the default output
	// format before initializing log.
	if _, ok := os.LookupEnv(log.EnvLogFormat); !ok && deploy.IsSingleBinary() {
		os.Setenv(log.EnvLogFormat, string(output.FormatConsole))
	}

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

	app := cli.NewApp()
	app.Name = filepath.Base(args[0])
	app.Usage = "The Cody app"
	app.Version = version.Version()
	app.Flags = []cli.Flag{
		&cli.PathFlag{
			Name:        "cacheDir",
			DefaultText: "OS default cache",
			Usage:       "Which directory should be used to cache data",
			EnvVars:     []string{"SRC_APP_CACHE"},
			TakesFile:   false,
			Action: func(ctx *cli.Context, p cli.Path) error {
				return os.Setenv("SRC_APP_CACHE", p)
			},
		},
		&cli.PathFlag{
			Name:        "configDir",
			DefaultText: "OS default config",
			Usage:       "Directory where the configuration should be saved",
			EnvVars:     []string{"SRC_APP_CONFIG"},
			TakesFile:   false,
			Action: func(ctx *cli.Context, p cli.Path) error {
				return os.Setenv("SRC_APP_CONFIG", p)
			},
		},
	}
	app.Action = func(_ *cli.Context) error {
		logger := log.Scoped("sourcegraph")
		cleanup := singleprogram.Init(logger)
		defer func() {
			err := cleanup()
			if err != nil {
				logger.Error("cleaning up", log.Error(err))
			}
		}()
		run(liblog, logger, services, nil)
		return nil
	}

	if err := app.Run(args); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// SingleServiceMain is called from the `main` function of a command to start a single
// service (such as frontend or gitserver). It assumes the service can access site
// configuration and initializes the conf package, and sets up some default hooks for
// watching site configuration for instrumentation services like tracing and logging.
//
// If your service cannot access site configuration, use SingleServiceMainWithoutConf
// instead.
func SingleServiceMain(svc sgservice.Service) {
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
	logger := log.Scoped("sourcegraph")
	run(liblog, logger, []sgservice.Service{svc}, nil)
}

// OutOfBandConfiguration declares additional configuration that happens continuously,
// separate from service startup. In most cases this is configuration based on site config
// (the conf package).
type OutOfBandConfiguration struct {
	// Logging is used to configure logging.
	Logging conf.LogSinksSource

	// Tracing is used to configure tracing.
	Tracing tracer.WatchableConfigurationSource
}

// SingleServiceMainWithConf is called from the `main` function of a command to start a single
// service WITHOUT site configuration enabled by default. This is only useful for services
// that are not part of the core Sourcegraph deployment, such as executors and managed
// services. Use with care!
func SingleServiceMainWithoutConf(svc sgservice.Service, oobConfig OutOfBandConfiguration) {
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
	logger := log.Scoped("sourcegraph")
	run(liblog, logger, []sgservice.Service{svc}, &oobConfig)
}

func run(
	liblog *log.PostInitCallbacks,
	logger log.Logger,
	services []sgservice.Service,
	// If nil, will use site config
	oobConfig *OutOfBandConfiguration,
) {
	defer liblog.Sync()

	// Initialize log15. Even though it's deprecated, it's still fairly widely used.
	logging.Init() //nolint:staticcheck // Deprecated, but logs unmigrated to sourcegraph/log look really bad without this.

	// If no oobConfig is provided, we're in conf mode
	if oobConfig == nil {
		conf.Init()
		oobConfig = &OutOfBandConfiguration{
			Logging: conf.NewLogsSinksSource(conf.DefaultClient()),
			Tracing: tracer.ConfConfigurationSource{WatchableSiteConfig: conf.DefaultClient()},
		}
	}

	if oobConfig.Logging != nil {
		go oobConfig.Logging.Watch(liblog.Update(oobConfig.Logging.SinksConfig))
	}
	if oobConfig.Tracing != nil {
		tracer.Init(log.Scoped("tracer"), oobConfig.Tracing)
	}

	profiler.Init()

	obctx := observation.NewContext(logger)
	ctx := context.Background()

	allReady := make(chan struct{})

	// Run the services' Configure funcs before env vars are locked.
	var (
		serviceConfigs          = make([]env.Config, len(services))
		allDebugserverEndpoints []debugserver.Endpoint
	)
	for i, s := range services {
		var debugserverEndpoints []debugserver.Endpoint
		serviceConfigs[i], debugserverEndpoints = s.Configure()
		allDebugserverEndpoints = append(allDebugserverEndpoints, debugserverEndpoints...)
	}

	// Validate each service's configuration.
	//
	// This cannot be done for executor, see the executorcmd package for details.
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

	// Start the debug server. The ready boolean state it publishes will become true when *all*
	// services report ready.
	var allReadyWG sync.WaitGroup
	var allDoneWG sync.WaitGroup
	go debugserver.NewServerRoutine(allReady, allDebugserverEndpoints...).Start()

	// Start the services.
	for i := range services {
		service := services[i]
		serviceConfig := serviceConfigs[i]
		allReadyWG.Add(1)
		allDoneWG.Add(1)
		go func() {
			// TODO(sqs): TODO(single-binary): Consider using the goroutine package and/or the errgroup package to report
			// errors and listen to signals to initiate cleanup in a consistent way across all
			// services.
			obctx := observation.ContextWithLogger(log.Scoped(service.Name()), obctx)

			// ensure ready is only called once and always call it.
			ready := sync.OnceFunc(allReadyWG.Done)
			defer ready()

			// Don't run executors for Cody App
			if deploy.IsApp() && service.Name() == "executor" {
				logger.Info("Skipping", log.String("service", service.Name()))
				return
			}

			// TODO: It's not clear or enforced but all the service.Start calls block until the service is completed
			// This should be made explicit or refactored to accept to done channel or function in addition to ready.
			err := service.Start(ctx, obctx, ready, serviceConfig)
			allDoneWG.Done()
			if err != nil {
				// Special case in App: continue without executor if it fails to start.
				if deploy.IsApp() && service.Name() == "executor" {
					logger.Warn("failed to start service (skipping)", log.String("service", service.Name()), log.Error(err))
				} else {
					logger.Fatal("failed to start service", log.String("service", service.Name()), log.Error(err))
				}
			}
		}()
	}

	// Pass along the signal to the debugserver that all started services are ready.
	go func() {
		allReadyWG.Wait()
		close(allReady)
	}()

	// wait for all services to stop
	allDoneWG.Wait()
}
