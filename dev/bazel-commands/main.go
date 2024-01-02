package main

import (
	"context"

	"aspect.build/cli/pkg/bazel"
	"aspect.build/cli/pkg/ioutils"
	"aspect.build/cli/pkg/plugin/sdk/v1alpha4/config"
	aspectplugin "aspect.build/cli/pkg/plugin/sdk/v1alpha4/plugin"
	goplugin "github.com/hashicorp/go-plugin"
)

func main() {
	goplugin.Serve(config.NewConfigFor(&customCommandsPlugin{}))
}

type customCommandsPlugin struct {
	aspectplugin.Base
}

func (plugin *customCommandsPlugin) CustomCommands() ([]*aspectplugin.Command, error) {
	return []*aspectplugin.Command{
		aspectplugin.NewCommand(
			"generate",
			"Generates everything configured to be generated with Bazel.",
			"Invokes `bazel run //dev:write_all_generated` to generate all files configured to be generated with Bazel.", func(ctx context.Context, args, bazelStartupArgs []string) error {
				_, err := bazel.New(".").RunCommand(ioutils.DefaultStreams, nil, "run", "//dev:write_all_generated")
				return err
			}),
	}, nil
}
