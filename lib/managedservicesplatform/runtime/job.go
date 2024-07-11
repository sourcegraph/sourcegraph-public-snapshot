package runtime

import (
	"context"
	"flag"
	"os"

	"cloud.google.com/go/profiler"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime/contract"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime/internal/opentelemetry"
)

type Job[ConfigT any] interface {
	contract.ServiceMetadataProvider
	// Execute should use given configuration to perform the desired
	// job execution.
	Execute(
		ctx context.Context,
		logger log.Logger,
		contract JobContract,
		config ConfigT,
	) error
}

// ExecuteJob handles the entire lifecycle of the program performing a
// job execution, and should be the only thing called in a MSP job
// program's main package, for example:
//
//	runtime.ExecuteJob[example.Config](example.Job{})
//
// Where example.Config is your runtime.ConfigLoader implementation, and
// example.Job is your runtime.Job implementation.
func ExecuteJob[
	ConfigT any,
	LoaderT ConfigLoader[ConfigT],
](job Job[ConfigT]) {
	flag.Parse()
	passSanityCheck(job)

	// Resource representing the job
	res := log.Resource{
		Name:       job.Name(),
		Version:    job.Version(),
		Namespace:  "",
		InstanceID: "",
	}

	liblog := log.Init(res, log.NewSentrySink())
	defer liblog.Sync()

	ctx := context.Background()

	// startLogger should only be used within Start - jobs should
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
	ctr := contract.NewJob(log.Scoped("msp.contract"), job, env)

	// Fast-exit with configuration facts if requested
	if *showHelp {
		renderHelp(job, env)
		return
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
			Service:        job.Name(),
			ServiceVersion: job.Version(),
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

	executionID, done, err := ctr.Diagnostics.JobExecutionCheckIn(ctx)
	if err != nil {
		startLogger.Warn("failed to register job execution check-in", log.Error(err))
	}
	startLogger.Info("starting job execution",
		log.String("job", job.Name()),
		log.String("executionID", executionID),
		log.Bool("msp", ctr.MSP),
		log.Bool("sentry", sentryEnabled))

	err = job.Execute(ctx, log.Scoped("job"), ctr, *config)
	done(err)

	// Logging the error is handled by `done(err)` above, so here we just exit
	// with an appropriate exit code.
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
