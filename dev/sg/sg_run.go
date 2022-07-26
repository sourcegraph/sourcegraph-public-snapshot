package main

import (
	"flag"
	"fmt"
	"sort"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func init() {
	postInitHooks = append(postInitHooks, func(cmd *cli.Context) {
		// Create 'sg run' help text after flag (and config) initialization
		runCommand.Description = constructRunCmdLongHelp()
	})
}

var runCommand = &cli.Command{
	Name:      "run",
	Usage:     "Run the given commands",
	ArgsUsage: "[command]",
	UsageText: `
# Run specific commands:
sg run gitserver
sg run frontend

# List available commands (defined under 'commands:' in 'sg.config.yaml'):
sg run -help

# Run multiple commands:
sg run gitserver frontend repo-updater
	`,
	Category: CategoryDev,
	Flags:    []cli.Flag{},
	Action:   runExec,
	BashComplete: completeOptions(func() (options []string) {
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

	args := ctx.Args().Slice()
	if len(args) == 0 {
		std.Out.WriteLine(output.Styled(output.StyleWarning, "No command specified"))
		return flag.ErrHelp
	}

	var cmds []run.Command
	for _, arg := range args {
		cmd, ok := config.Commands[arg]
		if !ok {
			std.Out.WriteLine(output.Styledf(output.StyleWarning, "ERROR: command %q not found :(", arg))
			return flag.ErrHelp
		}
		cmds = append(cmds, cmd)
	}

	return run.Commands(ctx.Context, config.Env, verbose, cmds...)
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
		if command.Description != "" {
			name = fmt.Sprintf("%s: %s", name, command.Description)
		}
		names = append(names, name)
	}
	sort.Strings(names)
	fmt.Fprint(&out, "\n* "+strings.Join(names, "\n* "))

	return out.String()
}
