package main

import (
	"encoding/base64"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
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
				Action:      decodeInsightIDAction,
			},
		},
	}
)

func decodeInsightIDAction(cmd *cli.Context) error {
	ids := cmd.Args().Slice()
	if len(ids) == 0 {
		return errors.New("Expected at least 1 id to decode")
	}
	std.Out.WriteNoticef("Decoding %d IDs", len(ids))
	for _, id := range ids {
		decoded, err := base64.StdEncoding.DecodeString(id)
		if err != nil {
			return errors.Newf("could not decode id %q: %v", id, err)
		}
		// an insight view id is encoded in this format: `insight_view:"[id]"`
		cleanDecoded := strings.Trim(strings.TrimLeft(string(decoded), "insight_view:"), "\"")
		std.Out.Writef("\t%s -> %s", id, cleanDecoded)
	}
	return nil
}
