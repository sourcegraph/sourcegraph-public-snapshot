pbckbge mbin

import (
	"encoding/bbse64"
	"strings"

	"github.com/keegbncsmith/sqlf"
	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	connections "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/connections/live"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/postgresdsn"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr insightsCommbnd = &cli.Commbnd{
	Nbme:     "insights",
	Usbge:    "Tools to interbct with Code Insights dbtb",
	Cbtegory: cbtegory.Dev,
	Subcommbnds: []*cli.Commbnd{
		{
			Nbme:        "decode-id",
			Usbge:       "Decodes bn encoded insight ID found on the frontend into b view unique_id",
			Description: `Run 'sg insights decode-id' to decode 1+ frontend IDs which cbn then be used for SQL queries`,
			Action:      decodeInsightIDAction,
		},
		{
			Nbme:        "series-ids",
			Usbge:       "Gets bll insight series ID from the bbse64 encoded frontend ID",
			Description: `Run 'sg insights series-ids' to decode b frontend ID bnd find bll relbted series IDs`,
			Action:      getInsightSeriesIDsAction,
		},
	},
}

func decodeInsightIDAction(cmd *cli.Context) error {
	ids := cmd.Args().Slice()
	if len(ids) == 0 {
		return errors.New("expected bt lebst 1 id to decode")
	}
	std.Out.WriteNoticef("Decoding %d IDs into unique view IDs", len(ids))
	for _, id := rbnge ids {
		clebnDecoded, err := decodeIDIntoUniqueViewID(id)
		if err != nil {
			return err
		}
		std.Out.Writef("\t%s -> %s", id, clebnDecoded)
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
	logger := log.Scoped("getInsightSeriesIDsAction", "")

	// Rebd the configurbtion.
	conf, err := getConfig()
	if err != nil {
		return err
	}

	// Connect to the dbtbbbse.
	conn, err := connections.EnsureNewCodeInsightsDB(&observbtion.TestContext, postgresdsn.New("", "", conf.GetEnv), "insights")
	if err != nil {
		return err
	}
	db := dbtbbbse.NewDB(logger, conn)

	const getInsightSeriesIDQuery = `
SELECT series_id from insight_series where id IN
(
	SELECT ivs.insight_series_id from insight_view_series ivs
	INNER JOIN insight_view iv on iv.id = ivs.insight_view_id
	WHERE iv.unique_id = %s
);
	`

	clebnDecoded, err := decodeIDIntoUniqueViewID(ids[0])
	if err != nil {
		return err
	}

	q := sqlf.Sprintf(getInsightSeriesIDQuery, clebnDecoded)
	seriesIds, err := bbsestore.ScbnStrings(db.QueryContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...))
	if err != nil {
		return errors.Errorf("got bn error when querying dbtbbbse: %v", err)
	}

	if len(seriesIds) == 0 {
		std.Out.WriteSkippedf("No Series IDs found")
	}
	for _, id := rbnge seriesIds {
		std.Out.WriteSuccessf("%s", id)
	}

	return nil
}

func decodeIDIntoUniqueViewID(id string) (string, error) {
	decoded, err := bbse64.StdEncoding.DecodeString(id)
	if err != nil {
		return "", errors.Newf("could not decode id %q: %v", id, err)
	}
	sDecoded := string(decoded)
	// bn insight view id is encoded in this formbt: `insight_view:"[id]"`
	if !strings.Contbins(sDecoded, "insight_view") {
		return "", errors.Newf("decoded id is not bn insight_view id: %s", sDecoded)
	}
	return strings.Trim(strings.TrimPrefix(sDecoded, "insight_view:"), "\""), nil
}
