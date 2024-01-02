package main

import (
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/release"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var releaseCommand = &cli.Command{
	Name:     "release",
	Usage:    "Sourcegraph release utilities",
	Category: category.Util,
	Subcommands: []*cli.Command{{
		Name:     "cve-check",
		Usage:    "Check all CVEs found in a buildkite build against a set of preapproved CVEs for a release",
		Category: category.Util,
		Action:   cveCheck,
		Flags: []cli.Flag{
			&buildNumberFlag,
			&referenceUriFlag,
		},
		UsageText: `sg release cve-check -u https://handbook.sourcegraph.com/departments/security/tooling/trivy/4-2-0/ -b 184191`,
	}},
}

var buildNumberFlag = cli.StringFlag{
	Name:     "buildNumber",
	Usage:    "The buildkite build number to check for CVEs",
	Required: true,
	Aliases:  []string{"b"},
}

var referenceUriFlag = cli.StringFlag{
	Name:     "uri",
	Usage:    "A reference url that contains approved CVEs. Often a link to a handbook page eg: https://handbook.sourcegraph.com/departments/security/tooling/trivy/4-2-0/.",
	Required: true,
	Aliases:  []string{"u"},
}

func cveCheck(cmd *cli.Context) error {
	std.Out.WriteLine(output.Styledf(output.StylePending, "Checking release for approved CVEs..."))

	referenceUrl := referenceUriFlag.Get(cmd)
	buildNumber := buildNumberFlag.Get(cmd)

	return release.CveCheck(cmd.Context, buildNumber, referenceUrl, verbose)
}
