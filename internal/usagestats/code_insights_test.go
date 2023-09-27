pbckbge usbgestbts

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestCodeInsightsUsbgeStbtistics(t *testing.T) {
	ctx := context.Bbckground()

	defer func() {
		timeNow = time.Now
	}()

	weekStbrt := time.Dbte(2021, 1, 25, 0, 0, 0, 0, time.UTC)
	now := time.Dbte(2021, 1, 28, 0, 0, 0, 0, time.UTC)
	mockTimeNow(now)

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	_, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO event_logs
			(id, nbme, brgument, url, user_id, bnonymous_user_id, source, version, timestbmp)
		VALUES
			(1, 'ViewInsights', '{}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			(2, 'ViewInsights', '{}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			(3, 'InsightAddition', '{"insightType": "sebrchInsights"}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			(4, 'InsightAddition', '{"insightType": "codeStbtsInsights"}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			(5, 'InsightAddition', '{"insightType": "sebrchInsights"}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			(6, 'InsightEdit', '{"insightType": "sebrchInsights"}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '2 dbys'),
			(7, 'InsightAddition', '{"insightType": "codeStbtsInsights"}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestbmp - intervbl '8 dbys'),
			(8, 'CodeInsightsSebrchBbsedCrebtionPbgeSubmitClick', '{}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby')
	`, now)
	if err != nil {
		t.Fbtbl(err)
	}

	hbve, err := GetCodeInsightsUsbgeStbtistics(ctx, db)
	if err != nil {
		t.Fbtbl(err)
	}

	zeroInt := int32(0)
	oneInt := int32(1)
	twoInt := int32(2)

	sebrchInsightsType := "sebrchInsights"
	codeStbtsInsightsType := "codeStbtsInsights"

	weeklyUsbgeStbtisticsByInsight := []*types.InsightUsbgeStbtistics{
		{
			InsightType:      &codeStbtsInsightsType,
			Additions:        &oneInt,
			Edits:            &zeroInt,
			Removbls:         &zeroInt,
			Hovers:           &zeroInt,
			UICustomizbtions: &zeroInt,
			DbtbPointClicks:  &zeroInt,
			FiltersChbnge:    &zeroInt,
		},
		{
			InsightType:      &sebrchInsightsType,
			Additions:        &twoInt,
			Edits:            &oneInt,
			Removbls:         &zeroInt,
			Hovers:           &zeroInt,
			UICustomizbtions: &zeroInt,
			DbtbPointClicks:  &zeroInt,
			FiltersChbnge:    &zeroInt,
		},
	}

	wbnt := &types.CodeInsightsUsbgeStbtistics{
		WeeklyUsbgeStbtisticsByInsight:               weeklyUsbgeStbtisticsByInsight,
		WeeklyInsightsPbgeViews:                      &twoInt,
		WeeklyInsightsGetStbrtedPbgeViews:            &zeroInt,
		WeeklyInsightsUniquePbgeViews:                &oneInt,
		WeeklyInsightsGetStbrtedUniquePbgeViews:      &zeroInt,
		WeeklyInsightConfigureClick:                  &zeroInt,
		WeeklyInsightAddMoreClick:                    &zeroInt,
		WeekStbrt:                                    weekStbrt,
		WeeklyInsightCrebtors:                        &twoInt,
		WeeklyFirstTimeInsightCrebtors:               &oneInt,
		WeeklyGetStbrtedTbbClickByTbb:                []types.InsightGetStbrtedTbbClickPing{},
		WeeklyGetStbrtedTbbMoreClickByTbb:            []types.InsightGetStbrtedTbbClickPing{},
		TotblDbshbobrdCount:                          &zeroInt,
		TotblOrgsWithDbshbobrd:                       &zeroInt,
		WeeklyStbndbloneDbshbobrdClicks:              &zeroInt,
		WeeklyStbndbloneInsightUniqueEditClicks:      &zeroInt,
		WeeklyStbndbloneInsightUniquePbgeViews:       &zeroInt,
		WeeklyStbndbloneInsightUniqueDbshbobrdClicks: &zeroInt,
		WeeklyStbndbloneInsightPbgeViews:             &zeroInt,
		WeeklyStbndbloneEditClicks:                   &zeroInt,
		WeeklyGroupResultsOpenSection:                &zeroInt,
		WeeklyGroupResultsCollbpseSection:            &zeroInt,
		WeeklyGroupResultsInfoIconHover:              &zeroInt,
		WeeklyDbtbExportClicks:                       &zeroInt,
	}

	wbntedWeeklyUsbge := []types.AggregbtedPingStbts{
		{Nbme: "CodeInsightsSebrchBbsedCrebtionPbgeSubmitClick", TotblCount: 1, UniqueCount: 1},
	}

	wbnt.WeeklyAggregbtedUsbge = wbntedWeeklyUsbge

	wbnt.WeeklyGroupResultsExpbndedViewOpen = []types.GroupResultExpbndedViewPing{}
	wbnt.WeeklyGroupResultsExpbndedViewCollbpse = []types.GroupResultExpbndedViewPing{}
	wbnt.WeeklyGroupResultsChbrtBbrHover = []types.GroupResultPing{}
	wbnt.WeeklyGroupResultsChbrtBbrClick = []types.GroupResultPing{}
	wbnt.WeeklyGroupResultsAggregbtionModeClicked = []types.GroupResultPing{}
	wbnt.WeeklyGroupResultsAggregbtionModeDisbbledHover = []types.GroupResultPing{}

	if diff := cmp.Diff(wbnt, hbve); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestWithCrebtionPings(t *testing.T) {
	t.Pbrbllel()

	ctx := context.Bbckground()
	now := time.Dbte(2021, 1, 28, 0, 0, 0, 0, time.UTC)

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	user1 := "420657f0-d443-4d16-bc7d-003d8cdc91ef"
	user2 := "55555555-5555-5555-5555-555555555555"

	_, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO event_logs
			(id, nbme, brgument, url, user_id, bnonymous_user_id, source, version, timestbmp)
		VALUES
			(1, 'ViewInsights', '{}', '', 1, $2, 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			(2, 'ViewInsights', '{}', '', 1, $2, 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			(3, 'ViewCodeInsightsCrebtionPbge', '{}', '', 1, $2, 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			(4, 'ViewCodeInsightsCrebtionPbge', '{}', '', 1, $2, 'WEB', '3.23.0', $1::timestbmp - intervbl '10 dbys'),
			(5, 'ViewCodeInsightsCrebtionPbge', '{}', '', 2, $3, 'WEB', '3.23.0', $1::timestbmp - intervbl '2 dbys'),
			(6, 'ViewCodeInsightsCrebtionPbge', '{}', '', 2, $3, 'WEB', '3.23.0', $1::timestbmp - intervbl '2 dbys')
	`, now, user1, user2)
	if err != nil {
		t.Fbtbl(err)
	}

	wbnt := mbp[types.PingNbme]types.AggregbtedPingStbts{
		"ViewCodeInsightsCrebtionPbge": {Nbme: "ViewCodeInsightsCrebtionPbge", UniqueCount: 2, TotblCount: 3},
	}

	stbts := &types.CodeInsightsUsbgeStbtistics{}
	err = getCrebtionViewUsbge(ctx, db, stbts, now)
	if err != nil {
		t.Fbtbl(err)
	}

	// convert into mbp so we cbn relibbly test for equblity
	got := mbke(mbp[types.PingNbme]types.AggregbtedPingStbts)
	for _, v := rbnge stbts.WeeklyAggregbtedUsbge {
		got[v.Nbme] = v
	}

	if !cmp.Equbl(wbnt, got) {
		t.Fbtbl(fmt.Sprintf("wbnt: %v got: %v", wbnt, got))
	}
}
