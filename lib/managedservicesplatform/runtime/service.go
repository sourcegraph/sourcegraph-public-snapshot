package runtime

import (
	"context"
	"flag"
	"os"

	"cloud.google.com/go/profiler"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/background"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime/contract"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime/internal/opentelemetry"
)

type Service[ConfigT any] interface {
	contract.ServiceMetadataProvider
	// Initialize should use given configuration to build a combined background
	// routine (such as background.CombinedRoutine or background.LIFOStopRoutine)
	// that implements starting and stopping the service.
	Initialize(
		ctx context.Context,
		logger log.Logger,
		contract ServiceContract,
		config ConfigT,
	) (background.Routine, error)
}

var showHelp = flag.Bool("help", false, "Show service help text")

// Start handles the entire lifecycle of the program running Service, and should
// be the only thing called in a MSP program's main package, for example:
//
//	runtime.Start[example.Config](example.Service{})
//
// Where example.Config is your runtime.ConfigLoader implementation, and
// example.Service is your runtime.Service implementation.
func Start[
	ConfigT any,
	LoaderT ConfigLoader[ConfigT],
](service Service[ConfigT]) {
	flag.Parse()
	passSanityCheck(service)

	// Resource representing the service
	res := log.Resource{
		Name:       service.Name(),
		Version:    service.Version(),
		Namespace:  "",
		InstanceID: "",
	}

	liblog := log.Init(res, log.NewSentrySink())
	defer liblog.Sync()

	ctx := context.Background()

	// startLogger should only be used within Start - longer-lived processes should
	// create a separate top-level startLogger for their usage.
	startLogger := log.Scoped("msp.start")

	env, err := contract.ParseEnv(os.Environ())
	if err != nil {
		startLogger.Fatal("failed to load environment", log.Error(err))
	}

	// Initialize LoaderT implementation as non-zero *ConfigT
	var config LoaderT = new(ConfigT)

	// Load configuration variables from environment
	config.Load(env)
	ctr := contract.NewService(log.Scoped("msp.contract"), service, env)

	// Fast-exit with configuration facts if requested
	if *showHelp {
		renderHelp(service, env)
		os.Exit(0)
	}

	// Enable Sentry error log reporting
	sentryEnabled := ctr.Diagnostics.ConfigureSentry(liblog)

	// Check for environment errors
	if err := env.Validate(); err != nil {
		startLogger.Fatal("environment configuration error encountered", log.Error(err))
	}

	// Initialize things dependent on configuration being loaded
	otelCleanup, err := opentelemetry.Init(ctx, log.Scoped("msp.otel"), ctr.Diagnostics.OpenTelemetry, res)
	if err != nil {
		startLogger.Fatal("failed to initialize OpenTelemetry", log.Error(err))
	}
	defer otelCleanup()

	if ctr.MSP {
		if err := profiler.Start(profiler.Config{
			Service:        service.Name(),
			ServiceVersion: service.Version(),
			// Options used in sourcegraph/sourcegraph
			MutexProfiling: true,
			AllocForceGC:   true,
		}); err != nil {
			// For now, keep this optional and don't prevent startup
			startLogger.Error("failed to initialize profiler", log.Error(err))
		} else {
			startLogger.Debug("Cloud Profiler enabled")
		}
	}

	// Initialize the service
	routine, err := service.Initialize(
		ctx,
		log.Scoped("service"),
		ctr,
		*config,
	)
	if err != nil {
		startLogger.Fatal("service startup failed", log.Error(err))
	}

	// Start service routine, and block until it stops.
	startLogger.Info("starting service",
		log.Int("port", ctr.Port),
		log.Bool("msp", ctr.MSP),
		log.Bool("sentry", sentryEnabled))
	err = background.Monitor(ctx, routine)
	if err != nil {
		startLogger.Error("error stopping service routine", log.Error(err))
	}
	startLogger.Info("service stopped")
}
