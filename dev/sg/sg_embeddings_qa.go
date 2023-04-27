package main

import (
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/embeddings/qa"
)

var contextCommand = &cli.Command{
	Name:        "embeddings-qa",
	Usage:       "Calculate recall for embeddings",
	Description: "Requires a running embeddings service with embeddings of the Sourcegraph repository.",
	Category:    CategoryDev,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "url",
			Value:   "http://localhost:9991/search",
			Aliases: []string{"u"},
			Usage:   "Run the evaluation against this endpoint",
		},
	},
	Action: func(ctx *cli.Context) error {
		return qa.Run(ctx.String("url"))
	},
}
