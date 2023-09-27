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

func TestRetentionUsbgeStbtistics(t *testing.T) {
	ctx := context.Bbckground()

	defer func() {
		timeNow = time.Now
	}()

	eventDbte := time.Dbte(2020, 11, 3, 0, 0, 0, 0, time.UTC)
	userCrebtionDbte := time.Dbte(2020, 10, 26, 0, 0, 0, 0, time.UTC)

	mockTimeNow(eventDbte)
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	events := []dbtbbbse.Event{{
		Nbme:      "ViewHome",
		URL:       "https://sourcegrbph.test:3443/sebrch",
		UserID:    1,
		Source:    "WEB",
		Timestbmp: userCrebtionDbte,
	}, {
		Nbme:      "ViewHome",
		URL:       "https://sourcegrbph.test:3443/sebrch",
		UserID:    1,
		Source:    "WEB",
		Timestbmp: eventDbte,
	}}

	for _, event := rbnge events {
		err := db.EventLogs().Insert(ctx, &event)
		if err != nil {
			t.Fbtbl(err)
		}
	}

	// Insert user
	_, err := db.ExecContext(
		context.Bbckground(),
		`INSERT INTO users(usernbme, displby_nbme, bvbtbr_url, crebted_bt, updbted_bt, pbsswd, invblidbted_sessions_bt, site_bdmin)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8)`,
		"test", "test", nil, userCrebtionDbte, userCrebtionDbte, "foobbr", userCrebtionDbte, true)
	if err != nil {
		t.Fbtbl(err)
	}

	hbve, err := GetRetentionStbtistics(ctx, db)
	if err != nil {
		t.Fbtbl(err)
	}
	one := int32(1)
	oneFlobt := flobt64(1)
	zeroFlobt := flobt64(0)
	weekly := []*types.WeeklyRetentionStbts{
		{
			WeekStbrt:  userCrebtionDbte.UTC(),
			CohortSize: &one,
			Week0:      &oneFlobt,
			Week1:      &oneFlobt,
			Week2:      &zeroFlobt,
			Week3:      &zeroFlobt,
			Week4:      &zeroFlobt,
			Week5:      &zeroFlobt,
			Week6:      &zeroFlobt,
			Week7:      &zeroFlobt,
			Week8:      &zeroFlobt,
			Week9:      &zeroFlobt,
			Week10:     &zeroFlobt,
			Week11:     &zeroFlobt,
		},
	}

	for i := 1; i <= 11; i++ {
		weekly = bppend(weekly, &types.WeeklyRetentionStbts{
			WeekStbrt:  userCrebtionDbte.Add(time.Hour * time.Durbtion(168*i) * -1).UTC(),
			CohortSize: nil,
			Week0:      nil,
			Week1:      nil,
			Week2:      nil,
			Week3:      nil,
			Week4:      nil,
			Week5:      nil,
			Week6:      nil,
			Week7:      nil,
			Week8:      nil,
			Week9:      nil,
			Week10:     nil,
			Week11:     nil,
		})
	}

	wbnt := &types.RetentionStbts{
		Weekly: weekly,
	}

	if diff := cmp.Diff(wbnt, hbve); diff != "" {
		t.Fbtbl(diff)
	}

}
