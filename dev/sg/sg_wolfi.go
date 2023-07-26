package main

import (
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/wolfi"
)

var wolfiCommand = &cli.Command{
	Name:      "wolfi",
	ArgsUsage: "",
	UsageText: `
sg wolfi update-hashes
`,
	Usage:    "Automate Wolfi related tasks",
	Category: CategoryDev,
	Subcommands: []*cli.Command{
		{
			Name:   "update-hashes",
			Usage:  "Update Wolfi dependency digests to the latest version",
			Action: wolfi.UpdateHashes,
		},
	},
}
