package advise

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/src-cli/internal/scout"
)

func Docker(ctx context.Context, client client.Client, opts ...Option) error {
	cfg := &scout.Config{
		Namespace:    "default",
		Docker:       true,
		Pod:          "",
		Container:    "",
		Spy:          false,
		DockerClient: &client,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	containers, err := client.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return errors.Wrap(err, "could not get list of containers")
	}

	PrintContainers(containers)
	return nil
}

func PrintContainers(containers []types.Container) {
	for _, c := range containers {
		fmt.Println(c.Image)
	}
}
