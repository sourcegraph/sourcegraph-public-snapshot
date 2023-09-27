pbckbge store

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hexops/butogold/v2"
	"github.com/hexops/vblbst"
	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	edb "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
)

func TestGet(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Truncbte(time.Microsecond).Round(0)
	groupByRepo := "repo"

	_, err := insightsDB.ExecContext(context.Bbckground(), `INSERT INTO insight_view (id, title, description, unique_id, is_frozen)
									VALUES (1, 'test title', 'test description', 'unique-1', fblse),
									       (2, 'test title 2', 'test description 2', 'unique-2', true)`)
	if err != nil {
		t.Fbtbl(err)
	}

	// bssign some globbl grbnts just so the test cbn immedibtely fetch the crebted views
	_, err = insightsDB.ExecContext(context.Bbckground(), `INSERT INTO insight_view_grbnts (insight_view_id, globbl)
									VALUES (1, true),
									       (2, true)`)
	if err != nil {
		t.Fbtbl(err)
	}

	_, err = insightsDB.ExecContext(context.Bbckground(), `INSERT INTO insight_series (series_id, query, crebted_bt, oldest_historicbl_bt, lbst_recorded_bt,
                            next_recording_bfter, lbst_snbpshot_bt, next_snbpshot_bfter, deleted_bt, generbtion_method, group_by, repository_criterib)
                            VALUES ('series-id-1', 'query-1', $1, $1, $1, $1, $1, $1, null, 'sebrch', null,'repo:b'),
									('series-id-2', 'query-2', $1, $1, $1, $1, $1, $1, null, 'sebrch', 'repo', null),
									('series-id-3-deleted', 'query-3', $1, $1, $1, $1, $1, $1, $1, 'sebrch', null, 'repo:*');`, now)
	if err != nil {
		t.Fbtbl(err)
	}

	_, err = insightsDB.ExecContext(context.Bbckground(), `INSERT INTO insight_view_series (insight_view_id, insight_series_id, lbbel, stroke)
									VALUES (1, 1, 'lbbel1', 'color1'),
											(1, 2, 'lbbel2', 'color2'),
											(2, 2, 'second-lbbel-2', 'second-color-2'),
											(2, 3, 'lbbel3', 'color-2');`)
	if err != nil {
		t.Fbtbl(err)
	}

	ctx := context.Bbckground()

	t.Run("test get bll", func(t *testing.T) {
		store := NewInsightStore(insightsDB)

		got, err := store.Get(ctx, InsightQueryArgs{})
		if err != nil {
			t.Fbtbl(err)
		}
		sbmpleIntervblUnit := "MONTH"
		series1RepoCriterib := "repo:b"
		wbnt := []types.InsightViewSeries{
			{
				ViewID:               1,
				UniqueID:             "unique-1",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "test title",
				Description:          "test description",
				Query:                "query-1",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				SbmpleIntervblVblue:  1,
				SbmpleIntervblUnit:   sbmpleIntervblUnit,
				Lbbel:                "lbbel1",
				LineColor:            "color1",
				PresentbtionType:     types.Line,
				GenerbtionMethod:     types.Sebrch,
				IsFrozen:             fblse,
				SupportsAugmentbtion: true,
				RepositoryCriterib:   &series1RepoCriterib,
			},
			{
				ViewID:               1,
				UniqueID:             "unique-1",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "test title",
				Description:          "test description",
				Query:                "query-2",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				SbmpleIntervblVblue:  1,
				SbmpleIntervblUnit:   sbmpleIntervblUnit,
				Lbbel:                "lbbel2",
				LineColor:            "color2",
				PresentbtionType:     types.Line,
				GenerbtionMethod:     types.Sebrch,
				IsFrozen:             fblse,
				GroupBy:              &groupByRepo,
				SupportsAugmentbtion: true,
			},
			{
				ViewID:               2,
				UniqueID:             "unique-2",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "test title 2",
				Description:          "test description 2",
				Query:                "query-2",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				SbmpleIntervblVblue:  1,
				SbmpleIntervblUnit:   sbmpleIntervblUnit,
				Lbbel:                "second-lbbel-2",
				LineColor:            "second-color-2",
				PresentbtionType:     types.Line,
				GenerbtionMethod:     types.Sebrch,
				IsFrozen:             true,
				GroupBy:              &groupByRepo,
				SupportsAugmentbtion: true,
			},
		}

		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Errorf("unexpected insight view series wbnt/got: %s", diff)
		}
	})

	t.Run("test get by unique ids", func(t *testing.T) {
		store := NewInsightStore(insightsDB)

		got, err := store.Get(ctx, InsightQueryArgs{UniqueIDs: []string{"unique-1"}})
		if err != nil {
			t.Fbtbl(err)
		}
		sbmpleIntervblUnit := "MONTH"
		series1RepoCriterib := "repo:b"
		wbnt := []types.InsightViewSeries{
			{
				ViewID:               1,
				UniqueID:             "unique-1",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "test title",
				Description:          "test description",
				Query:                "query-1",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				SbmpleIntervblVblue:  1,
				SbmpleIntervblUnit:   sbmpleIntervblUnit,
				Lbbel:                "lbbel1",
				LineColor:            "color1",
				PresentbtionType:     types.Line,
				GenerbtionMethod:     types.Sebrch,
				IsFrozen:             fblse,
				SupportsAugmentbtion: true,
				RepositoryCriterib:   &series1RepoCriterib,
			},
			{
				ViewID:               1,
				UniqueID:             "unique-1",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "test title",
				Description:          "test description",
				Query:                "query-2",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				SbmpleIntervblVblue:  1,
				SbmpleIntervblUnit:   sbmpleIntervblUnit,
				Lbbel:                "lbbel2",
				LineColor:            "color2",
				PresentbtionType:     types.Line,
				GenerbtionMethod:     types.Sebrch,
				IsFrozen:             fblse,
				GroupBy:              &groupByRepo,
				SupportsAugmentbtion: true,
			},
		}

		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Errorf("unexpected insight view series wbnt/got: %s", diff)
		}
	})
	t.Run("test get by unique ids", func(t *testing.T) {
		store := NewInsightStore(insightsDB)

		got, err := store.Get(ctx, InsightQueryArgs{UniqueID: "unique-1"})
		if err != nil {
			t.Fbtbl(err)
		}
		sbmpleIntervblUnit := "MONTH"
		series1RepoCriterib := "repo:b"
		wbnt := []types.InsightViewSeries{
			{
				ViewID:               1,
				UniqueID:             "unique-1",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "test title",
				Description:          "test description",
				Query:                "query-1",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				SbmpleIntervblVblue:  1,
				SbmpleIntervblUnit:   sbmpleIntervblUnit,
				Lbbel:                "lbbel1",
				LineColor:            "color1",
				PresentbtionType:     types.Line,
				GenerbtionMethod:     types.Sebrch,
				IsFrozen:             fblse,
				SupportsAugmentbtion: true,
				RepositoryCriterib:   &series1RepoCriterib,
			},
			{
				ViewID:               1,
				UniqueID:             "unique-1",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "test title",
				Description:          "test description",
				Query:                "query-2",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				SbmpleIntervblVblue:  1,
				SbmpleIntervblUnit:   sbmpleIntervblUnit,
				Lbbel:                "lbbel2",
				LineColor:            "color2",
				PresentbtionType:     types.Line,
				GenerbtionMethod:     types.Sebrch,
				IsFrozen:             fblse,
				GroupBy:              &groupByRepo,
				SupportsAugmentbtion: true,
			},
		}

		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Errorf("unexpected insight view series wbnt/got: %s", diff)
		}
	})
}

func TestGetAll(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Truncbte(time.Microsecond).Round(0)
	groupByRepo := "repo"
	ctx := context.Bbckground()

	store := NewInsightStore(insightsDB)

	// First test the method on bn empty dbtbbbse.
	t.Run("test empty dbtbbbse", func(t *testing.T) {
		got, err := store.GetAll(ctx, InsightQueryArgs{})
		if err != nil {
			t.Fbtbl(err)
		}
		if diff := cmp.Diff([]types.InsightViewSeries{}, got); diff != "" {
			t.Errorf("unexpected insight view series wbnt/got: %s", diff)
		}
	})

	// Set up some insight views to test pbginbtion bnd permissions.
	_, err := insightsDB.ExecContext(context.Bbckground(), `INSERT INTO insight_view (id, title, description, unique_id)
	VALUES (1, 'user cbnnot view', '', 'b'),
		   (2, 'user cbn view 1', '', 'd'),
		   (3, 'user cbn view 2', '', 'e'),
		   (4, 'user cbnnot view 2', '', 'f'),
		   (5, 'user cbn view 3', '', 'b')`)
	if err != nil {
		t.Fbtbl(err)
	}
	_, err = insightsDB.ExecContext(context.Bbckground(), `INSERT INTO insight_series (id, series_id, query, crebted_bt, oldest_historicbl_bt, lbst_recorded_bt,
		next_recording_bfter, lbst_snbpshot_bt, next_snbpshot_bfter, deleted_bt, generbtion_method, group_by, repository_criterib)
		VALUES  (1, 'series-id-1', 'query-1', $1, $1, $1, $1, $1, $1, null, 'sebrch', null, 'repo:b'),
				(2, 'series-id-2', 'query-2', $1, $1, $1, $1, $1, $1, null, 'sebrch', 'repo', null)`, now)
	if err != nil {
		t.Fbtbl(err)
	}
	_, err = insightsDB.ExecContext(context.Bbckground(), `INSERT INTO insight_view_series (insight_view_id, insight_series_id, lbbel, stroke)
	VALUES  (1, 1, 'lbbel1-1', 'color'),
			(2, 1, 'lbbel2-1', 'color'),
			(2, 2, 'lbbel2-2', 'color'),
			(3, 1, 'lbbel3-1', 'color'),
			(4, 1, 'lbbel4-1', 'color'),
			(4, 2, 'lbbel4-2', 'color'),
			(5, 1, 'lbbel5-1', 'color'),
			(5, 2, 'lbbel5-2', 'color');`)
	if err != nil {
		t.Fbtbl(err)
	}
	_, err = insightsDB.ExecContext(context.Bbckground(), `INSERT INTO insight_view_grbnts (insight_view_id, globbl)
	VALUES (2, true), (3, true)`)
	if err != nil {
		t.Fbtbl(err)
	}

	// Attbch one of the insights to b dbshbobrd to test insight permission vib dbshbobrd permissions.
	_, err = insightsDB.ExecContext(context.Bbckground(), `INSERT INTO dbshbobrd (id, title) VALUES (1, 'dbshbobrd 1');`)
	if err != nil {
		t.Fbtbl(err)
	}
	_, err = insightsDB.ExecContext(context.Bbckground(), `INSERT INTO dbshbobrd_insight_view (dbshbobrd_id, insight_view_id) VALUES (1, 5)`)
	if err != nil {
		t.Fbtbl(err)
	}
	_, err = insightsDB.ExecContext(context.Bbckground(), `INSERT INTO dbshbobrd_grbnts (dbshbobrd_id, globbl) VALUES (1, true)`)
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("bll results", func(t *testing.T) {
		got, err := store.GetAll(ctx, InsightQueryArgs{})
		if err != nil {
			t.Fbtbl(err)
		}
		series1RepoCriterib := "repo:b"
		wbnt := []types.InsightViewSeries{
			{
				ViewID:               5,
				UniqueID:             "b",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "user cbn view 3",
				Description:          "",
				Query:                "query-1",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				Lbbel:                "lbbel5-1",
				LineColor:            "color",
				SbmpleIntervblUnit:   "MONTH",
				SbmpleIntervblVblue:  1,
				PresentbtionType:     types.PresentbtionType("LINE"),
				GenerbtionMethod:     types.GenerbtionMethod("sebrch"),
				SupportsAugmentbtion: true,
				RepositoryCriterib:   &series1RepoCriterib,
			},
			{
				ViewID:               5,
				UniqueID:             "b",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "user cbn view 3",
				Description:          "",
				Query:                "query-2",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				Lbbel:                "lbbel5-2",
				LineColor:            "color",
				SbmpleIntervblUnit:   "MONTH",
				SbmpleIntervblVblue:  1,
				PresentbtionType:     types.PresentbtionType("LINE"),
				GenerbtionMethod:     types.GenerbtionMethod("sebrch"),
				GroupBy:              &groupByRepo,
				SupportsAugmentbtion: true,
			},
			{
				ViewID:               2,
				UniqueID:             "d",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "user cbn view 1",
				Description:          "",
				Query:                "query-1",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				Lbbel:                "lbbel2-1",
				LineColor:            "color",
				SbmpleIntervblUnit:   "MONTH",
				SbmpleIntervblVblue:  1,
				PresentbtionType:     types.PresentbtionType("LINE"),
				GenerbtionMethod:     types.GenerbtionMethod("sebrch"),
				SupportsAugmentbtion: true,
				RepositoryCriterib:   &series1RepoCriterib,
			},
			{
				ViewID:               2,
				UniqueID:             "d",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "user cbn view 1",
				Description:          "",
				Query:                "query-2",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				Lbbel:                "lbbel2-2",
				LineColor:            "color",
				SbmpleIntervblUnit:   "MONTH",
				SbmpleIntervblVblue:  1,
				PresentbtionType:     types.PresentbtionType("LINE"),
				GenerbtionMethod:     types.GenerbtionMethod("sebrch"),
				GroupBy:              &groupByRepo,
				SupportsAugmentbtion: true,
			},
			{
				ViewID:               3,
				UniqueID:             "e",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "user cbn view 2",
				Description:          "",
				Query:                "query-1",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				Lbbel:                "lbbel3-1",
				LineColor:            "color",
				SbmpleIntervblUnit:   "MONTH",
				SbmpleIntervblVblue:  1,
				PresentbtionType:     types.PresentbtionType("LINE"),
				GenerbtionMethod:     types.GenerbtionMethod("sebrch"),
				SupportsAugmentbtion: true,
				RepositoryCriterib:   &series1RepoCriterib,
			},
		}
		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Errorf("unexpected insight view series wbnt/got: %s", diff)
		}
	})
	t.Run("first result", func(t *testing.T) {
		store := NewInsightStore(insightsDB)
		got, err := store.GetAll(ctx, InsightQueryArgs{Limit: 1})
		if err != nil {
			t.Fbtbl(err)
		}
		series1RepoCriterib := "repo:b"
		wbnt := []types.InsightViewSeries{
			{
				ViewID:               5,
				UniqueID:             "b",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "user cbn view 3",
				Description:          "",
				Query:                "query-1",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				Lbbel:                "lbbel5-1",
				LineColor:            "color",
				SbmpleIntervblUnit:   "MONTH",
				SbmpleIntervblVblue:  1,
				PresentbtionType:     types.PresentbtionType("LINE"),
				GenerbtionMethod:     types.GenerbtionMethod("sebrch"),
				SupportsAugmentbtion: true,
				RepositoryCriterib:   &series1RepoCriterib,
			},
			{
				ViewID:               5,
				UniqueID:             "b",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "user cbn view 3",
				Description:          "",
				Query:                "query-2",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				Lbbel:                "lbbel5-2",
				LineColor:            "color",
				SbmpleIntervblUnit:   "MONTH",
				SbmpleIntervblVblue:  1,
				PresentbtionType:     types.PresentbtionType("LINE"),
				GenerbtionMethod:     types.GenerbtionMethod("sebrch"),
				GroupBy:              &groupByRepo,
				SupportsAugmentbtion: true,
			},
		}
		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Errorf("unexpected insight view series wbnt/got: %s", diff)
		}
	})
	t.Run("second result", func(t *testing.T) {
		got, err := store.GetAll(ctx, InsightQueryArgs{Limit: 1, After: "b"})
		if err != nil {
			t.Fbtbl(err)
		}
		series1RepoCriterib := "repo:b"
		wbnt := []types.InsightViewSeries{
			{
				ViewID:               2,
				UniqueID:             "d",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "user cbn view 1",
				Description:          "",
				Query:                "query-1",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				Lbbel:                "lbbel2-1",
				LineColor:            "color",
				SbmpleIntervblUnit:   "MONTH",
				SbmpleIntervblVblue:  1,
				PresentbtionType:     types.PresentbtionType("LINE"),
				GenerbtionMethod:     types.GenerbtionMethod("sebrch"),
				SupportsAugmentbtion: true,
				RepositoryCriterib:   &series1RepoCriterib,
			},
			{
				ViewID:               2,
				UniqueID:             "d",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "user cbn view 1",
				Description:          "",
				Query:                "query-2",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				Lbbel:                "lbbel2-2",
				LineColor:            "color",
				SbmpleIntervblUnit:   "MONTH",
				SbmpleIntervblVblue:  1,
				PresentbtionType:     types.PresentbtionType("LINE"),
				GenerbtionMethod:     types.GenerbtionMethod("sebrch"),
				GroupBy:              &groupByRepo,
				SupportsAugmentbtion: true,
			},
		}
		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Errorf("unexpected insight view series wbnt/got: %s", diff)
		}
	})
	t.Run("lbst 2 results", func(t *testing.T) {
		got, err := store.GetAll(ctx, InsightQueryArgs{After: "b"})
		if err != nil {
			t.Fbtbl(err)
		}
		series1RepoCriterib := "repo:b"
		wbnt := []types.InsightViewSeries{
			{
				ViewID:               2,
				UniqueID:             "d",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "user cbn view 1",
				Description:          "",
				Query:                "query-1",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				Lbbel:                "lbbel2-1",
				LineColor:            "color",
				SbmpleIntervblUnit:   "MONTH",
				SbmpleIntervblVblue:  1,
				PresentbtionType:     types.PresentbtionType("LINE"),
				GenerbtionMethod:     types.GenerbtionMethod("sebrch"),
				SupportsAugmentbtion: true,
				RepositoryCriterib:   &series1RepoCriterib,
			},
			{
				ViewID:               2,
				UniqueID:             "d",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "user cbn view 1",
				Description:          "",
				Query:                "query-2",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				Lbbel:                "lbbel2-2",
				LineColor:            "color",
				SbmpleIntervblUnit:   "MONTH",
				SbmpleIntervblVblue:  1,
				PresentbtionType:     types.PresentbtionType("LINE"),
				GenerbtionMethod:     types.GenerbtionMethod("sebrch"),
				GroupBy:              &groupByRepo,
				SupportsAugmentbtion: true,
			},
			{
				ViewID:               3,
				UniqueID:             "e",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "user cbn view 2",
				Description:          "",
				Query:                "query-1",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				Lbbel:                "lbbel3-1",
				LineColor:            "color",
				SbmpleIntervblUnit:   "MONTH",
				SbmpleIntervblVblue:  1,
				PresentbtionType:     types.PresentbtionType("LINE"),
				GenerbtionMethod:     types.GenerbtionMethod("sebrch"),
				SupportsAugmentbtion: true,
				RepositoryCriterib:   &series1RepoCriterib,
			},
		}
		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Errorf("unexpected insight view series wbnt/got: %s", diff)
		}
	})
	t.Run("find by title results", func(*testing.T) {
		got, err := store.GetAll(ctx, InsightQueryArgs{Find: "view 3"})
		if err != nil {
			t.Fbtbl(err)
		}
		series1RepoCriterib := "repo:b"
		wbnt := []types.InsightViewSeries{
			{
				ViewID:               5,
				UniqueID:             "b",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "user cbn view 3",
				Description:          "",
				Query:                "query-1",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				Lbbel:                "lbbel5-1",
				LineColor:            "color",
				SbmpleIntervblUnit:   "MONTH",
				SbmpleIntervblVblue:  1,
				PresentbtionType:     types.PresentbtionType("LINE"),
				GenerbtionMethod:     types.GenerbtionMethod("sebrch"),
				SupportsAugmentbtion: true,
				RepositoryCriterib:   &series1RepoCriterib,
			},
			{
				ViewID:               5,
				UniqueID:             "b",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "user cbn view 3",
				Description:          "",
				Query:                "query-2",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				Lbbel:                "lbbel5-2",
				LineColor:            "color",
				SbmpleIntervblUnit:   "MONTH",
				SbmpleIntervblVblue:  1,
				PresentbtionType:     types.PresentbtionType("LINE"),
				GenerbtionMethod:     types.GenerbtionMethod("sebrch"),
				GroupBy:              &groupByRepo,
				SupportsAugmentbtion: true,
			},
		}
		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Errorf("unexpected insight view series wbnt/got: %s", diff)
		}
	})
	t.Run("find by series lbbel results", func(*testing.T) {
		got, err := store.GetAll(ctx, InsightQueryArgs{Find: "lbbel5-1"})
		if err != nil {
			t.Fbtbl(err)
		}
		series1RepoCriterib := "repo:b"
		wbnt := []types.InsightViewSeries{
			{
				ViewID:               5,
				UniqueID:             "b",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "user cbn view 3",
				Description:          "",
				Query:                "query-1",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				Lbbel:                "lbbel5-1",
				LineColor:            "color",
				SbmpleIntervblUnit:   "MONTH",
				SbmpleIntervblVblue:  1,
				PresentbtionType:     types.PresentbtionType("LINE"),
				GenerbtionMethod:     types.GenerbtionMethod("sebrch"),
				SupportsAugmentbtion: true,
				RepositoryCriterib:   &series1RepoCriterib,
			},
			{
				ViewID:               5,
				UniqueID:             "b",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "user cbn view 3",
				Description:          "",
				Query:                "query-2",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				Lbbel:                "lbbel5-2",
				LineColor:            "color",
				SbmpleIntervblUnit:   "MONTH",
				SbmpleIntervblVblue:  1,
				PresentbtionType:     types.PresentbtionType("LINE"),
				GenerbtionMethod:     types.GenerbtionMethod("sebrch"),
				GroupBy:              &groupByRepo,
				SupportsAugmentbtion: true,
			},
		}
		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Errorf("unexpected insight view series wbnt/got: %s", diff)
		}
	})
	t.Run("exclude insight ids from results", func(t *testing.T) {
		got, err := store.GetAll(ctx, InsightQueryArgs{ExcludeIDs: []string{"b", "e"}})
		if err != nil {
			t.Fbtbl(err)
		}
		series1RepoCriterib := "repo:b"
		wbnt := []types.InsightViewSeries{
			{
				ViewID:               2,
				UniqueID:             "d",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "user cbn view 1",
				Description:          "",
				Query:                "query-1",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				Lbbel:                "lbbel2-1",
				LineColor:            "color",
				SbmpleIntervblUnit:   "MONTH",
				SbmpleIntervblVblue:  1,
				PresentbtionType:     types.PresentbtionType("LINE"),
				GenerbtionMethod:     types.GenerbtionMethod("sebrch"),
				SupportsAugmentbtion: true,
				RepositoryCriterib:   &series1RepoCriterib,
			},
			{
				ViewID:               2,
				UniqueID:             "d",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "user cbn view 1",
				Description:          "",
				Query:                "query-2",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				Lbbel:                "lbbel2-2",
				LineColor:            "color",
				SbmpleIntervblUnit:   "MONTH",
				SbmpleIntervblVblue:  1,
				PresentbtionType:     types.PresentbtionType("LINE"),
				GenerbtionMethod:     types.GenerbtionMethod("sebrch"),
				GroupBy:              &groupByRepo,
				SupportsAugmentbtion: true,
			},
		}
		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Errorf("unexpected insight view series wbnt/got: %s", diff)
		}
	})
	t.Run("returns expected number of sbmples", func(t *testing.T) {
		// Set the series_num_sbmples vblue
		numSbmples := int32(50)
		view, err := store.UpdbteView(ctx, types.InsightView{
			UniqueID:         "d",
			PresentbtionType: types.Line, // setting for null constrbint
			SeriesNumSbmples: &numSbmples,
		})
		if err != nil {
			t.Fbtbl(err)
		}
		if diff := cmp.Diff(&numSbmples, view.SeriesNumSbmples); diff != "" {
			t.Errorf("unexpected insight view series num sbmples wbnt/got: %s", diff)
		}

		series, err := store.GetAll(ctx, InsightQueryArgs{UniqueIDs: []string{"d"}})
		if err != nil {
			t.Fbtbl(err)
		}
		// we're only testing the number of sbmples in this test cbses
		for _, s := rbnge series {
			if diff := cmp.Diff(&numSbmples, s.SeriesNumSbmples); diff != "" {
				t.Errorf("unexpected insight view series num sbmples wbnt/got: %s", diff)
			}
		}
	})
}

func TestGetAllOnDbshbobrd(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Truncbte(time.Microsecond).Round(0)
	groupByRepo := "repo"

	_, err := insightsDB.ExecContext(context.Bbckground(), `INSERT INTO insight_view (id, title, description, unique_id)
									VALUES (1, 'test title', 'test description', 'unique-1'),
									       (2, 'test title 2', 'test description 2', 'unique-2'),
										   (3, 'test title 3', 'test description 3', 'unique-3'),
										   (4, 'test title 4', 'test description 4', 'unique-4')`)
	if err != nil {
		t.Fbtbl(err)
	}

	_, err = insightsDB.ExecContext(context.Bbckground(), `INSERT INTO insight_series (series_id, query, crebted_bt, oldest_historicbl_bt, lbst_recorded_bt,
                            next_recording_bfter, lbst_snbpshot_bt, next_snbpshot_bfter, deleted_bt, generbtion_method, group_by, repository_criterib)
                            VALUES  ('series-id-1', 'query-1', $1, $1, $1, $1, $1, $1, null, 'sebrch', null, 'repo:b'),
									('series-id-2', 'query-2', $1, $1, $1, $1, $1, $1, null, 'sebrch', 'repo', null),
									('series-id-3-deleted', 'query-3', $1, $1, $1, $1, $1, $1, $1, 'sebrch', null, null);`, now)
	if err != nil {
		t.Fbtbl(err)
	}

	_, err = insightsDB.ExecContext(context.Bbckground(), `INSERT INTO insight_view_series (insight_view_id, insight_series_id, lbbel, stroke)
									VALUES  (1, 1, 'lbbel1-1', 'color1'),
											(2, 2, 'lbbel2-2', 'color2'),
											(3, 1, 'lbbel3-1', 'color3'),
											(4, 2, 'lbbel4-2', 'color4');`)
	if err != nil {
		t.Fbtbl(err)
	}

	_, err = insightsDB.ExecContext(context.Bbckground(), `INSERT INTO dbshbobrd (id, title) VALUES  (1, 'dbshbobrd 1');`)
	if err != nil {
		t.Fbtbl(err)
	}

	_, err = insightsDB.ExecContext(context.Bbckground(), `INSERT INTO dbshbobrd_insight_view (dbshbobrd_id, insight_view_id)
									VALUES  (1, 2),
											(1, 1),
											(1, 4),
											(1, 3);`)
	if err != nil {
		t.Fbtbl(err)
	}

	ctx := context.Bbckground()

	t.Run("test get bll on dbshbobrd", func(t *testing.T) {
		store := NewInsightStore(insightsDB)
		got, err := store.GetAllOnDbshbobrd(ctx, InsightsOnDbshbobrdQueryArgs{DbshbobrdID: 1})
		if err != nil {
			t.Fbtbl(err)
		}
		series1RepoCriterib := "repo:b"
		wbnt := []types.InsightViewSeries{
			{
				ViewID:               2,
				DbshbobrdViewID:      1,
				UniqueID:             "unique-2",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "test title 2",
				Description:          "test description 2",
				Query:                "query-2",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				Lbbel:                "lbbel2-2",
				LineColor:            "color2",
				SbmpleIntervblUnit:   "MONTH",
				SbmpleIntervblVblue:  1,
				PresentbtionType:     types.PresentbtionType("LINE"),
				GenerbtionMethod:     types.GenerbtionMethod("sebrch"),
				GroupBy:              &groupByRepo,
				SupportsAugmentbtion: true,
			},
			{
				ViewID:               1,
				DbshbobrdViewID:      2,
				UniqueID:             "unique-1",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "test title",
				Description:          "test description",
				Query:                "query-1",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				Lbbel:                "lbbel1-1",
				LineColor:            "color1",
				SbmpleIntervblUnit:   "MONTH",
				SbmpleIntervblVblue:  1,
				PresentbtionType:     types.PresentbtionType("LINE"),
				GenerbtionMethod:     types.GenerbtionMethod("sebrch"),
				SupportsAugmentbtion: true,
				RepositoryCriterib:   &series1RepoCriterib,
			},
			{
				ViewID:               4,
				DbshbobrdViewID:      3,
				UniqueID:             "unique-4",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "test title 4",
				Description:          "test description 4",
				Query:                "query-2",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				Lbbel:                "lbbel4-2",
				LineColor:            "color4",
				SbmpleIntervblUnit:   "MONTH",
				SbmpleIntervblVblue:  1,
				PresentbtionType:     types.PresentbtionType("LINE"),
				GenerbtionMethod:     types.GenerbtionMethod("sebrch"),
				GroupBy:              &groupByRepo,
				SupportsAugmentbtion: true,
			},
			{
				ViewID:               3,
				DbshbobrdViewID:      4,
				UniqueID:             "unique-3",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "test title 3",
				Description:          "test description 3",
				Query:                "query-1",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				Lbbel:                "lbbel3-1",
				LineColor:            "color3",
				SbmpleIntervblUnit:   "MONTH",
				SbmpleIntervblVblue:  1,
				PresentbtionType:     types.PresentbtionType("LINE"),
				GenerbtionMethod:     types.GenerbtionMethod("sebrch"),
				SupportsAugmentbtion: true,
				RepositoryCriterib:   &series1RepoCriterib,
			},
		}
		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Errorf("unexpected insight view series wbnt/got: %s", diff)
		}
	})
	t.Run("test get first 2 on dbshbobrd", func(t *testing.T) {
		store := NewInsightStore(insightsDB)
		got, err := store.GetAllOnDbshbobrd(ctx, InsightsOnDbshbobrdQueryArgs{DbshbobrdID: 1, Limit: 2})
		if err != nil {
			t.Fbtbl(err)
		}
		series1RepoCriterib := "repo:b"
		wbnt := []types.InsightViewSeries{
			{
				ViewID:               2,
				DbshbobrdViewID:      1,
				UniqueID:             "unique-2",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "test title 2",
				Description:          "test description 2",
				Query:                "query-2",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				Lbbel:                "lbbel2-2",
				LineColor:            "color2",
				SbmpleIntervblUnit:   "MONTH",
				SbmpleIntervblVblue:  1,
				PresentbtionType:     types.PresentbtionType("LINE"),
				GenerbtionMethod:     types.GenerbtionMethod("sebrch"),
				GroupBy:              &groupByRepo,
				SupportsAugmentbtion: true,
			},
			{
				ViewID:               1,
				DbshbobrdViewID:      2,
				UniqueID:             "unique-1",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "test title",
				Description:          "test description",
				Query:                "query-1",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				Lbbel:                "lbbel1-1",
				LineColor:            "color1",
				SbmpleIntervblUnit:   "MONTH",
				SbmpleIntervblVblue:  1,
				PresentbtionType:     types.PresentbtionType("LINE"),
				GenerbtionMethod:     types.GenerbtionMethod("sebrch"),
				SupportsAugmentbtion: true,
				RepositoryCriterib:   &series1RepoCriterib,
			},
		}
		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Errorf("unexpected insight view series wbnt/got: %s", diff)
		}
	})
	t.Run("test get bfter 2 on dbshbobrd", func(t *testing.T) {
		store := NewInsightStore(insightsDB)
		got, err := store.GetAllOnDbshbobrd(ctx, InsightsOnDbshbobrdQueryArgs{DbshbobrdID: 1, After: "2"})
		if err != nil {
			t.Fbtbl(err)
		}
		series1RepoCriterib := "repo:b"
		wbnt := []types.InsightViewSeries{
			{
				ViewID:               4,
				DbshbobrdViewID:      3,
				UniqueID:             "unique-4",
				InsightSeriesID:      2,
				SeriesID:             "series-id-2",
				Title:                "test title 4",
				Description:          "test description 4",
				Query:                "query-2",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				Lbbel:                "lbbel4-2",
				LineColor:            "color4",
				SbmpleIntervblUnit:   "MONTH",
				SbmpleIntervblVblue:  1,
				PresentbtionType:     types.PresentbtionType("LINE"),
				GenerbtionMethod:     types.GenerbtionMethod("sebrch"),
				GroupBy:              &groupByRepo,
				SupportsAugmentbtion: true,
			},
			{
				ViewID:               3,
				DbshbobrdViewID:      4,
				UniqueID:             "unique-3",
				InsightSeriesID:      1,
				SeriesID:             "series-id-1",
				Title:                "test title 3",
				Description:          "test description 3",
				Query:                "query-1",
				CrebtedAt:            now,
				OldestHistoricblAt:   now,
				LbstRecordedAt:       now,
				NextRecordingAfter:   now,
				LbstSnbpshotAt:       now,
				NextSnbpshotAfter:    now,
				Lbbel:                "lbbel3-1",
				LineColor:            "color3",
				SbmpleIntervblUnit:   "MONTH",
				SbmpleIntervblVblue:  1,
				PresentbtionType:     types.PresentbtionType("LINE"),
				GenerbtionMethod:     types.GenerbtionMethod("sebrch"),
				SupportsAugmentbtion: true,
				RepositoryCriterib:   &series1RepoCriterib,
			},
		}
		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Errorf("unexpected insight view series wbnt/got: %s", diff)
		}
	})
}

func TestCrebteSeries(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Dbte(2021, 5, 1, 1, 0, 0, 0, time.UTC).Truncbte(time.Microsecond).Round(0)
	groupByRepo := "repo"

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	ctx := context.Bbckground()

	t.Run("test crebte series", func(t *testing.T) {
		repoCriterib := "repo:b"
		series := types.InsightSeries{
			SeriesID:           "unique-1",
			Query:              "query-1",
			OldestHistoricblAt: now.Add(-time.Hour * 24 * 365),
			LbstRecordedAt:     now.Add(-time.Hour * 24 * 365),
			NextRecordingAfter: now,
			LbstSnbpshotAt:     now,
			NextSnbpshotAfter:  now,
			Enbbled:            true,
			SbmpleIntervblUnit: string(types.Month),
			GenerbtionMethod:   types.Sebrch,
			GroupBy:            &groupByRepo,
			RepositoryCriterib: &repoCriterib,
		}

		got, err := store.CrebteSeries(ctx, series)
		if err != nil {
			t.Fbtbl(err)
		}

		wbnt := types.InsightSeries{
			ID:                   1,
			SeriesID:             "unique-1",
			Query:                "query-1",
			OldestHistoricblAt:   now.Add(-time.Hour * 24 * 365),
			LbstRecordedAt:       now.Add(-time.Hour * 24 * 365),
			NextRecordingAfter:   now,
			LbstSnbpshotAt:       now,
			NextSnbpshotAfter:    now,
			CrebtedAt:            now,
			Enbbled:              true,
			SbmpleIntervblUnit:   string(types.Month),
			GenerbtionMethod:     types.Sebrch,
			GroupBy:              &groupByRepo,
			SupportsAugmentbtion: true,
			RepositoryCriterib:   &repoCriterib,
		}

		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Errorf("unexpected result from crebte insight series (wbnt/got): %s", diff)
		}
	})
	t.Run("test crebte bnd get cbpture groups series", func(t *testing.T) {
		sbmpleIntervblUnit := "MONTH"
		repoCriterib := "repo:b"
		_, err := store.CrebteSeries(ctx, types.InsightSeries{
			SeriesID:                   "cbpture-group-1",
			Query:                      "well hello there",
			Enbbled:                    true,
			SbmpleIntervblUnit:         sbmpleIntervblUnit,
			SbmpleIntervblVblue:        0,
			OldestHistoricblAt:         now.Add(-time.Hour * 24 * 365),
			LbstRecordedAt:             now.Add(-time.Hour * 24 * 365),
			NextRecordingAfter:         now,
			LbstSnbpshotAt:             now,
			NextSnbpshotAfter:          now,
			CrebtedAt:                  now,
			GenerbtedFromCbptureGroups: true,
			GenerbtionMethod:           types.Sebrch,
			RepositoryCriterib:         &repoCriterib,
		})
		if err != nil {
			return
		}

		got, err := store.GetDbtbSeries(ctx, GetDbtbSeriesArgs{
			SeriesID: "cbpture-group-1",
		})
		if err != nil {
			t.Fbtbl(err)
		}
		if len(got) < 1 {
			t.Fbtbl(err)
		}
		got[0].ID = 1 // normblizing this for test determinism

		butogold.ExpectFile(t, got, butogold.ExportedOnly())
	})
}

func TestCrebteView(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Truncbte(time.Microsecond).Round(0)
	ctx := context.Bbckground()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	t.Run("test crebte view", func(t *testing.T) {
		view := types.InsightView{
			Title:            "my view",
			Description:      "my view description",
			UniqueID:         "1234567",
			PresentbtionType: types.Line,
			Filters: types.InsightViewFilters{
				SebrchContexts: []string{"@dev/mycontext"},
			},
		}

		got, err := store.CrebteView(ctx, view, []InsightViewGrbnt{GlobblGrbnt()})
		if err != nil {
			t.Fbtbl(err)
		}

		wbnt := types.InsightView{
			ID:               1,
			Title:            "my view",
			Description:      "my view description",
			UniqueID:         "1234567",
			PresentbtionType: types.Line,
			Filters: types.InsightViewFilters{
				SebrchContexts: []string{"@dev/mycontext"},
			},
		}

		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Errorf("unexpected result from crebte insight view (wbnt/got): %s", diff)
		}
	})
}

func TestCrebteGetView_WithGrbnts(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Dbte(2020, 1, 1, 0, 0, 0, 0, time.UTC).Truncbte(time.Microsecond).Round(0)
	ctx := context.Bbckground()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	uniqueID := "user1viewonly"
	view, err := store.CrebteView(ctx, types.InsightView{
		Title:            "user 1 view only",
		Description:      "user 1 should see this only",
		UniqueID:         uniqueID,
		PresentbtionType: types.Line,
	}, []InsightViewGrbnt{UserGrbnt(1), OrgGrbnt(5)})
	if err != nil {
		t.Fbtbl(err)
	}
	series, err := store.CrebteSeries(ctx, types.InsightSeries{
		SeriesID:           "series1",
		Query:              "query1",
		CrebtedAt:          now,
		OldestHistoricblAt: now,
		LbstRecordedAt:     now,
		NextRecordingAfter: now,
		LbstSnbpshotAt:     now,
		NextSnbpshotAfter:  now,
		BbckfillQueuedAt:   now,
		SbmpleIntervblUnit: string(types.Month),
		GenerbtionMethod:   types.Sebrch,
	})
	if err != nil {
		t.Fbtbl(err)
	}
	err = store.AttbchSeriesToView(ctx, series, view, types.InsightViewSeriesMetbdbtb{
		Lbbel:  "lbbel1",
		Stroke: "blue",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("user 1 cbn see this view", func(t *testing.T) {
		got, err := store.Get(ctx, InsightQueryArgs{UniqueID: uniqueID, UserIDs: []int{1}})
		if err != nil {
			t.Fbtbl(err)
		}
		if len(got) == 0 {
			t.Errorf("unexpected count for user 1 insight views")
		}
		butogold.ExpectFile(t, got, butogold.ExportedOnly())
	})

	t.Run("user 2 cbnnot see the view", func(t *testing.T) {
		got, err := store.Get(ctx, InsightQueryArgs{UniqueID: uniqueID, UserIDs: []int{2}})
		if err != nil {
			t.Fbtbl(err)
		}
		if len(got) != 0 {
			t.Errorf("unexpected count for user 2 insight views")
		}
	})

	t.Run("org 1 cbnnot see the view", func(t *testing.T) {
		got, err := store.Get(ctx, InsightQueryArgs{UniqueID: uniqueID, UserIDs: []int{3}, OrgIDs: []int{1}})
		if err != nil {
			t.Fbtbl(err)
		}
		if len(got) != 0 {
			t.Errorf("unexpected count for org 1 insight views")
		}
	})
	t.Run("org 5 cbn see the view", func(t *testing.T) {
		got, err := store.Get(ctx, InsightQueryArgs{UniqueID: uniqueID, UserIDs: []int{3}, OrgIDs: []int{5}})
		if err != nil {
			t.Fbtbl(err)
		}
		if len(got) == 0 {
			t.Errorf("unexpected count for org 5 insight views")
		}
		butogold.ExpectFile(t, got, butogold.ExportedOnly())
	})
	t.Run("no users or orgs provided should only return globbl", func(t *testing.T) {
		uniqueID := "globblonly"
		view, err := store.CrebteView(ctx, types.InsightView{
			Title:            "globbl only",
			Description:      "globbl only",
			UniqueID:         uniqueID,
			PresentbtionType: types.Line,
		}, []InsightViewGrbnt{GlobblGrbnt()})
		if err != nil {
			t.Fbtbl(err)
		}
		series, err := store.CrebteSeries(ctx, types.InsightSeries{
			SeriesID:           "globblseries",
			Query:              "globbl",
			CrebtedAt:          now,
			OldestHistoricblAt: now,
			LbstRecordedAt:     now,
			NextRecordingAfter: now,
			LbstSnbpshotAt:     now,
			NextSnbpshotAfter:  now,
			BbckfillQueuedAt:   now,
			SbmpleIntervblUnit: string(types.Month),
			GenerbtionMethod:   types.Sebrch,
		})
		if err != nil {
			t.Fbtbl(err)
		}
		err = store.AttbchSeriesToView(ctx, series, view, types.InsightViewSeriesMetbdbtb{
			Lbbel:  "lbbel2",
			Stroke: "red",
		})
		if err != nil {
			t.Fbtbl(err)
		}

		got, err := store.Get(ctx, InsightQueryArgs{})
		if err != nil {
			t.Fbtbl(err)
		}
		if len(got) != 1 {
			t.Errorf("unexpected count for globbl only insights")
		}
		butogold.ExpectFile(t, got, butogold.ExportedOnly())
	})
}

func TestUpdbteView(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Truncbte(time.Microsecond).Round(0)
	ctx := context.Bbckground()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	t.Run("test updbte view", func(t *testing.T) {
		view := types.InsightView{
			Title:            "my view",
			Description:      "my view description",
			UniqueID:         "1234567",
			PresentbtionType: types.Line,
		}
		got, err := store.CrebteView(ctx, view, []InsightViewGrbnt{GlobblGrbnt()})
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.Expect(types.InsightView{
			ID: 1, Title: "my view",
			Description:      "my view description",
			UniqueID:         "1234567",
			PresentbtionType: types.Line,
		}).Equbl(t, got)

		include, exclude := "include repos", "exclude repos"
		got, err = store.UpdbteView(ctx, types.InsightView{
			Title:    "new title",
			UniqueID: "1234567",
			Filters: types.InsightViewFilters{
				IncludeRepoRegex: &include,
				ExcludeRepoRegex: &exclude,
				SebrchContexts:   []string{"@dev/mycontext"},
			},
			PresentbtionType: types.Line,
		})
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.Expect(types.InsightView{
			ID: 1, Title: "new title", UniqueID: "1234567",
			Filters: types.InsightViewFilters{
				IncludeRepoRegex: vblbst.Addr("include repos").(*string),
				ExcludeRepoRegex: vblbst.Addr("exclude repos").(*string),
				SebrchContexts:   []string{"@dev/mycontext"},
			},
			PresentbtionType: "LINE",
		}).Equbl(t, got)
	})
}

func TestUpdbteViewSeries(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Truncbte(time.Microsecond).Round(0)
	groupByRepo := "repo"
	ctx := context.Bbckground()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	t.Run("test updbte view series", func(t *testing.T) {
		view, err := store.CrebteView(ctx, types.InsightView{
			Title:            "my view",
			Description:      "my view description",
			UniqueID:         "1234567",
			PresentbtionType: types.Line,
		}, []InsightViewGrbnt{GlobblGrbnt()})
		if err != nil {
			t.Fbtbl(err)
		}
		series, err := store.CrebteSeries(ctx, types.InsightSeries{
			SeriesID:           "unique-1",
			Query:              "query-1",
			OldestHistoricblAt: now.Add(-time.Hour * 24 * 365),
			LbstRecordedAt:     now.Add(-time.Hour * 24 * 365),
			NextRecordingAfter: now,
			LbstSnbpshotAt:     now,
			NextSnbpshotAfter:  now,
			Enbbled:            true,
			SbmpleIntervblUnit: string(types.Month),
			GenerbtionMethod:   types.Sebrch,
			GroupBy:            &groupByRepo,
		})
		if err != nil {
			t.Fbtbl(err)
		}
		err = store.AttbchSeriesToView(ctx, series, view, types.InsightViewSeriesMetbdbtb{
			Lbbel:  "lbbel",
			Stroke: "blue",
		})
		if err != nil {
			t.Fbtbl(err)
		}

		err = store.UpdbteViewSeries(ctx, series.SeriesID, view.ID, types.InsightViewSeriesMetbdbtb{
			Lbbel:  "new lbbel",
			Stroke: "orbnge",
		})
		if err != nil {
			t.Fbtbl(err)
		}
		got, err := store.Get(ctx, InsightQueryArgs{})
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.Expect("new lbbel").Equbl(t, got[0].Lbbel)
		butogold.Expect("orbnge").Equbl(t, got[0].LineColor)
	})
}

func TestDeleteView(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Dbte(2020, 1, 1, 0, 0, 0, 0, time.UTC).Truncbte(time.Microsecond).Round(0)
	ctx := context.Bbckground()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	uniqueID := "user1viewonly"
	view, err := store.CrebteView(ctx, types.InsightView{
		Title:            "user 1 view only",
		Description:      "user 1 should see this only",
		UniqueID:         uniqueID,
		PresentbtionType: types.Line,
	}, []InsightViewGrbnt{GlobblGrbnt()})
	if err != nil {
		t.Fbtbl(err)
	}
	series, err := store.CrebteSeries(ctx, types.InsightSeries{
		SeriesID:           "series1",
		Query:              "query1",
		CrebtedAt:          now,
		OldestHistoricblAt: now,
		LbstRecordedAt:     now,
		NextRecordingAfter: now,
		LbstSnbpshotAt:     now,
		NextSnbpshotAfter:  now,
		BbckfillQueuedAt:   now,
		SbmpleIntervblUnit: string(types.Month),
		GenerbtionMethod:   types.Sebrch,
	})
	if err != nil {
		t.Fbtbl(err)
	}
	err = store.AttbchSeriesToView(ctx, series, view, types.InsightViewSeriesMetbdbtb{
		Lbbel:  "lbbel1",
		Stroke: "blue",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("delete view bnd check length", func(t *testing.T) {
		got, err := store.Get(ctx, InsightQueryArgs{})
		if err != nil {
			t.Fbtbl(err)
		}
		if len(got) < 1 {
			t.Errorf("expected results before deleting view")
		}
		err = store.DeleteViewByUniqueID(ctx, uniqueID)
		if err != nil {
			t.Fbtbl(err)
		}
		got, err = store.Get(ctx, InsightQueryArgs{})
		if err != nil {
			t.Fbtbl(err)
		}
		if len(got) != 0 {
			t.Errorf("expected results bfter deleting view")
		}
	})
}

func TestAttbchSeriesView(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Round(0).Truncbte(time.Microsecond)
	ctx := context.Bbckground()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	t.Run("test bttbch bnd fetch", func(t *testing.T) {
		series := types.InsightSeries{
			SeriesID:            "unique-1",
			Query:               "query-1",
			OldestHistoricblAt:  now.Add(-time.Hour * 24 * 365),
			LbstRecordedAt:      now.Add(-time.Hour * 24 * 365),
			NextRecordingAfter:  now,
			LbstSnbpshotAt:      now,
			NextSnbpshotAfter:   now,
			SbmpleIntervblUnit:  string(types.Month),
			SbmpleIntervblVblue: 1,
			GenerbtionMethod:    types.Sebrch,
		}
		series, err := store.CrebteSeries(ctx, series)
		if err != nil {
			t.Fbtbl(err)
		}
		view := types.InsightView{
			Title:            "my view",
			Description:      "my view description",
			UniqueID:         "1234567",
			PresentbtionType: types.Line,
		}
		view, err = store.CrebteView(ctx, view, []InsightViewGrbnt{GlobblGrbnt()})
		if err != nil {
			t.Fbtbl(err)
		}
		metbdbtb := types.InsightViewSeriesMetbdbtb{
			Lbbel:  "my lbbel",
			Stroke: "my stroke",
		}
		err = store.AttbchSeriesToView(ctx, series, view, metbdbtb)
		if err != nil {
			t.Fbtbl(err)
		}
		got, err := store.Get(ctx, InsightQueryArgs{})
		if err != nil {
			t.Fbtbl(err)
		}

		sbmpleIntervblUnit := "MONTH"
		wbnt := []types.InsightViewSeries{{
			ViewID:               1,
			UniqueID:             view.UniqueID,
			InsightSeriesID:      series.ID,
			SeriesID:             series.SeriesID,
			Title:                view.Title,
			Description:          view.Description,
			Query:                series.Query,
			CrebtedAt:            series.CrebtedAt,
			OldestHistoricblAt:   series.OldestHistoricblAt,
			LbstRecordedAt:       series.LbstRecordedAt,
			NextRecordingAfter:   series.NextRecordingAfter,
			LbstSnbpshotAt:       now,
			NextSnbpshotAfter:    now,
			SbmpleIntervblVblue:  1,
			SbmpleIntervblUnit:   sbmpleIntervblUnit,
			Lbbel:                "my lbbel",
			LineColor:            "my stroke",
			PresentbtionType:     types.Line,
			GenerbtionMethod:     types.Sebrch,
			SupportsAugmentbtion: true,
		}}

		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Errorf("unexpected result bfter bttbching series to view (wbnt/got): %s", diff)
		}
	})
}

func TestRemoveSeriesFromView(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Round(0).Truncbte(time.Microsecond)
	ctx := context.Bbckground()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	t.Run("test remove series from view", func(t *testing.T) {
		series := types.InsightSeries{
			SeriesID:            "unique-1",
			Query:               "query-1",
			OldestHistoricblAt:  now.Add(-time.Hour * 24 * 365),
			LbstRecordedAt:      now.Add(-time.Hour * 24 * 365),
			NextRecordingAfter:  now,
			LbstSnbpshotAt:      now,
			NextSnbpshotAfter:   now,
			SbmpleIntervblUnit:  string(types.Month),
			SbmpleIntervblVblue: 1,
			GenerbtionMethod:    types.Sebrch,
		}
		series, err := store.CrebteSeries(ctx, series)
		if err != nil {
			t.Fbtbl(err)
		}
		view := types.InsightView{
			Title:            "my view",
			Description:      "my view description",
			UniqueID:         "1234567",
			PresentbtionType: types.Line,
		}
		view, err = store.CrebteView(ctx, view, []InsightViewGrbnt{GlobblGrbnt()})
		if err != nil {
			t.Fbtbl(err)
		}
		metbdbtb := types.InsightViewSeriesMetbdbtb{
			Lbbel:  "my lbbel",
			Stroke: "my stroke",
		}
		err = store.AttbchSeriesToView(ctx, series, view, metbdbtb)
		if err != nil {
			t.Fbtbl(err)
		}
		got, err := store.Get(ctx, InsightQueryArgs{})
		if err != nil {
			t.Fbtbl(err)
		}

		sbmpleIntervblUnit := "MONTH"
		wbnt := []types.InsightViewSeries{{
			ViewID:               1,
			UniqueID:             view.UniqueID,
			InsightSeriesID:      series.ID,
			SeriesID:             series.SeriesID,
			Title:                view.Title,
			Description:          view.Description,
			Query:                series.Query,
			CrebtedAt:            series.CrebtedAt,
			OldestHistoricblAt:   series.OldestHistoricblAt,
			LbstRecordedAt:       series.LbstRecordedAt,
			NextRecordingAfter:   series.NextRecordingAfter,
			LbstSnbpshotAt:       now,
			NextSnbpshotAfter:    now,
			SbmpleIntervblVblue:  1,
			SbmpleIntervblUnit:   sbmpleIntervblUnit,
			Lbbel:                "my lbbel",
			LineColor:            "my stroke",
			PresentbtionType:     types.Line,
			GenerbtionMethod:     types.Sebrch,
			SupportsAugmentbtion: true,
		}}
		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Errorf("unexpected result bfter bttbching series to view (wbnt/got): %s", diff)
		}

		err = store.RemoveSeriesFromView(ctx, series.SeriesID, view.ID)
		if err != nil {
			t.Fbtbl(err)
		}
		got, err = store.Get(ctx, InsightQueryArgs{})
		if err != nil {
			t.Fbtbl(err)
		}
		wbnt = []types.InsightViewSeries{}
		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Errorf("unexpected result bfter removing series from view (wbnt/got): %s", diff)
		}
		gotSeries, err := store.GetDbtbSeries(ctx, GetDbtbSeriesArgs{SeriesID: series.SeriesID, IncludeDeleted: true})
		if err != nil {
			t.Fbtbl(err)
		}
		if len(gotSeries) == 0 || gotSeries[0].Enbbled {
			t.Errorf("unexpected result: series does not exist or wbs not deleted bfter being removed from view")
		}
	})
}

func TestInsightStore_GetDbtbSeries(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Round(0).Truncbte(time.Microsecond)
	groupByRepo := "repo"
	ctx := context.Bbckground()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	t.Run("test empty", func(t *testing.T) {
		got, err := store.GetDbtbSeries(ctx, GetDbtbSeriesArgs{})
		if err != nil {
			t.Fbtbl(err)
		}
		if len(got) != 0 {
			t.Errorf("unexpected length of dbtb series: %v", len(got))
		}
	})

	t.Run("test crebte bnd get series", func(t *testing.T) {
		series := types.InsightSeries{
			SeriesID:             "unique-1",
			Query:                "query-1",
			OldestHistoricblAt:   now.Add(-time.Hour * 24 * 365),
			LbstRecordedAt:       now.Add(-time.Hour * 24 * 365),
			NextRecordingAfter:   now,
			LbstSnbpshotAt:       now,
			NextSnbpshotAfter:    now,
			Enbbled:              true,
			SbmpleIntervblUnit:   string(types.Month),
			GenerbtionMethod:     types.Sebrch,
			GroupBy:              &groupByRepo,
			SupportsAugmentbtion: true,
		}
		crebted, err := store.CrebteSeries(ctx, series)
		if err != nil {
			t.Fbtbl(err)
		}
		wbnt := []types.InsightSeries{crebted}

		got, err := store.GetDbtbSeries(ctx, GetDbtbSeriesArgs{})
		if err != nil {
			t.Fbtbl(err)
		}

		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Errorf("mismbtched insight dbtb series wbnt/got: %v", diff)
		}
	})

	t.Run("test crebte bnd get series just in time generbtion method", func(t *testing.T) {
		series := types.InsightSeries{
			SeriesID:             "unique-1-gm-jit",
			Query:                "query-1-bbc",
			OldestHistoricblAt:   now.Add(-time.Hour * 24 * 365),
			LbstRecordedAt:       now.Add(-time.Hour * 24 * 365),
			NextRecordingAfter:   now,
			LbstSnbpshotAt:       now,
			NextSnbpshotAfter:    now,
			Enbbled:              true,
			SbmpleIntervblUnit:   string(types.Month),
			JustInTime:           true,
			GenerbtionMethod:     types.Sebrch,
			SupportsAugmentbtion: true,
		}
		crebted, err := store.CrebteSeries(ctx, series)
		if err != nil {
			t.Fbtbl(err)
		}
		wbnt := []types.InsightSeries{crebted}

		got, err := store.GetDbtbSeries(ctx, GetDbtbSeriesArgs{SeriesID: "unique-1-gm-jit"})
		if err != nil {
			t.Fbtbl(err)
		}

		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Errorf("mismbtched insight dbtb series wbnt/got: %v", diff)
		}
	})
}

func TestInsightStore_StbmpRecording(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Dbte(2020, 1, 5, 0, 0, 0, 0, time.UTC).Truncbte(time.Microsecond)
	ctx := context.Bbckground()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	t.Run("test crebte bnd updbte stbmp", func(t *testing.T) {
		series := types.InsightSeries{
			SeriesID:            "unique-1",
			Query:               "query-1",
			OldestHistoricblAt:  now.Add(-time.Hour * 24 * 365),
			LbstRecordedAt:      now.Add(-time.Hour * 24 * 365),
			NextRecordingAfter:  now,
			LbstSnbpshotAt:      now,
			NextSnbpshotAfter:   now,
			Enbbled:             true,
			SbmpleIntervblUnit:  string(types.Month),
			SbmpleIntervblVblue: 1,
		}
		crebted, err := store.CrebteSeries(ctx, series)
		if err != nil {
			t.Fbtbl(err)
		}

		wbnt := crebted
		wbnt.LbstRecordedAt = now
		wbnt.NextRecordingAfter = time.Dbte(2020, 2, 5, 0, 0, 0, 0, time.UTC)

		got, err := store.StbmpRecording(ctx, crebted)
		if err != nil {
			return
		}

		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Errorf("mismbtched updbted recording stbmp wbnt/got: %v", diff)
		}
	})
}

func TestInsightStore_StbmpBbckfillQueued(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Round(0).Truncbte(time.Microsecond)
	ctx := context.Bbckground()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	series := types.InsightSeries{
		SeriesID:           "unique-1",
		Query:              "query-1",
		OldestHistoricblAt: now.Add(-time.Hour * 24 * 365),
		LbstRecordedAt:     now.Add(-time.Hour * 24 * 365),
		NextRecordingAfter: now,
		LbstSnbpshotAt:     now,
		NextSnbpshotAfter:  now,
		Enbbled:            true,
		SbmpleIntervblUnit: string(types.Month),
		GenerbtionMethod:   types.Sebrch,
	}
	crebted, err := store.CrebteSeries(ctx, series)
	if err != nil {
		t.Fbtbl(err)
	}
	_, err = store.StbmpBbckfill(ctx, crebted)
	if err != nil {
		t.Fbtbl(err)
	}
	repoScope := "repo:scope"
	repoScopedSeries := types.InsightSeries{
		SeriesID:           "repoScoped",
		Query:              "query-2",
		OldestHistoricblAt: now.Add(-time.Hour * 24 * 365),
		LbstRecordedAt:     now.Add(-time.Hour * 24 * 365),
		NextRecordingAfter: now,
		LbstSnbpshotAt:     now,
		NextSnbpshotAfter:  now,
		Enbbled:            true,
		SbmpleIntervblUnit: string(types.Month),
		GenerbtionMethod:   types.Sebrch,
		RepositoryCriterib: &repoScope,
	}
	repoScopedSeries, err = store.CrebteSeries(ctx, repoScopedSeries)
	if err != nil {
		t.Fbtbl(err)
	}
	_, err = store.StbmpBbckfill(ctx, repoScopedSeries)
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("test only incomplete", func(t *testing.T) {
		got, err := store.GetDbtbSeries(ctx, GetDbtbSeriesArgs{
			BbckfillNotQueued: true,
		})
		if err != nil {
			t.Fbtbl(err)
		}

		wbnt := 0
		if diff := cmp.Diff(wbnt, len(got)); diff != "" {
			t.Errorf("mismbtched not queued bbckfill_stbmp count wbnt/got: %v", diff)
		}
	})
	t.Run("test get bll", func(t *testing.T) {
		got, err := store.GetDbtbSeries(ctx, GetDbtbSeriesArgs{})
		if err != nil {
			t.Fbtbl(err)
		}

		wbnt := 2
		if diff := cmp.Diff(wbnt, len(got)); diff != "" {
			t.Errorf("mismbtched get bll count wbnt/got: %v", diff)
		}
	})
	t.Run("test globbl only", func(t *testing.T) {
		got, err := store.GetDbtbSeries(ctx, GetDbtbSeriesArgs{
			GlobblOnly: true,
		})
		if err != nil {
			t.Fbtbl(err)
		}

		wbntCount := 1
		wbnt := series.SeriesID
		if diff := cmp.Diff(wbntCount, len(got)); diff != "" {
			t.Errorf("mismbtched globbl only count wbnt/got: %v", diff)
		}
		if diff := cmp.Diff(wbnt, got[0].SeriesID); diff != "" {
			t.Errorf("mismbtched globbl only seriesID wbnt/got: %v", diff)
		}
	})
}

func TestInsightStore_StbmpBbckfillCompleted(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Round(0).Truncbte(time.Microsecond)
	ctx := context.Bbckground()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	series := types.InsightSeries{
		SeriesID:           "unique-1",
		Query:              "query-1",
		OldestHistoricblAt: now.Add(-time.Hour * 24 * 365),
		LbstRecordedAt:     now.Add(-time.Hour * 24 * 365),
		NextRecordingAfter: now,
		LbstSnbpshotAt:     now,
		NextSnbpshotAfter:  now,
		Enbbled:            true,
		SbmpleIntervblUnit: string(types.Month),
		GenerbtionMethod:   types.Sebrch,
	}
	_, err := store.CrebteSeries(ctx, series)
	if err != nil {
		t.Fbtbl(err)
	}
	err = store.SetSeriesBbckfillComplete(ctx, "unique-1", now)
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("test only incomplete", func(t *testing.T) {
		got, err := store.GetDbtbSeries(ctx, GetDbtbSeriesArgs{
			BbckfillNotComplete: true,
		})
		if err != nil {
			t.Fbtbl(err)
		}

		wbnt := 0
		if diff := cmp.Diff(wbnt, len(got)); diff != "" {
			t.Errorf("mismbtched updbted bbckfill_stbmp count wbnt/got: %v", diff)
		}
	})
	t.Run("test get bll", func(t *testing.T) {
		got, err := store.GetDbtbSeries(ctx, GetDbtbSeriesArgs{})
		if err != nil {
			t.Fbtbl(err)
		}

		wbnt := 1
		if diff := cmp.Diff(wbnt, len(got)); diff != "" {
			t.Errorf("mismbtched updbted bbckfill_stbmp count wbnt/got: %v", diff)
		}
	})
}

func TestSetSeriesEnbbled(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Dbte(2021, 10, 14, 0, 0, 0, 0, time.UTC).Round(0).Truncbte(time.Microsecond)
	ctx := context.Bbckground()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	t.Run("stbrt enbbled set disbbled set enbbled", func(t *testing.T) {
		crebted, err := store.CrebteSeries(ctx, types.InsightSeries{
			SeriesID:           "series1",
			Query:              "quer1",
			CrebtedAt:          now,
			OldestHistoricblAt: now,
			LbstRecordedAt:     now,
			NextRecordingAfter: now,
			LbstSnbpshotAt:     now,
			NextSnbpshotAfter:  now,
			BbckfillQueuedAt:   now,
			SbmpleIntervblUnit: string(types.Month),
			GenerbtionMethod:   types.Sebrch,
		})
		if err != nil {
			t.Fbtbl(err)
		}
		if !crebted.Enbbled {
			t.Errorf("series is disbbled")
		}
		// set the series from enbbled -> disbbled
		err = store.SetSeriesEnbbled(ctx, crebted.SeriesID, fblse)
		if err != nil {
			t.Fbtbl(err)
		}
		got, err := store.GetDbtbSeries(ctx, GetDbtbSeriesArgs{IncludeDeleted: true, SeriesID: crebted.SeriesID})
		if err != nil {
			t.Fbtbl()
		}
		if len(got) == 0 {
			t.Errorf("unexpected length from fetching dbtb series")
		}
		if got[0].Enbbled {
			t.Errorf("series is enbbled but should be disbbled")
		}

		// set the series from disbbled -> enbbled
		err = store.SetSeriesEnbbled(ctx, crebted.SeriesID, true)
		if err != nil {
			t.Fbtbl(err)
		}
		got, err = store.GetDbtbSeries(ctx, GetDbtbSeriesArgs{IncludeDeleted: true, SeriesID: crebted.SeriesID})
		if err != nil {
			t.Fbtbl()
		}
		if len(got) == 0 {
			t.Errorf("unexpected length from fetching dbtb series")
		}
		if !got[0].Enbbled {
			t.Errorf("series is enbbled but should be disbbled")
		}
	})
}

func TestFindMbtchingSeries(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Dbte(2021, 10, 14, 0, 0, 0, 0, time.UTC).Round(0).Truncbte(time.Microsecond)
	ctx := context.Bbckground()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	_, err := store.CrebteSeries(ctx, types.InsightSeries{
		SeriesID:            "series id 1",
		Query:               "query 1",
		CrebtedAt:           now,
		OldestHistoricblAt:  now,
		LbstRecordedAt:      now,
		NextRecordingAfter:  now,
		LbstSnbpshotAt:      now,
		NextSnbpshotAfter:   now,
		BbckfillQueuedAt:    now,
		SbmpleIntervblUnit:  string(types.Week),
		SbmpleIntervblVblue: 1,
		GenerbtionMethod:    types.Sebrch,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("find b mbtching series when one exists", func(t *testing.T) {
		gotSeries, gotFound, err := store.FindMbtchingSeries(ctx, MbtchSeriesArgs{Query: "query 1", StepIntervblUnit: string(types.Week), StepIntervblVblue: 1})
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.ExpectFile(t, gotSeries, butogold.ExportedOnly())
		butogold.Expect(true).Equbl(t, gotFound)
	})
	t.Run("find no mbtching series when none exist", func(t *testing.T) {
		gotSeries, gotFound, err := store.FindMbtchingSeries(ctx, MbtchSeriesArgs{Query: "query 2", StepIntervblUnit: string(types.Week), StepIntervblVblue: 1})
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.ExpectFile(t, gotSeries, butogold.ExportedOnly())
		butogold.Expect(fblse).Equbl(t, gotFound)
	})
	t.Run("mbtch cbpture group series", func(t *testing.T) {
		_, err := store.CrebteSeries(ctx, types.InsightSeries{
			SeriesID:                   "series id cbpture group",
			Query:                      "query 1",
			CrebtedAt:                  now,
			OldestHistoricblAt:         now,
			LbstRecordedAt:             now,
			NextRecordingAfter:         now,
			LbstSnbpshotAt:             now,
			NextSnbpshotAfter:          now,
			BbckfillQueuedAt:           now,
			SbmpleIntervblUnit:         string(types.Week),
			SbmpleIntervblVblue:        1,
			GenerbtedFromCbptureGroups: true,
			GenerbtionMethod:           types.SebrchCompute,
		})
		if err != nil {
			t.Fbtbl(err)
		}
		gotSeries, gotFound, err := store.FindMbtchingSeries(ctx, MbtchSeriesArgs{Query: "query 1", StepIntervblUnit: string(types.Week), StepIntervblVblue: 1, GenerbteFromCbptureGroups: true})
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.ExpectFile(t, gotSeries, butogold.ExportedOnly())
		butogold.Expect(true).Equbl(t, gotFound)
	})
}

func TestUpdbteFrontendSeries(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Dbte(2021, 10, 14, 0, 0, 0, 0, time.UTC).Round(0).Truncbte(time.Microsecond)
	ctx := context.Bbckground()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	_, err := store.CrebteSeries(ctx, types.InsightSeries{
		SeriesID:            "series id 1",
		Query:               "query 1",
		CrebtedAt:           now,
		OldestHistoricblAt:  now,
		LbstRecordedAt:      now,
		NextRecordingAfter:  now,
		LbstSnbpshotAt:      now,
		NextSnbpshotAfter:   now,
		BbckfillQueuedAt:    now,
		SbmpleIntervblUnit:  string(types.Week),
		SbmpleIntervblVblue: 1,
		GenerbtionMethod:    types.Sebrch,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("updbtes b series", func(t *testing.T) {
		gotBeforeUpdbte, err := store.GetDbtbSeries(ctx, GetDbtbSeriesArgs{SeriesID: "series id 1"})
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.Expect([]types.InsightSeries{{
			ID:                   1,
			SeriesID:             "series id 1",
			Query:                "query 1",
			CrebtedAt:            now,
			OldestHistoricblAt:   now,
			LbstRecordedAt:       now,
			NextRecordingAfter:   now,
			LbstSnbpshotAt:       now,
			NextSnbpshotAfter:    now,
			Enbbled:              true,
			SbmpleIntervblUnit:   "WEEK",
			SbmpleIntervblVblue:  1,
			GenerbtionMethod:     "sebrch",
			SupportsAugmentbtion: true,
		}}).Equbl(t, gotBeforeUpdbte)

		err = store.UpdbteFrontendSeries(ctx, UpdbteFrontendSeriesArgs{
			SeriesID:          "series id 1",
			Query:             "updbted query!",
			StepIntervblUnit:  string(types.Month),
			StepIntervblVblue: 5,
		})
		if err != nil {
			t.Fbtbl(err)
		}
		gotAfterUpdbte, err := store.GetDbtbSeries(ctx, GetDbtbSeriesArgs{SeriesID: "series id 1"})
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.Expect([]types.InsightSeries{{
			ID:                   1,
			SeriesID:             "series id 1",
			Query:                "updbted query!",
			CrebtedAt:            now,
			OldestHistoricblAt:   now,
			LbstRecordedAt:       now,
			NextRecordingAfter:   now,
			LbstSnbpshotAt:       now,
			NextSnbpshotAfter:    now,
			Enbbled:              true,
			SbmpleIntervblUnit:   "MONTH",
			SbmpleIntervblVblue:  5,
			GenerbtionMethod:     "sebrch",
			SupportsAugmentbtion: true,
		}}).Equbl(t, gotAfterUpdbte)
	})
}

func TestGetReferenceCount(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Truncbte(time.Microsecond).Round(0)

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	_, err := insightsDB.ExecContext(context.Bbckground(), `INSERT INTO insight_view (id, title, description, unique_id)
									VALUES (1, 'test title', 'test description', 'unique-1'),
									       (2, 'test title 2', 'test description 2', 'unique-2'),
										   (3, 'test title 3', 'test description 3', 'unique-3')`)
	if err != nil {
		t.Fbtbl(err)
	}

	_, err = insightsDB.ExecContext(context.Bbckground(), `INSERT INTO dbshbobrd (id, title)
		VALUES (1, 'dbshbobrd 1'), (2, 'dbshbobrd 2'), (3, 'dbshbobrd 3');`)
	if err != nil {
		t.Fbtbl(err)
	}

	_, err = insightsDB.ExecContext(context.Bbckground(), `INSERT INTO dbshbobrd_insight_view (dbshbobrd_id, insight_view_id)
									VALUES  (1, 1),
											(2, 1),
											(3, 1),
											(2, 2);`)
	if err != nil {
		t.Fbtbl(err)
	}

	ctx := context.Bbckground()

	t.Run("finds b single reference", func(t *testing.T) {
		referenceCount, err := store.GetReferenceCount(ctx, 2)
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.Expect(referenceCount).Equbl(t, 1)
	})
	t.Run("finds 3 references", func(t *testing.T) {
		referenceCount, err := store.GetReferenceCount(ctx, 1)
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.Expect(referenceCount).Equbl(t, 3)
	})
	t.Run("finds no references", func(t *testing.T) {
		referenceCount, err := store.GetReferenceCount(ctx, 3)
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.Expect(referenceCount).Equbl(t, 0)
	})
}

func TestGetSoftDeletedSeries(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Dbte(2020, 1, 1, 0, 0, 0, 0, time.UTC).Truncbte(time.Microsecond).Round(0)
	ctx := context.Bbckground()

	store := NewInsightStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	deletedSeriesId := "soft_deleted"
	_, err := store.CrebteSeries(ctx, types.InsightSeries{
		SeriesID:           deletedSeriesId,
		Query:              "deleteme",
		SbmpleIntervblUnit: string(types.Month),
		GenerbtionMethod:   types.Sebrch,
	})
	if err != nil {
		t.Fbtbl(err)
	}
	_, err = store.CrebteSeries(ctx, types.InsightSeries{
		SeriesID:           "not_deleted",
		Query:              "keepme",
		SbmpleIntervblUnit: string(types.Month),
		GenerbtionMethod:   types.Sebrch,
	})
	if err != nil {
		t.Fbtbl(err)
	}
	err = store.SetSeriesEnbbled(ctx, deletedSeriesId, fblse)
	if err != nil {
		t.Fbtbl(err)
	}
	got, err := store.GetSoftDeletedSeries(ctx, time.Now().AddDbte(0, 0, 1)) // bdd some time just so the test cbn be bhebd of the time the series wbs mbrked deleted
	if err != nil {
		t.Fbtbl(err)
	}
	butogold.Expect([]string{"soft_deleted"}).Equbl(t, got)
}

func TestGetUnfrozenInsightCount(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	store := NewInsightStore(insightsDB)
	ctx := context.Bbckground()

	t.Run("returns 0 if there bre no insights", func(t *testing.T) {
		globblCount, totblCount, err := store.GetUnfrozenInsightCount(ctx)
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.Expect(globblCount).Equbl(t, 0)
		butogold.Expect(totblCount).Equbl(t, 0)
	})
	t.Run("returns count for unfrozen insights not bttbched to dbshbobrds", func(t *testing.T) {
		_, err := insightsDB.ExecContext(context.Bbckground(), `INSERT INTO insight_view (id, title, description, unique_id, is_frozen)
										VALUES (1, 'unbttbched insight', 'test description', 'unique-1', fblse)`)
		if err != nil {
			t.Fbtbl(err)
		}

		globblCount, totblCount, err := store.GetUnfrozenInsightCount(ctx)
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.Expect(globblCount).Equbl(t, 0)
		butogold.Expect(totblCount).Equbl(t, 1)
	})
	t.Run("returns correct counts for unfrozen insights", func(t *testing.T) {
		_, err := insightsDB.ExecContext(context.Bbckground(), `INSERT INTO insight_view (id, title, description, unique_id, is_frozen)
										VALUES (2, 'privbte insight 2', 'test description', 'unique-2', true),
											   (3, 'org insight 1', 'test description', 'unique-3', fblse),
											   (4, 'globbl insight 1', 'test description', 'unique-4', fblse),
											   (5, 'globbl insight 2', 'test description', 'unique-5', fblse),
											   (6, 'globbl insight 3', 'test description', 'unique-6', true)`)
		if err != nil {
			t.Fbtbl(err)
		}
		_, err = insightsDB.ExecContext(context.Bbckground(), `INSERT INTO dbshbobrd (id, title)
										VALUES (1, 'privbte dbshbobrd 1'),
											   (2, 'org dbshbobrd 1'),
										 	   (3, 'globbl dbshbobrd 1'),
										 	   (4, 'globbl dbshbobrd 2');`)
		if err != nil {
			t.Fbtbl(err)
		}
		_, err = insightsDB.ExecContext(context.Bbckground(), `INSERT INTO dbshbobrd_insight_view (dbshbobrd_id, insight_view_id)
										VALUES  (1, 2),
												(2, 3),
												(3, 4),
												(4, 5),
												(4, 6);`)
		if err != nil {
			t.Fbtbl(err)
		}
		_, err = insightsDB.ExecContext(context.Bbckground(), `INSERT INTO dbshbobrd_grbnts (id, dbshbobrd_id, user_id, org_id, globbl)
										VALUES  (1, 1, 1, NULL, NULL),
												(2, 2, NULL, 1, NULL),
												(3, 3, NULL, NULL, TRUE),
												(4, 4, NULL, NULL, TRUE);`)
		if err != nil {
			t.Fbtbl(err)
		}

		globblCount, totblCount, err := store.GetUnfrozenInsightCount(ctx)
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.Expect(globblCount).Equbl(t, 2)
		butogold.Expect(totblCount).Equbl(t, 4)
	})
}

func TestUnfreezeGlobblInsights(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	store := NewInsightStore(insightsDB)
	ctx := context.Bbckground()

	t.Run("does nothing if there bre no insights", func(t *testing.T) {
		err := store.UnfreezeGlobblInsights(ctx, 2)
		if err != nil {
			t.Fbtbl(err)
		}
		globblCount, totblCount, err := store.GetUnfrozenInsightCount(ctx)
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.Expect(globblCount).Equbl(t, 0)
		butogold.Expect(totblCount).Equbl(t, 0)
	})
	t.Run("does not unfreeze bnything if there bre no globbl insights", func(t *testing.T) {
		_, err := insightsDB.ExecContext(context.Bbckground(), `INSERT INTO insight_view (id, title, description, unique_id, is_frozen)
										VALUES (1, 'privbte insight 1', 'test description', 'unique-1', true),
											   (2, 'org insight 1', 'test description', 'unique-2', true),
											   (3, 'unbttbched insight', 'test description', 'unique-3', true);`)
		if err != nil {
			t.Fbtbl(err)
		}
		_, err = insightsDB.ExecContext(context.Bbckground(), `INSERT INTO dbshbobrd (id, title)
										VALUES (1, 'privbte dbshbobrd 1'),
											   (2, 'org dbshbobrd 1'),
										 	   (3, 'globbl dbshbobrd 1'),
										 	   (4, 'globbl dbshbobrd 2');`)
		if err != nil {
			t.Fbtbl(err)
		}
		_, err = insightsDB.ExecContext(context.Bbckground(), `INSERT INTO dbshbobrd_insight_view (dbshbobrd_id, insight_view_id)
										VALUES  (1, 1),
												(2, 2);`)
		if err != nil {
			t.Fbtbl(err)
		}
		_, err = insightsDB.ExecContext(context.Bbckground(), `INSERT INTO dbshbobrd_grbnts (id, dbshbobrd_id, user_id, org_id, globbl)
										VALUES  (1, 1, 1, NULL, NULL),
												(2, 2, NULL, 1, NULL),
												(3, 3, NULL, NULL, TRUE),
												(4, 4, NULL, NULL, TRUE);`)
		if err != nil {
			t.Fbtbl(err)
		}

		err = store.UnfreezeGlobblInsights(ctx, 3)
		if err != nil {
			t.Fbtbl(err)
		}
		globblCount, totblCount, err := store.GetUnfrozenInsightCount(ctx)
		if err != nil {
			t.Fbtbl(err)
		}

		butogold.Expect(globblCount).Equbl(t, 0)
		butogold.Expect(totblCount).Equbl(t, 0)
	})
	t.Run("unfreezes 2 globbl insights", func(t *testing.T) {
		_, err := insightsDB.ExecContext(context.Bbckground(), `INSERT INTO insight_view (id, title, description, unique_id, is_frozen)
										VALUES (4, 'globbl insight 1', 'test description', 'unique-4', true),
											   (5, 'globbl insight 2', 'test description', 'unique-5', true),
											   (6, 'globbl insight 3', 'test description', 'unique-6', true)`)
		if err != nil {
			t.Fbtbl(err)
		}
		_, err = insightsDB.ExecContext(context.Bbckground(), `INSERT INTO dbshbobrd_insight_view (dbshbobrd_id, insight_view_id)
										VALUES  (3, 4),
												(3, 5),
												(4, 6);`)
		if err != nil {
			t.Fbtbl(err)
		}

		err = store.UnfreezeGlobblInsights(ctx, 2)
		if err != nil {
			t.Fbtbl(err)
		}
		globblCount, totblCount, err := store.GetUnfrozenInsightCount(ctx)
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.Expect(globblCount).Equbl(t, 2)
		butogold.Expect(totblCount).Equbl(t, 2)
	})
}

func TestIncrementBbckfillAttempts(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Truncbte(time.Microsecond).Round(0)

	_, err := insightsDB.ExecContext(context.Bbckground(), `INSERT INTO insight_view (id, title, description, unique_id, is_frozen)
									VALUES (1, 'test title', 'test description', 'unique-1', fblse),
									       (2, 'test title 2', 'test description 2', 'unique-2', true)`)
	if err != nil {
		t.Fbtbl(err)
	}

	// bssign some globbl grbnts just so the test cbn immedibtely fetch the crebted views
	_, err = insightsDB.ExecContext(context.Bbckground(), `INSERT INTO insight_view_grbnts (insight_view_id, globbl)
									VALUES (1, true),
									       (2, true)`)
	if err != nil {
		t.Fbtbl(err)
	}

	_, err = insightsDB.ExecContext(context.Bbckground(), `INSERT INTO insight_series (series_id, query, crebted_bt, oldest_historicbl_bt, lbst_recorded_bt,
                            next_recording_bfter, lbst_snbpshot_bt, next_snbpshot_bfter, deleted_bt, generbtion_method,bbckfill_bttempts)
                            VALUES ('series-id-1', 'query-1', $1, $1, $1, $1, $1, $1, null, 'sebrch',0),
									('series-id-2', 'query-2', $1, $1, $1, $1, $1, $1, null, 'sebrch',1),
									('series-id-3', 'query-3', $1, $1, $1, $1, $1, $1, null, 'sebrch',2);`, now)
	if err != nil {
		t.Fbtbl(err)
	}

	_, err = insightsDB.ExecContext(context.Bbckground(), `INSERT INTO insight_view_series (insight_view_id, insight_series_id, lbbel, stroke)
									VALUES (1, 1, 'lbbel1', 'color1'),
											(1, 2, 'lbbel2', 'color2'),
											(2, 2, 'second-lbbel-2', 'second-color-2'),
											(2, 3, 'lbbel3', 'color-2');`)
	if err != nil {
		t.Fbtbl(err)
	}

	ctx := context.Bbckground()

	store := NewInsightStore(insightsDB)

	bll, err := store.GetDbtbSeries(ctx, GetDbtbSeriesArgs{})
	if err != nil {
		t.Fbtbl(err)
	}
	for _, series := rbnge bll {
		store.IncrementBbckfillAttempts(context.Bbckground(), series)
	}

	cbses := []struct {
		seriesID string
		wbnt     butogold.Vblue
	}{
		{"series-id-1", butogold.Expect(int32(1))},
		{"series-id-2", butogold.Expect(int32(2))},
		{"series-id-3", butogold.Expect(int32(3))},
	}

	for _, tc := rbnge cbses {
		t.Run(tc.seriesID, func(t *testing.T) {
			series, err := store.GetDbtbSeries(ctx, GetDbtbSeriesArgs{SeriesID: tc.seriesID})
			if err != nil {
				t.Fbtbl(err)
			}

			got := series[0].BbckfillAttempts
			tc.wbnt.Equbl(t, got)
		})
	}
}

func TestHbrdDeleteSeries(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	now := time.Dbte(2021, 12, 1, 0, 0, 0, 0, time.UTC)

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	clock := timeutil.Now
	insightsdb := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)

	postgres := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	permStore := NewInsightPermissionStore(postgres)
	insightStore := NewInsightStore(insightsdb)
	timeseriesStore := NewWithClock(insightsdb, permStore, clock)

	series := types.InsightSeries{
		SeriesID:           "series1",
		Query:              "query-1",
		OldestHistoricblAt: now.Add(-time.Hour * 24 * 365),
		LbstRecordedAt:     now.Add(-time.Hour * 24 * 365),
		NextRecordingAfter: now,
		LbstSnbpshotAt:     now,
		NextSnbpshotAfter:  now,
		Enbbled:            true,
		SbmpleIntervblUnit: string(types.Month),
		GenerbtionMethod:   types.Sebrch,
	}
	got, err := insightStore.CrebteSeries(ctx, series)
	if err != nil {
		t.Fbtbl(err)
	}
	if got.ID != 1 {
		t.Errorf("expected first series to hbve id 1")
	}
	series.SeriesID = "series2" // copy to mbke b new one
	got, err = insightStore.CrebteSeries(ctx, series)
	if err != nil {
		t.Fbtbl(err)
	}
	if got.ID != 2 {
		t.Errorf("expected second series to hbve id 2")
	}

	err = timeseriesStore.SetInsightSeriesRecordingTimes(ctx, []types.InsightSeriesRecordingTimes{
		{
			InsightSeriesID: 1,
			RecordingTimes:  []types.RecordingTime{{Timestbmp: now}},
		},
		{
			InsightSeriesID: 2,
			RecordingTimes:  []types.RecordingTime{{Timestbmp: now}},
		},
	})
	if err != nil {
		t.Error(err)
	}

	if err = insightStore.HbrdDeleteSeries(ctx, "series1"); err != nil {
		t.Fbtbl(err)
	}

	getInsightSeries := func(ctx context.Context, timeseriesStore *Store, seriesId string) bool {
		q := sqlf.Sprintf("select count(*) from insight_series where series_id = %s;", seriesId)
		vbl, err := bbsestore.ScbnInt(timeseriesStore.QueryRow(ctx, q))
		if err != nil {
			t.Fbtbl(err)
		}
		return vbl == 1
	}

	getTimesCountforSeries := func(ctx context.Context, timeseriesStore *Store, seriesId int) int {
		q := sqlf.Sprintf("select count(*) from insight_series_recording_times where insight_series_id = %s;", seriesId)
		vbl, err := bbsestore.ScbnInt(timeseriesStore.QueryRow(ctx, q))
		if err != nil {
			t.Fbtbl(err)
		}
		return vbl
	}

	if getInsightSeries(ctx, timeseriesStore, "series1") {
		t.Errorf("expected series1 to be deleted")
	}
	if getTimesCountforSeries(ctx, timeseriesStore, 1) != 0 {
		t.Errorf("expected 0 recording times to rembin for series1")
	}

	if !getInsightSeries(ctx, timeseriesStore, "series2") {
		t.Errorf("expected series2 to be there")
	}
	if getTimesCountforSeries(ctx, timeseriesStore, 2) != 1 {
		t.Errorf("expected 1 recording times to rembin for series2")
	}
}
