// The monitoring generator is now called by Bazel targets instead of go generate
//
// To run monitoring generator run:
// - bazel build //monitoring:generate_config # see bazel-bin/monitoring/outputs
// - bazel build //monitoring:generate_config_zip # see bazel-bin/monitoring/monitoring.zip
// - bazel build //monitoring:generate_grafana_config_tar # see bazel-bin/monitoring/monitoring.tar
package main

import (
	"os"

	"github.com/sourcegraph/log"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/monitoring/command"
)

func main() {
	// Configure logger
	if _, set := os.LookupEnv(log.EnvDevelopment); !set {
		os.Setenv(log.EnvDevelopment, "true")
	}
	if _, set := os.LookupEnv(log.EnvLogFormat); !set {
		os.Setenv(log.EnvLogFormat, "console")
	}

	liblog := log.Init(log.Resource{Name: "monitoring-generator"})
	defer liblog.Sync()
	logger := log.Scoped("monitoring")

	// Create an app that only runs the generate command
	app := &cli.App{
		Name: "monitoring-generator",
		Commands: []*cli.Command{
			command.Generate("", "../"),
		},
		DefaultCommand: "generate",
	}
	if err := app.Run(os.Args); err != nil {
		// Render in plain text for human readability
		println(err.Error())
		logger.Fatal("error encountered")
	}
}
