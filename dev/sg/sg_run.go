package main

import (
	"context"
	"flag"
	"fmt"
	"sort"
	"strings"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	"github.com/sourcegraph/conc/pool"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/interrupt"
	"github.com/sourcegraph/sourcegraph/lib/cliutil/completions"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func init() {
	postInitHooks = append(postInitHooks,
		func(cmd *cli.Context) {
			// Create 'sg run' help text after flag (and config) initialization
			runCommand.Description = constructRunCmdLongHelp()
		},
		func(cmd *cli.Context) {
			ctx, cancel := context.WithCancel(cmd.Context)
			interrupt.Register(func() {
				cancel()
			})
			cmd.Context = ctx
		},
	)

}

var runCommand = &cli.Command{
	Name:      "run",
	Usage:     "Run the given commands",
	ArgsUsage: "[command]",
	UsageText: `
# Run specific commands
sg run gitserver
sg run frontend

# List available commands (defined under 'commands:' in 'sg.config.yaml')
sg run -help

# Run multiple commands
sg run gitserver frontend repo-updater

# View configuration for a command
sg run -describe jaeger
`,
	Category: category.Dev,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "describe",
			Usage: "Print details about selected run target",
		},
		&cli.BoolFlag{
			Name:  "legacy",
			Usage: "Force run to pick the non-bazel variant of the command",
		},
	},
	Action: runExec,
	BashComplete: completions.CompleteArgs(func() (options []string) {
		config, _ := getConfig()
		if config == nil {
			return
		}
		for name := range config.Commands {
			options = append(options, name)
		}
		return
	}),
}

func runExec(ctx *cli.Context) error {
	config, err := getConfig()
	if err != nil {
		return err
	}
	legacy := ctx.Bool("legacy")

	args := ctx.Args().Slice()
	if len(args) == 0 {
		std.Out.WriteLine(output.Styled(output.StyleWarning, "No command specified"))
		return flag.ErrHelp
	}

	cmds := make([]run.SGConfigCommand, 0, len(args))
	for _, arg := range args {
		if bazelCmd, ok := config.BazelCommands[arg]; ok && !legacy {
			cmds = append(cmds, bazelCmd)
		} else if cmd, ok := config.Commands[arg]; ok {
			cmds = append(cmds, cmd)
		} else {
			std.Out.WriteLine(output.Styledf(output.StyleWarning, "ERROR: command %q not found :(", arg))
			return flag.ErrHelp
		}
	}

	if ctx.Bool("describe") {
		for _, cmd := range cmds {
			out, err := yaml.Marshal(cmd)
			if err != nil {
				return err
			}
			if err = std.Out.WriteMarkdown(fmt.Sprintf("# %s\n\n```yaml\n%s\n```\n\n", cmd.GetConfig().Name, string(out))); err != nil {
				return err
			}
		}

		return nil
	}

	p := pool.New().WithContext(ctx.Context).WithCancelOnError()
	p.Go(func(ctx context.Context) error {
		return run.Commands(ctx, config.Env, verbose, cmds...)
	})

	return p.Wait()
}

func constructRunCmdLongHelp() string {
	var out strings.Builder

	fmt.Fprintf(&out, "Runs the given command. If given a whitespace-separated list of commands it runs the set of commands.\n")

	config, err := getConfig()
	if err != nil {
		out.Write([]byte("\n"))
		// Do not treat error message as a format string
		std.NewOutput(&out, false).WriteWarningf("%s", err.Error())
		return out.String()
	}

	fmt.Fprintf(&out, "\n")
	fmt.Fprintf(&out, "Available commands in `%s`:\n", configFile)

	var names []string
	for name, command := range config.Commands {
		if command.Config.Description != "" {
			name = fmt.Sprintf("%s: %s", name, command.Config.Description)
		}
		names = append(names, name)
	}
	sort.Strings(names)
	fmt.Fprint(&out, "\n* "+strings.Join(names, "\n* "))

	return out.String()
}
