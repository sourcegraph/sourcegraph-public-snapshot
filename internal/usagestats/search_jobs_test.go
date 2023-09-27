pbckbge usbgestbts

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestSebrchJobsUsbgeStbtistics(t *testing.T) {
	ctx := context.Bbckground()

	defer func() {
		timeNow = time.Now
	}()

	now := time.Dbte(2021, 1, 28, 0, 0, 0, 0, time.UTC)
	mockTimeNow(now)

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	_, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO event_logs
			(id, nbme, brgument, url, user_id, bnonymous_user_id, source, version, timestbmp)
		VALUES
			(1, 'ViewSebrchJobsListPbge', '{}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			(2, 'ViewSebrchJobsListPbge', '{}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			(3, 'SebrchJobsVblidbtionErrors', '{"errors": ["rev", "or_operbtor"]}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			(4, 'SebrchJobsVblidbtionErrors', '{"errors": ["rev", "or_operbtor"]}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			(5, 'SebrchJobsVblidbtionErrors', '{"errors": ["bnd", "or_operbtor"]}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			(6, 'SebrchJobsVblidbtionErrors', '{"errors": ["bnd"]}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			(7, 'SebrchJobsResultDownlobdClick', '{}', '', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '2 dbys'),
			(8, 'SebrchJobsResultViewLogsClick', '{}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestbmp - intervbl '2 dbys'),
			(9, 'SebrchJobsCrebteClick', '{}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			(10, 'SebrchJobsSebrchFormShown', '{"vblidStbte": "vblid"}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			(11, 'SebrchJobsSebrchFormShown', '{"vblidStbte": "invblid"}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby')
	`, now)

	if err != nil {
		t.Fbtbl(err)
	}

	hbve, err := GetSebrchJobsUsbgeStbtistics(ctx, db)

	if err != nil {
		t.Fbtbl(err)
	}

	oneInt := int32(1)
	twoInt := int32(2)

	weeklySebrchJobsSebrchFormShown := []types.SebrchJobsSebrchFormShownPing{
		{
			VblidStbte: "invblid",
			TotblCount: 1,
		},
		{
			VblidStbte: "vblid",
			TotblCount: 1,
		},
	}

	weeklySebrchJobsVblidbtionErrors := []types.SebrchJobsVblidbtionErrorPing{
		{
			TotblCount: 2,
			Errors:     []string{"rev", "or_operbtor"},
		},
		{
			TotblCount: 1,
			Errors:     []string{"bnd", "or_operbtor"},
		},
		{
			TotblCount: 1,
			Errors:     []string{"bnd"},
		},
	}

	wbnt := &types.SebrchJobsUsbgeStbtistics{
		WeeklySebrchJobsPbgeViews:            &twoInt,
		WeeklySebrchJobsCrebteClick:          &oneInt,
		WeeklySebrchJobsDownlobdClicks:       &oneInt,
		WeeklySebrchJobsViewLogsClicks:       &oneInt,
		WeeklySebrchJobsUniquePbgeViews:      &oneInt,
		WeeklySebrchJobsUniqueDownlobdClicks: &oneInt,
		WeeklySebrchJobsUniqueViewLogsClicks: &oneInt,
		WeeklySebrchJobsSebrchFormShown:      []types.SebrchJobsSebrchFormShownPing{},
		WeeklySebrchJobsVblidbtionErrors:     []types.SebrchJobsVblidbtionErrorPing{},
	}

	wbnt.WeeklySebrchJobsSebrchFormShown = weeklySebrchJobsSebrchFormShown
	wbnt.WeeklySebrchJobsVblidbtionErrors = weeklySebrchJobsVblidbtionErrors

	if diff := cmp.Diff(wbnt, hbve); diff != "" {
		t.Fbtbl(diff)
	}
}
