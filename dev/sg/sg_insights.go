package main

import (
	"encoding/base64"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var insightsCommand = &cli.Command{
	Name:     "insights",
	Usage:    "Tools to interact with Code Insights data",
	Category: category.Dev,
	Subcommands: []*cli.Command{
		{
			Name:        "decode-id",
			Usage:       "Decodes an encoded insight ID found on the frontend into a view unique_id",
			Description: `Run 'sg insights decode-id' to decode 1+ frontend IDs which can then be used for SQL queries`,
			Action:      decodeInsightIDAction,
		},
		{
			Name:        "series-ids",
			Usage:       "Gets all insight series ID from the base64 encoded frontend ID",
			Description: `Run 'sg insights series-ids' to decode a frontend ID and find all related series IDs`,
			Action:      getInsightSeriesIDsAction,
		},
	},
}

func decodeInsightIDAction(cmd *cli.Context) error {
	ids := cmd.Args().Slice()
	if len(ids) == 0 {
		return errors.New("expected at least 1 id to decode")
	}
	std.Out.WriteNoticef("Decoding %d IDs into unique view IDs", len(ids))
	for _, id := range ids {
		cleanDecoded, err := decodeIDIntoUniqueViewID(id)
		if err != nil {
			return err
		}
		std.Out.Writef("\t%s -> %s", id, cleanDecoded)
	}
	return nil
}

func getInsightSeriesIDsAction(cmd *cli.Context) error {
	ids := cmd.Args().Slice()
	if len(ids) != 1 {
		return errors.New("expected 1 id to decode")
	}
	std.Out.WriteNoticef("Finding the Series IDs for %s", ids[0])

	ctx := cmd.Context
	logger := log.Scoped("getInsightSeriesIDsAction")

	// Read the configuration.
	conf, err := getConfig()
	if err != nil {
		return err
	}

	// Connect to the database.
	conn, err := connections.EnsureNewCodeInsightsDB(&observation.TestContext, postgresdsn.New("", "", conf.GetEnv), "insights")
	if err != nil {
		return err
	}
	db := database.NewDB(logger, conn)

	const getInsightSeriesIDQuery = `
SELECT series_id from insight_series where id IN
(
	SELECT ivs.insight_series_id from insight_view_series ivs
	INNER JOIN insight_view iv on iv.id = ivs.insight_view_id
	WHERE iv.unique_id = %s
);
	`

	cleanDecoded, err := decodeIDIntoUniqueViewID(ids[0])
	if err != nil {
		return err
	}

	q := sqlf.Sprintf(getInsightSeriesIDQuery, cleanDecoded)
	seriesIds, err := basestore.ScanStrings(db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...))
	if err != nil {
		return errors.Errorf("got an error when querying database: %v", err)
	}

	if len(seriesIds) == 0 {
		std.Out.WriteSkippedf("No Series IDs found")
	}
	for _, id := range seriesIds {
		std.Out.WriteSuccessf("%s", id)
	}

	return nil
}

func decodeIDIntoUniqueViewID(id string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return "", errors.Newf("could not decode id %q: %v", id, err)
	}
	sDecoded := string(decoded)
	// an insight view id is encoded in this format: `insight_view:"[id]"`
	if !strings.Contains(sDecoded, "insight_view") {
		return "", errors.Newf("decoded id is not an insight_view id: %s", sDecoded)
	}
	return strings.Trim(strings.TrimPrefix(sDecoded, "insight_view:"), "\""), nil
}
