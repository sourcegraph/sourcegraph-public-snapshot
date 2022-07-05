package main

import (
	"encoding/base64"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/sgconf"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/internal/observation"
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
			{
				Name:        "series-id",
				Usage:       "Gets all insight series ID from the base64 encoded frontend ID",
				Description: `Run 'sg insights series-id' to decode a frontend ID and find all related series IDs`,
				Action:      getInsightSeriesIDAction,
			},
		},
	}
)

func decodeInsightIDAction(cmd *cli.Context) error {
	ids := cmd.Args().Slice()
	if len(ids) == 0 {
		return errors.New("Expected at least 1 id to decode")
	}
	std.Out.WriteNoticef("Decoding %d IDs into insight_view_ids", len(ids))
	for _, id := range ids {
		cleanDecoded, err := decodeIDIntoSeriesID(id)
		if err != nil {
			return err
		}
		std.Out.Writef("\t%s -> %s", id, cleanDecoded)
	}
	return nil
}

func getInsightSeriesIDAction(cmd *cli.Context) error {
	ids := cmd.Args().Slice()
	if len(ids) != 1 {
		return errors.New("Expected 1 id to decode")
	}
	std.Out.WriteNoticef("Finding all insight_series_ids for %s", ids[0])

	ctx := cmd.Context
	logger := log.Scoped("getInsightSeriesIDAction", "")

	// Read the configuration.
	conf, _ := sgconf.Get(configFile, configOverwriteFile)
	if conf == nil {
		return errors.New("failed to read sg.config.yaml. This command needs to be run in the `sourcegraph` repository")
	}

	// Connect to the database.
	conn, err := connections.EnsureNewCodeInsightsDB(postgresdsn.New("", "", conf.GetEnv), "insights", &observation.TestContext)
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

	cleanDecoded, err := decodeIDIntoSeriesID(ids[0])
	if err != nil {
		return err
	}

	q := sqlf.Sprintf(getInsightSeriesIDQuery, cleanDecoded)
	rows, err := db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return errors.Errorf("Got an error when querying database: %v", err)
	}
	defer rows.Close()

	hit := false
	for rows.Next() {
		hit = true
		var insightViewSeries string
		if err := rows.Scan(&insightViewSeries); err != nil {
			return err
		}
		std.Out.WriteSuccessf("%s", insightViewSeries)
	}
	if !hit {
		std.Out.WriteSkippedf("No IDs found")
	}
	return nil
}

func decodeIDIntoSeriesID(id string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return "", errors.Newf("could not decode id %q: %v", id, err)
	}
	// an insight view id is encoded in this format: `insight_view:"[id]"`
	return strings.Trim(strings.TrimLeft(string(decoded), "insight_view:"), "\""), nil
}
