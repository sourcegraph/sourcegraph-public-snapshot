package usage

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
)

func Docker(ctx context.Context, client client.Client, opts ...Option) error {
	cfg := &Config{
		namespace:     "default",
		docker:        true,
		pod:           "",
		container:     "",
		spy:           false,
		k8sClient:     nil,
		dockerClient:  &client,
		metricsClient: nil,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	fmt.Println("Command is routed, but not yet fixed.")
	return nil
}
