package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	testFlagSet = flag.NewFlagSet("sg test", flag.ExitOnError)
	testCommand = &ffcli.Command{
		Name:       "test",
		ShortUsage: "sg test <testsuite>",
		ShortHelp:  "Run the given test suite.",
		LongHelp:   constructTestCmdLongHelp(),
		FlagSet:    testFlagSet,
		Exec:       testExec,
	}
)

func testExec(ctx context.Context, args []string) error {
	ok, errLine := parseConf(*configFlag, *overwriteConfigFlag)
	if !ok {
		stdout.Out.WriteLine(errLine)
		os.Exit(1)
	}

	if len(args) == 0 {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "No test suite specified"))
		return flag.ErrHelp
	}

	cmd, ok := globalConf.Tests[args[0]]
	if !ok {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: test suite %q not found :(", args[0]))
		return flag.ErrHelp
	}

	return run.Test(ctx, cmd, args[1:], globalConf.Env)
}

func constructTestCmdLongHelp() string {
	var out strings.Builder

	fmt.Fprintf(&out, "  Runs the given testsuite.")

	// Attempt to parse config to list available testsuites, but don't fail on
	// error, because we should never error when the user wants --help output.
	cfg := parseConfAndReset()

	if cfg != nil {
		fmt.Fprintf(&out, "\n\n")
		fmt.Fprintf(&out, "AVAILABLE TESTSUITES IN %s%s%s:\n", output.StyleBold, *configFlag, output.StyleReset)
		fmt.Fprintf(&out, "\n")

		var names []string
		for name := range cfg.Tests {
			names = append(names, name)
		}
		sort.Strings(names)
		fmt.Fprint(&out, strings.Join(names, "\n"))

	}

	return out.String()
}
