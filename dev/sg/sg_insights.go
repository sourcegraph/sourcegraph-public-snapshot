package main

import (
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	insightsCommand = &cli.Command{
		Name:     "insights",
		Usage:    "Tools to interact with Code Insights data",
		Category: CategoryUtil,
		Subcommands: []*cli.Command{
			{
				Name:        "decode-id",
				Usage:       "Decodes an encoded insight ID found on the frontend into an insight_view_id",
				Description: `Run 'sg insights decode-id' to decode 1+ frontend IDs which can then be used for SQL queries`,
				Action:      decodeInsightAction,
			},
		},
	}
)

func decodeInsightAction(cmd *cli.Context) error {
	ids := cmd.Args().Slice()
	if len(ids) == 0 {
		writeFailureLinef("Unexpected argument usage")
		return errors.New("Expected at least 1 id to decode")
	}
	return nil
}
