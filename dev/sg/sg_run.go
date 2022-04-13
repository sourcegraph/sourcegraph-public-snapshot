package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func init() {
	postInitHooks = append(postInitHooks, func(cmd *cli.Context) {
		// Create 'sg run' help text after flag (and config) initialization
		runCommand.Description = constructRunCmdLongHelp()
	})
}

var runCommand = &cli.Command{
	Name:        "run",
	Usage:       "Run the given commands",
	ArgsUsage:   "[command]",
	Description: constructRunCmdLongHelp(),
	Category:    CategoryDev,
	Flags: []cli.Flag{
		addToMacOSFirewallFlag,
	},
	Action: execAdapter(runExec),
	BashComplete: completeOptions(func() (options []string) {
		for name := range globalConf.Commands {
			options = append(options, name)
		}
		return
	}),
}

func runExec(ctx context.Context, args []string) error {
	ok, errLine := parseConf(configFlag, overwriteConfigFlag)
	if !ok {
		stdout.Out.WriteLine(errLine)
		os.Exit(1)
	}

	if len(args) == 0 {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "No command specified"))
		return flag.ErrHelp
	}

	var cmds []run.Command
	for _, arg := range args {
		cmd, ok := globalConf.Commands[arg]
		if !ok {
			stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: command %q not found :(", arg))
			return flag.ErrHelp
		}
		cmds = append(cmds, cmd)
	}

	return run.Commands(ctx, globalConf.Env, addToMacOSFirewall, verbose, cmds...)
}
func constructRunCmdLongHelp() string {
	var out strings.Builder

	fmt.Fprintf(&out, "  Runs the given command. If given a whitespace-separated list of commands it runs the set of commands.\n")

	ok, warning := parseConf(configFlag, overwriteConfigFlag)
	if ok {
		fmt.Fprintf(&out, "\n")
		fmt.Fprintf(&out, "AVAILABLE COMMANDS IN %s%s%s\n", output.StyleBold, configFlag, output.StyleReset)

		var names []string
		for name := range globalConf.Commands {
			names = append(names, name)
		}
		sort.Strings(names)
		fmt.Fprint(&out, strings.Join(names, "\n"))
	} else {
		out.Write([]byte("\n"))
		output.NewOutput(&out, output.OutputOpts{}).WriteLine(warning)
	}

	return out.String()
}
