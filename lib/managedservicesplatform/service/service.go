package service

import (
	"context"

	"github.com/sourcegraph/log"
)

type Service[ConfigT any] interface {
	Name() string
	Version() string
	Start(
		ctx context.Context,
		logger log.Logger,
		contract Contract,
		config ConfigT,
	) error
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
	logger := log.Scoped("msp.run", "Managed Services Platform service initialization")

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

	// Start up service
	if err := service.Start(
		ctx,
		log.Scoped("service", service.Name()),
		contract,
		*config,
	); err != nil {
		logger.Fatal("service startup failed", log.Error(err))
	}
}
