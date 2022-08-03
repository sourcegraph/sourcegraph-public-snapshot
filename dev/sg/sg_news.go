package main

import (
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/news"
)

var newsCommand = &cli.Command{
	Name:     "news",
	Category: CategoryCompany,
	Action: func(cmd *cli.Context) error {
		content, err := news.Render(cmd.Context, BuildCommit)
		if err != nil {
			return err
		}
		return std.Out.WriteMarkdown(content)
	},
}
