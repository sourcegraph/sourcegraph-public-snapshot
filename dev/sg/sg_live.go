package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var liveCommand = &cli.Command{
	Name:        "live",
	ArgsUsage:   "<environment>",
	Usage:       "Reports which version of Sourcegraph is currently live in the given environment",
	Category:    CategoryCompany,
	Description: constructLiveCmdLongHelp(),
	Action:      execAdapter(liveExec),
	BashComplete: completeOptions(func() (options []string) {
		return append(environmentNames(), `https\://...`)
	}),
}

func constructLiveCmdLongHelp() string {
	var out strings.Builder

	fmt.Fprintf(&out, "Prints the Sourcegraph version deployed to the given environment.")
	fmt.Fprintf(&out, "\n")
	fmt.Fprintf(&out, "\n")
	fmt.Fprintf(&out, "AVAILABLE PRESET ENVIRONMENTS:\n")

	for _, name := range environmentNames() {
		fmt.Fprintf(&out, "  %s\n", name)
	}

	return out.String()
}

func liveExec(ctx context.Context, args []string) error {
	if len(args) == 0 {
		std.Out.WriteLine(output.Styled(output.StyleWarning, "No environment specified"))
		return flag.ErrHelp
	}

	if len(args) != 1 {
		std.Out.WriteLine(output.Styled(output.StyleWarning, "ERROR: too many arguments"))
		return flag.ErrHelp
	}

	e, ok := getEnvironment(args[0])
	if !ok {
		if customURL, err := url.Parse(args[0]); err == nil {
			e = environment{Name: customURL.Host, URL: customURL.String()}
		} else {
			std.Out.WriteLine(output.Styledf(output.StyleWarning, "ERROR: environment %q not found, or is not a valid URL :(", args[0]))
			return flag.ErrHelp
		}
	}

	return printDeployedVersion(e)
}
