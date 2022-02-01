package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	liveFlagSet = flag.NewFlagSet("sg live", flag.ExitOnError)
	liveCommand = &ffcli.Command{
		Name:       "live",
		ShortUsage: "sg live <environment>",
		ShortHelp:  "Reports which version of Sourcegraph is currently live in the given environment",
		LongHelp:   constructLiveCmdLongHelp(),
		FlagSet:    liveFlagSet,
		Exec:       liveExec,
	}
)

func constructLiveCmdLongHelp() string {
	var out strings.Builder

	fmt.Fprintf(&out, "Prints the Sourcegraph version deployed to the given environment.")
	fmt.Fprintf(&out, "\n")
	fmt.Fprintf(&out, "\n")
	fmt.Fprintf(&out, "AVAILABLE PRESET ENVIRONMENTS\n")

	for _, name := range environmentNames() {
		fmt.Fprintf(&out, "  %s\n", name)
	}

	return out.String()
}

func liveExec(ctx context.Context, args []string) error {
	if len(args) == 0 {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "No environment specified"))
		return flag.ErrHelp
	}

	if len(args) != 1 {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments"))
		return flag.ErrHelp
	}

	e, ok := getEnvironment(args[0])
	if !ok {
		if customURL, err := url.Parse(args[0]); err == nil {
			e = environment{Name: customURL.Host, URL: customURL.String()}
		} else {
			stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: environment %q not found, or is not a valid URL :(", args[0]))
			return flag.ErrHelp
		}
	}

	return printDeployedVersion(e)
}
