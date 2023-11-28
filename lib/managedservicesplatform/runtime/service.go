package runtime

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/background"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime/internal/opentelemetry"
)

type Service[ConfigT any] interface {
	Name() string
	Version() string
	// Initialize should use given configuration to build a combined background
	// routine that implements starting and stopping the service.
	Initialize(
		ctx context.Context,
		logger log.Logger,
		contract Contract,
		config ConfigT,
	) (background.CombinedRoutine, error)
}

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
	passSanityCheck()

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
	logger := log.Scoped("msp.run")

	env, err := newEnv()
	if err != nil {
		logger.Fatal("failed to load environment", log.Error(err))
	}

	// Initialize LoaderT implementation as non-zero *ConfigT
	var config LoaderT = new(ConfigT)

	// Load configuration variables from environment
	config.Load(env)
	contract := newContract(env)

	// Enable Sentry error log reporting
	if contract.sentryDSN != nil {
		liblog.Update(func() log.SinksConfig {
			return log.SinksConfig{
				Sentry: &log.SentrySink{
					ClientOptions: sentry.ClientOptions{
						Dsn: *contract.sentryDSN,
					},
				},
			}
		})()
	}

	// Check for environment errors
	if err := env.validate(); err != nil {
		logger.Fatal("environment configuration error encountered", log.Error(err))
	}

	// Initialize things dependent on configuration being loaded
	otelCleanup, err := opentelemetry.Init(ctx, logger, contract.opentelemetryContract, res)
	if err != nil {
		logger.Fatal("failed to initialize OpenTelemetry", log.Error(err))
	}
	defer otelCleanup()

	// Initialize the service
	routine, err := service.Initialize(
		ctx,
		log.Scoped("service"),
		contract,
		*config,
	)
	if err != nil {
		logger.Fatal("service startup failed", log.Error(err))
	}

	// Start service routine, and block until it stops.
	background.Monitor(ctx, routine)
}
