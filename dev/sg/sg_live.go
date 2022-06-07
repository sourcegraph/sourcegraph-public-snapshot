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
	Name:      "live",
	ArgsUsage: "<environment-name-or-url>",
	Usage:     "Reports which version of Sourcegraph is currently live in the given environment",
	UsageText: `
# See which version is deployed on a preset environment
sg live cloud
sg live k8s

# See which version is deployed on a custom environment
sg live https://demo.sourcegraph.com

# List environments:
sg live -help
	`,
	Category:    CategoryCompany,
	Description: constructLiveCmdLongHelp(),
	Action:      liveExec,
	BashComplete: completeOptions(func() (options []string) {
		return append(environmentNames(), `https\://...`)
	}),
}

func constructLiveCmdLongHelp() string {
	var out strings.Builder

	fmt.Fprintf(&out, "Prints the Sourcegraph version deployed to the given environment.")
	fmt.Fprintf(&out, "\n\n")
	fmt.Fprintf(&out, "Available preset environments:\n")

	for _, name := range environmentNames() {
		fmt.Fprintf(&out, "\n* %s", name)
	}

	return out.String()
}

func liveExec(ctx *cli.Context) error {
	if ctx.Args().Len() == 0 {
		std.Out.WriteLine(output.Styled(output.StyleWarning, "ERROR: No environment specified"))
		return flag.ErrHelp
	}

	if ctx.Args().Len() != 1 {
		std.Out.WriteLine(output.Styled(output.StyleWarning, "ERROR: Too many arguments"))
		return flag.ErrHelp
	}

	arg := ctx.Args().First()
	e, ok := getEnvironment(ctx.Args().First())
	if !ok {
		if customURL, err := url.Parse(arg); err == nil && customURL.Scheme != "" {
			e = environment{Name: customURL.Host, URL: customURL.String()}
		} else {
			std.Out.WriteLine(output.Styledf(output.StyleWarning, "ERROR: Environment %q not found, or is not a valid URL :(", arg))
			return flag.ErrHelp
		}
	}

	return printDeployedVersion(e)
}
