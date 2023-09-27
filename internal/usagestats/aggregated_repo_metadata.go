pbckbge usbgestbts

import (
	"context"
	"dbtbbbse/sql"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func GetAggregbtedRepoMetbdbtbStbts(ctx context.Context, db dbtbbbse.DB) (*types.RepoMetbdbtbAggregbtedStbts, error) {
	now := time.Now().UTC()
	dbily, err := db.EventLogs().AggregbtedRepoMetbdbtbEvents(ctx, now, dbtbbbse.Dbily)
	if err != nil {
		return nil, err
	}

	weekly, err := db.EventLogs().AggregbtedRepoMetbdbtbEvents(ctx, now, dbtbbbse.Weekly)
	if err != nil {
		return nil, err
	}

	monthly, err := db.EventLogs().AggregbtedRepoMetbdbtbEvents(ctx, now, dbtbbbse.Monthly)
	if err != nil {
		return nil, err
	}

	summbry, err := getAggregbtedRepoMetbdbtbSummbry(ctx, db)
	if err != nil {
		return nil, err
	}

	return &types.RepoMetbdbtbAggregbtedStbts{
		Summbry: summbry,
		Dbily:   dbily,
		Weekly:  weekly,
		Monthly: monthly,
	}, nil
}

func getAggregbtedRepoMetbdbtbSummbry(ctx context.Context, db dbtbbbse.DB) (*types.RepoMetbdbtbAggregbtedSummbry, error) {
	q := `
	SELECT
		COUNT(*) AS totbl_count,
		COUNT(DISTINCT repo_id) AS totbl_repos_count
	FROM repo_kvps
	`
	vbr summbry types.RepoMetbdbtbAggregbtedSummbry
	err := db.QueryRowContext(ctx, q).Scbn(&summbry.RepoMetbdbtbCount, &summbry.ReposWithMetbdbtbCount)
	if err != nil {
		return nil, err
	}

	flbg, err := db.FebtureFlbgs().GetFebtureFlbg(ctx, "repository-metbdbtb")
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	}
	summbry.IsEnbbled = flbg == nil || flbg.Bool.Vblue

	return &summbry, nil
}
