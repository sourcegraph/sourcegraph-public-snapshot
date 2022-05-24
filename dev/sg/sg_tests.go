package main

import (
	"context"
	"flag"
	"fmt"
	"sort"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/sgconf"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func init() {
	postInitHooks = append(postInitHooks, func(cmd *cli.Context) {
		// Create 'sg test' help text after flag (and config) initialization
		testCommand.Description = constructTestCmdLongHelp()
	})
}

var testCommand = &cli.Command{
	Name:      "test",
	ArgsUsage: "<testsuite>",
	Usage:     "Run the given test suite",
	Category:  CategoryDev,
	BashComplete: completeOptions(func() (options []string) {
		config, _ := sgconf.Get(configFile, configOverwriteFile)
		if config == nil {
			return
		}
		for name := range config.Tests {
			options = append(options, name)
		}
		return
	}),
	Action: execAdapter(testExec),
}

func testExec(ctx context.Context, args []string) error {
	config, err := sgconf.Get(configFile, configOverwriteFile)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		std.Out.WriteLine(output.Styled(output.StyleWarning, "No test suite specified"))
		return flag.ErrHelp
	}

	cmd, ok := config.Tests[args[0]]
	if !ok {
		std.Out.WriteLine(output.Styledf(output.StyleWarning, "ERROR: test suite %q not found :(", args[0]))
		return flag.ErrHelp
	}

	return run.Test(ctx, cmd, args[1:], config.Env)
}

func constructTestCmdLongHelp() string {
	var out strings.Builder

	fmt.Fprintf(&out, "  Runs the given testsuite.")

	// Attempt to parse config to list available testsuites, but don't fail on
	// error, because we should never error when the user wants --help output.
	config, err := sgconf.Get(configFile, configOverwriteFile)
	if err != nil {
		out.Write([]byte("\n"))
		std.NewOutput(&out, false).WriteWarningf(err.Error())
		return out.String()
	}

	fmt.Fprintf(&out, "\n\n")
	fmt.Fprintf(&out, "AVAILABLE TESTSUITES IN %s%s%s:\n", output.StyleBold, configFile, output.StyleReset)
	fmt.Fprintf(&out, "\n")

	var names []string
	for name := range config.Tests {
		names = append(names, name)
	}
	sort.Strings(names)
	fmt.Fprint(&out, strings.Join(names, "\n"))

	return out.String()
}
