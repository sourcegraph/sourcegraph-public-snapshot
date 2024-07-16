package main

import (
	"github.com/Masterminds/semver"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/backport"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var pullRequestIDFlag = cli.Int64Flag{
	Name:     "pullRequestID",
	Usage:    "The pull request ID to backport into the release branch",
	Required: true,
	Aliases:  []string{"p"},
}

var releaseBranchFlag = cli.StringFlag{
	Name:     "releaseBranch",
	Usage:    "The release branch to backport the PR into",
	Required: true,
	Aliases:  []string{"r"},
}

var backportCommand = &cli.Command{
	Name:      "backport",
	Category:  category.Dev,
	Usage:     "Backport commits from main to release branches",
	UsageText: "sg backport -r 5.3 -p 60932",
	Action: func(cmd *cli.Context) error {
		prNumber := pullRequestIDFlag.Get(cmd)
		releaseBranch := releaseBranchFlag.Get(cmd)

		_, err := semver.NewVersion(releaseBranch)
		if err != nil {
			return err
		}

		std.Out.WriteLine(output.Styledf(output.StylePending, "Backporting commits from main to release branch %q for PR %d...", releaseBranch, prNumber))
		return backport.Run(cmd, prNumber, releaseBranch)
	},
	Flags: []cli.Flag{&pullRequestIDFlag, &releaseBranchFlag},
}
