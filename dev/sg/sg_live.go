package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/cliutil/completions"
	"github.com/sourcegraph/sourcegraph/lib/cliutil/exit"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var liveCommand = &cli.Command{
	Name:      "live",
	ArgsUsage: "<environment-name-or-url>",
	Usage:     "Reports which version of Sourcegraph is currently live in the given environment",
	UsageText: `
# See which version is deployed on a preset environment
sg live s2
sg live dotcom
sg live k8s
sg live scaletesting

# See which version is deployed on a custom environment
sg live https://demo.sourcegraph.com

# List environments
sg live -help

# Check for commits further back in history
sg live -n 50 s2
	`,
	Category:    category.Company,
	Description: constructLiveCmdLongHelp(),
	Action:      liveExec,
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:    "commits",
			Aliases: []string{"c", "n"},
			Value:   20,
			Usage:   "Number of commits to check for live version",
		},
	},
	BashComplete: completions.CompleteArgs(func() (options []string) {
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

	fmt.Fprintf(&out, "\n\n")
	fmt.Fprintf(&out, "See more information about the deployments schedule here:\n")
	fmt.Fprintf(&out, "https://handbook.sourcegraph.com/departments/engineering/teams/dev-experience/#sourcegraph-instances-operated-by-us")

	return out.String()
}

func liveExec(ctx *cli.Context) error {
	args := ctx.Args().Slice()
	if len(args) == 0 {
		std.Out.WriteLine(output.Styled(output.StyleWarning, "ERROR: No environment specified"))
		return exit.NewEmptyExitErr(1)
	}
	if len(args) != 1 {
		std.Out.WriteLine(output.Styled(output.StyleWarning, "ERROR: Too many arguments"))
		return exit.NewEmptyExitErr(1)
	}

	e, ok := getEnvironment(args[0])
	if !ok {
		if customURL, err := url.Parse(args[0]); err == nil && customURL.Scheme != "" {
			e = environment{Name: customURL.Host, URL: customURL.String()}
		} else {
			std.Out.WriteLine(output.Styledf(output.StyleWarning, "ERROR: Environment %q not found, or is not a valid URL :(", args[0]))
			return exit.NewEmptyExitErr(1)
		}
	}

	return printDeployedVersion(e, ctx.Int("commits"))
}
