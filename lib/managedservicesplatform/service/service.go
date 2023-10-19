package service

import (
	"context"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/lib/background"
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

// Run handles the entire lifecycle of the program running Service.
func Run[
	ConfigT any,
	LoaderT ConfigLoader[ConfigT],
](service Service[ConfigT]) {
	passSanityCheck()

	r := log.Resource{
		Name:       service.Name(),
		Version:    service.Version(),
		Namespace:  "",
		InstanceID: "",
	}

	liblog := log.Init(r)
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

	// Check for environment errors
	if err := env.validate(); err != nil {
		logger.Fatal("environment configuration error encountered", log.Error(err))
	}

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
