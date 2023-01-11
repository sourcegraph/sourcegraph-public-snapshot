package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/output"
	"github.com/urfave/cli/v2"
)

var (
	BuildCommit = "dev"
	stdOut      *output.Output
)

func main() {
	if err := dx.RunContext(context.Background(), os.Args); err != nil {
		// We want to prefer an already-initialized std.Out no matter what happens,
		// because that can be configured (e.g. with '--disable-output-detection'). Only
		// if something went horribly wrong and std.Out is not yet initialized should we
		// attempt an initialization here.
		if stdOut == nil {
			stdOut = output.NewOutput(os.Stdout, output.OutputOpts{})
		}
		// Do not treat error message as a format string
		log.Fatal(err)
	}
}

var dx = &cli.App{
	Usage:       "The internal CLI used by the DevX team",
	Description: "TODO",
	Version:     BuildCommit,
	Compiled:    time.Now(),
	Commands:    []*cli.Command{scaletestingCommand},
}

var scaletestingCommand = &cli.Command{
	Name:      "scaletesting",
	Aliases:   []string{"sct"},
	UsageText: "TODO",
	Subcommands: []*cli.Command{
		{
			Name:        "dev",
			Description: "TODO",
			Subcommands: []*cli.Command{
				// TODO: add a command to shutdown the machine and one to turn it on.
				{
					Name:        "ssh",
					Description: "SSH to the devbox",
					Action: func(cmd *cli.Context) error {
						args := []string{
							"-c",
							`gcloud compute ssh --zone "us-central1-a" "devx" --project "sourcegraph-scaletesting" --tunnel-through-iap`,
						}

						c := exec.CommandContext(cmd.Context, os.Getenv("SHELL"), args...)
						c.Stdin = os.Stdin
						c.Stdout = os.Stdout
						c.Stderr = os.Stderr

						if err := c.Run(); err != nil {
							return err
						}
						return nil
					},
				},
			},
		},
	},
}
