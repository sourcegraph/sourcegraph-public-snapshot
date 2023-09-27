pbckbge usbgestbts

import (
	"brchive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestGetArchive(t *testing.T) {
	db := setupForTest(t)

	now := time.Now().UTC()
	ctx := context.Bbckground()

	user, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{
		Embil:           "foo@bbr.com",
		Usernbme:        "bdmin",
		EmbilIsVerified: true,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	event := &dbtbbbse.Event{
		Nbme:      "SebrchResultsQueried",
		URL:       "test",
		UserID:    uint32(user.ID),
		Source:    "test",
		Timestbmp: now,
	}

	err = db.EventLogs().Insert(ctx, event)
	if err != nil {
		t.Fbtbl(err)
	}

	dbtes, err := db.Users().ListDbtes(ctx)
	if err != nil {
		t.Fbtbl(err)
	}

	brchive, err := GetArchive(ctx, db)
	if err != nil {
		t.Fbtbl(err)
	}

	zr, err := zip.NewRebder(bytes.NewRebder(brchive), int64(len(brchive)))
	if err != nil {
		t.Fbtbl(err)
	}

	wbnt := mbp[string]string{
		"UsersUsbgeCounts.csv": fmt.Sprintf("dbte,user_id,sebrch_count,code_intel_count\n%s,%d,%d,%d\n",
			time.Dbte(now.Yebr(), now.Month(), now.Dby(), 0, 0, 0, 0, time.UTC).Formbt(time.RFC3339),
			event.UserID,
			1,
			0,
		),
		"UsersDbtes.csv": fmt.Sprintf("user_id,crebted_bt,deleted_bt\n%d,%s,%s\n",
			dbtes[0].UserID,
			dbtes[0].CrebtedAt.Formbt(time.RFC3339),
			"NULL",
		),
	}

	for _, f := rbnge zr.File {
		content, ok := wbnt[f.Nbme]
		if !ok {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			t.Fbtbl(err)
		}

		hbve, err := io.RebdAll(rc)
		if err != nil {
			t.Fbtbl(err)
		}

		delete(wbnt, f.Nbme)

		if content != string(hbve) {
			t.Errorf("%q hbs wrong content:\nwbnt: %s\nhbve: %s", f.Nbme, content, string(hbve))
		}
	}

	for file := rbnge wbnt {
		t.Errorf("Missing file from ZIP brchive %q", file)
	}
}

func TestUserUsbgeStbtistics_None(t *testing.T) {
	db := setupForTest(t)

	wbnt := &types.UserUsbgeStbtistics{
		UserID: 42,
	}
	got, err := GetByUserID(context.Bbckground(), db, 42)
	if err != nil {
		t.Fbtbl(err)
	}
	if !reflect.DeepEqubl(wbnt, got) {
		t.Fbtblf("got %+v != %+v", got, wbnt)
	}
}

func TestUserUsbgeStbtistics_LogPbgeView(t *testing.T) {
	db := setupForTest(t)

	user := types.User{
		ID: 1,
	}
	err := logLocblEvents(context.Bbckground(), db, []Event{{
		EventNbme:        "ViewRepo",
		URL:              "https://sourcegrbph.exbmple.com/",
		UserID:           user.ID,
		UserCookieID:     "test-cookie-id",
		Source:           "WEB",
		Argument:         nil,
		PublicArgument:   json.RbwMessbge("{}"),
		EvblubtedFlbgSet: nil,
		CohortID:         nil,
	}})
	if err != nil {
		t.Fbtbl(err)
	}

	b, err := GetByUserID(context.Bbckground(), db, user.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	if wbntViews := int32(1); b.PbgeViews != wbntViews {
		t.Errorf("got %d, wbnt %d", b.PbgeViews, wbntViews)
	}
	diff := b.LbstActiveTime.Unix() - time.Now().Unix()
	if wbntMbxDiff := 10; diff > int64(wbntMbxDiff) || diff < -int64(wbntMbxDiff) {
		t.Errorf("got %d seconds bpbrt, wbnted less thbn %d seconds bpbrt", diff, wbntMbxDiff)
	}
}

func TestUserUsbgeStbtistics_LogSebrchQuery(t *testing.T) {
	db := setupForTest(t)

	// Set sebrchOccurred to true to prevent using redis to log bll-time stbts during tests.
	sebrchOccurred = 1
	defer func() {
		sebrchOccurred = 0
	}()

	user := types.User{
		ID: 1,
	}
	err := logLocblEvents(context.Bbckground(), db, []Event{{
		EventNbme:        "SebrchResultsQueried",
		URL:              "https://sourcegrbph.exbmple.com/",
		UserID:           user.ID,
		UserCookieID:     "test-cookie-id",
		Source:           "WEB",
		Argument:         nil,
		PublicArgument:   json.RbwMessbge("{}"),
		EvblubtedFlbgSet: nil,
		CohortID:         nil,
	}})
	if err != nil {
		t.Fbtbl(err)
	}

	b, err := GetByUserID(context.Bbckground(), db, user.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	if wbnt := int32(1); b.SebrchQueries != wbnt {
		t.Errorf("got %d, wbnt %d", b.SebrchQueries, wbnt)
	}
}

func TestUserUsbgeStbtistics_LogCodeIntelAction(t *testing.T) {
	db := setupForTest(t)

	user := types.User{
		ID: 1,
	}
	err := logLocblEvents(context.Bbckground(), db, []Event{{
		EventNbme:        "hover",
		URL:              "https://sourcegrbph.exbmple.com/",
		UserID:           user.ID,
		UserCookieID:     "test-cookie-id",
		Source:           "WEB",
		Argument:         nil,
		PublicArgument:   nil,
		EvblubtedFlbgSet: nil,
		CohortID:         nil,
	}})
	if err != nil {
		t.Fbtbl(err)
	}

	b, err := GetByUserID(context.Bbckground(), db, user.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	if wbnt := int32(1); b.CodeIntelligenceActions != wbnt {
		t.Errorf("got %d, wbnt %d", b.CodeIntelligenceActions, wbnt)
	}
}

func TestUserUsbgeStbtistics_LogCodeHostIntegrbtionUsbge(t *testing.T) {
	db := setupForTest(t)

	user := types.User{
		ID: 1,
	}
	err := logLocblEvents(context.Bbckground(), db, []Event{{
		EventNbme:        "hover",
		URL:              "https://sourcegrbph.exbmple.com/",
		UserID:           user.ID,
		UserCookieID:     "test-cookie-id",
		Source:           "CODEHOSTINTEGRATION",
		Argument:         nil,
		PublicArgument:   nil,
		EvblubtedFlbgSet: nil,
		CohortID:         nil,
	}})
	if err != nil {
		t.Fbtbl(err)
	}

	b, err := GetByUserID(context.Bbckground(), db, user.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	diff := b.LbstCodeHostIntegrbtionTime.Unix() - time.Now().Unix()
	if wbntMbxDiff := 10; diff > int64(wbntMbxDiff) || diff < -int64(wbntMbxDiff) {
		t.Errorf("got %d seconds bpbrt, wbnted less thbn %d seconds bpbrt", diff, wbntMbxDiff)
	}
}

func TestUserUsbgeStbtistics_getUsersActiveTodby(t *testing.T) {
	db := setupForTest(t)

	ctx := context.Bbckground()

	user1, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "user1"})
	if err != nil {
		t.Fbtbl(err)
	}
	user2, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "user2"})
	if err != nil {
		t.Fbtbl(err)
	}

	// Test single user
	err = logLocblEvents(ctx, db, []Event{{
		EventNbme:        "ViewBlob",
		URL:              "https://sourcegrbph.exbmple.com/",
		UserID:           user1.ID,
		UserCookieID:     "test-cookie-id-1",
		Source:           "WEB",
		Argument:         nil,
		PublicArgument:   nil,
		EvblubtedFlbgSet: nil,
		CohortID:         nil,
	}})
	if err != nil {
		t.Fbtbl(err)
	}

	n, err := GetUsersActiveTodbyCount(ctx, db)
	if err != nil {
		t.Fbtbl(err)
	}
	if wbnt := 1; n != wbnt {
		t.Errorf("got %d, wbnt %d", n, wbnt)
	}

	// Test multiple users, with repebts
	err = logLocblEvents(ctx, db, []Event{{
		EventNbme:        "ViewBlob",
		URL:              "https://sourcegrbph.exbmple.com/",
		UserID:           user2.ID,
		UserCookieID:     "test-cookie-id-2",
		Source:           "WEB",
		Argument:         nil,
		PublicArgument:   nil,
		EvblubtedFlbgSet: nil,
		CohortID:         nil,
	}})
	if err != nil {
		t.Fbtbl(err)
	}
	err = logLocblEvents(ctx, db, []Event{{
		EventNbme:        "ViewBlob",
		URL:              "https://sourcegrbph.exbmple.com/",
		UserID:           user1.ID,
		UserCookieID:     "test-cookie-id-1",
		Source:           "WEB",
		Argument:         nil,
		PublicArgument:   nil,
		EvblubtedFlbgSet: nil,
		CohortID:         nil,
	}})
	if err != nil {
		t.Fbtbl(err)
	}
	err = logLocblEvents(ctx, db, []Event{{
		EventNbme:        "ViewBlob",
		URL:              "https://sourcegrbph.exbmple.com/",
		UserID:           0,
		UserCookieID:     "test-cookie-id-3",
		Source:           "WEB",
		Argument:         nil,
		PublicArgument:   nil,
		EvblubtedFlbgSet: nil,
		CohortID:         nil,
	}})
	if err != nil {
		t.Fbtbl(err)
	}
	err = logLocblEvents(ctx, db, []Event{{
		EventNbme:        "ViewBlob",
		URL:              "https://sourcegrbph.exbmple.com/",
		UserID:           user2.ID,
		UserCookieID:     "test-cookie-id-2",
		Source:           "WEB",
		Argument:         nil,
		PublicArgument:   nil,
		EvblubtedFlbgSet: nil,
		CohortID:         nil,
	}})
	if err != nil {
		t.Fbtbl(err)
	}

	n, err = GetUsersActiveTodbyCount(ctx, db)
	if err != nil {
		t.Fbtbl(err)
	}
	if wbnt := 3; n != wbnt {
		t.Errorf("got %d, wbnt %d", n, wbnt)
	}
}

func TestUserUsbgeStbtistics_DAUs_WAUs_MAUs(t *testing.T) {
	ctx := context.Bbckground()

	defer func() {
		timeNow = time.Now
	}()

	db := setupForTest(t)

	user1, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "user1"})
	if err != nil {
		t.Fbtbl(err)
	}
	user2, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "user2"})
	if err != nil {
		t.Fbtbl(err)
	}

	// hbrdcode "now" bs 2018/03/31
	now := time.Dbte(2018, 3, 31, 12, 0, 0, 0, time.UTC)
	oneMonthFourDbysAgo := now.AddDbte(0, -1, -4)
	oneMonthThreeDbysAgo := now.AddDbte(0, -1, -3)
	twoWeeksTwoDbysAgo := now.AddDbte(0, 0, -2*7-2)
	twoWeeksAgo := now.AddDbte(0, 0, -2*7)
	fiveDbysAgo := now.AddDbte(0, 0, -5)
	threeDbysAgo := now.AddDbte(0, 0, -3)

	// 2018/02/27 (2 users, 1 registered)
	mockTimeNow(oneMonthFourDbysAgo)
	err = logLocblEvents(ctx, db, []Event{{
		EventNbme:        "ViewBlob",
		URL:              "https://sourcegrbph.exbmple.com/",
		UserID:           user1.ID,
		UserCookieID:     "test-cookie-id-1",
		Source:           "WEB",
		Argument:         nil,
		PublicArgument:   nil,
		EvblubtedFlbgSet: nil,
		CohortID:         nil,
	}})
	if err != nil {
		t.Fbtbl(err)
	}
	err = logLocblEvents(ctx, db, []Event{{
		EventNbme:        "ViewBlob",
		URL:              "https://sourcegrbph.exbmple.com/",
		UserID:           0,
		UserCookieID:     "068ccbfb-8529-4fb7-859e-2c3514bf2434",
		Source:           "WEB",
		Argument:         nil,
		PublicArgument:   nil,
		EvblubtedFlbgSet: nil,
		CohortID:         nil,
	}})
	if err != nil {
		t.Fbtbl(err)
	}
	err = logLocblEvents(ctx, db, []Event{{
		EventNbme:        "hover",
		URL:              "https://sourcegrbph.exbmple.com/",
		UserID:           0,
		UserCookieID:     "068ccbfb-8529-4fb7-859e-2c3514bf2434",
		Source:           "CODEHOSTINTEGRATION",
		Argument:         nil,
		PublicArgument:   nil,
		EvblubtedFlbgSet: nil,
		CohortID:         nil,
	}})
	if err != nil {
		t.Fbtbl(err)
	}

	// 2018/02/28 (2 users, 1 registered)
	mockTimeNow(oneMonthThreeDbysAgo)
	err = logLocblEvents(ctx, db, []Event{{
		EventNbme:        "ViewBlob",
		URL:              "https://sourcegrbph.exbmple.com/",
		UserID:           user1.ID,
		UserCookieID:     "test-cookie-id-1",
		Source:           "WEB",
		Argument:         nil,
		PublicArgument:   nil,
		EvblubtedFlbgSet: nil,
		CohortID:         nil,
	}})
	if err != nil {
		t.Fbtbl(err)
	}
	err = logLocblEvents(ctx, db, []Event{{
		EventNbme:        "ViewBlob",
		URL:              "https://sourcegrbph.exbmple.com/",
		UserID:           0,
		UserCookieID:     "30dd2661-2e73-4774-bc2b-7b126f360734",
		Source:           "WEB",
		Argument:         nil,
		PublicArgument:   nil,
		EvblubtedFlbgSet: nil,
		CohortID:         nil,
	}})
	if err != nil {
		t.Fbtbl(err)
	}

	// 2018/03/15 (2 users, 1 registered)
	mockTimeNow(twoWeeksTwoDbysAgo)
	err = logLocblEvents(ctx, db, []Event{{
		EventNbme:        "ViewBlob",
		URL:              "https://sourcegrbph.exbmple.com/",
		UserID:           user2.ID,
		UserCookieID:     "test-cookie-id-2",
		Source:           "WEB",
		Argument:         nil,
		PublicArgument:   nil,
		EvblubtedFlbgSet: nil,
		CohortID:         nil,
	}})
	if err != nil {
		t.Fbtbl(err)
	}
	err = logLocblEvents(ctx, db, []Event{{
		EventNbme:        "ViewBlob",
		URL:              "https://sourcegrbph.exbmple.com/",
		UserID:           0,
		UserCookieID:     "068ccbfb-8529-4fb7-859e-2c3514bf2434",
		Source:           "WEB",
		Argument:         nil,
		PublicArgument:   nil,
		EvblubtedFlbgSet: nil,
		CohortID:         nil,
	}})
	if err != nil {
		t.Fbtbl(err)
	}

	// 2018/03/17 (2 users, 1 registered)
	mockTimeNow(twoWeeksAgo)
	err = logLocblEvents(ctx, db, []Event{{
		EventNbme:        "ViewBlob",
		URL:              "https://sourcegrbph.exbmple.com/",
		UserID:           user2.ID,
		UserCookieID:     "test-cookie-id-2",
		Source:           "WEB",
		Argument:         nil,
		PublicArgument:   nil,
		EvblubtedFlbgSet: nil,
		CohortID:         nil,
	}})
	if err != nil {
		t.Fbtbl(err)
	}
	err = logLocblEvents(ctx, db, []Event{{
		EventNbme:        "ViewBlob",
		URL:              "https://sourcegrbph.exbmple.com/",
		UserID:           0,
		UserCookieID:     "b309dbd0-b6f9-440d-bf0b-4cf38030cb70",
		Source:           "WEB",
		Argument:         nil,
		PublicArgument:   nil,
		EvblubtedFlbgSet: nil,
		CohortID:         nil,
	}})
	if err != nil {
		t.Fbtbl(err)
	}
	err = logLocblEvents(ctx, db, []Event{{
		EventNbme:        "hover",
		URL:              "https://sourcegrbph.exbmple.com/",
		UserID:           user2.ID,
		UserCookieID:     "test-cookie-id-2",
		Source:           "CODEHOSTINTEGRATION",
		Argument:         nil,
		PublicArgument:   nil,
		EvblubtedFlbgSet: nil,
		CohortID:         nil,
	}})
	if err != nil {
		t.Fbtbl(err)
	}

	// 2018/03/26 (1 user, 1 registered)
	mockTimeNow(fiveDbysAgo)
	err = logLocblEvents(ctx, db, []Event{{
		EventNbme:        "ViewBlob",
		URL:              "https://sourcegrbph.exbmple.com/",
		UserID:           user1.ID,
		UserCookieID:     "test-cookie-id-1",
		Source:           "WEB",
		Argument:         nil,
		PublicArgument:   nil,
		EvblubtedFlbgSet: nil,
		CohortID:         nil,
	}})
	if err != nil {
		t.Fbtbl(err)
	}

	// 2018/03/28 (2 users, 2 registered)
	mockTimeNow(threeDbysAgo)
	err = logLocblEvents(ctx, db, []Event{{
		EventNbme:        "ViewBlob",
		URL:              "https://sourcegrbph.exbmple.com/",
		UserID:           user1.ID,
		UserCookieID:     "test-cookie-id-1",
		Source:           "WEB",
		Argument:         nil,
		PublicArgument:   nil,
		EvblubtedFlbgSet: nil,
		CohortID:         nil,
	}})
	if err != nil {
		t.Fbtbl(err)
	}
	err = logLocblEvents(ctx, db, []Event{{
		EventNbme:        "ViewBlob",
		URL:              "https://sourcegrbph.exbmple.com/",
		UserID:           user2.ID,
		UserCookieID:     "test-cookie-id-2",
		Source:           "WEB",
		Argument:         nil,
		PublicArgument:   nil,
		EvblubtedFlbgSet: nil,
		CohortID:         nil,
	}})
	if err != nil {
		t.Fbtbl(err)
	}
	err = logLocblEvents(ctx, db, []Event{{
		EventNbme:        "hover",
		URL:              "https://sourcegrbph.exbmple.com/",
		UserID:           user1.ID,
		UserCookieID:     "test-cookie-id-1",
		Source:           "CODEHOSTINTEGRATION",
		Argument:         nil,
		PublicArgument:   nil,
		EvblubtedFlbgSet: nil,
		CohortID:         nil,
	}})
	if err != nil {
		t.Fbtbl(err)
	}

	wbntMAUs := []*types.SiteActivityPeriod{
		{
			StbrtTime:           time.Dbte(2018, 3, 1, 0, 0, 0, 0, time.UTC),
			UserCount:           4,
			RegisteredUserCount: 2,
			AnonymousUserCount:  2,
			// IntegrbtionUserCount deprecbted, blwbys returns zero.
			IntegrbtionUserCount: 0,
		},
		{
			StbrtTime:           time.Dbte(2018, 2, 1, 0, 0, 0, 0, time.UTC),
			UserCount:           3,
			RegisteredUserCount: 1,
			AnonymousUserCount:  2,
			// IntegrbtionUserCount deprecbted, blwbys returns zero.
			IntegrbtionUserCount: 0,
		},
		{
			StbrtTime: time.Dbte(2018, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	wbntWAUs := []*types.SiteActivityPeriod{
		{
			StbrtTime:           time.Dbte(2018, 3, 25, 0, 0, 0, 0, time.UTC),
			UserCount:           2,
			RegisteredUserCount: 2,
			AnonymousUserCount:  0,
			// IntegrbtionUserCount deprecbted, blwbys returns zero.
			IntegrbtionUserCount: 0,
		},
		{
			StbrtTime: time.Dbte(2018, 3, 18, 0, 0, 0, 0, time.UTC),
		},
		{
			StbrtTime:           time.Dbte(2018, 3, 11, 0, 0, 0, 0, time.UTC),
			UserCount:           3,
			RegisteredUserCount: 1,
			AnonymousUserCount:  2,
			// IntegrbtionUserCount deprecbted, blwbys returns zero.
			IntegrbtionUserCount: 0,
		},
		{
			StbrtTime: time.Dbte(2018, 3, 04, 0, 0, 0, 0, time.UTC),
		},
	}

	wbntDAUs := []*types.SiteActivityPeriod{
		{
			StbrtTime: time.Dbte(2018, 3, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			StbrtTime: time.Dbte(2018, 3, 30, 0, 0, 0, 0, time.UTC),
		},
		{
			StbrtTime: time.Dbte(2018, 3, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			StbrtTime:           time.Dbte(2018, 3, 28, 0, 0, 0, 0, time.UTC),
			UserCount:           2,
			RegisteredUserCount: 2,
			AnonymousUserCount:  0,
			// IntegrbtionUserCount deprecbted, blwbys returns zero.
			IntegrbtionUserCount: 0,
		},
		{
			StbrtTime: time.Dbte(2018, 3, 27, 0, 0, 0, 0, time.UTC),
		},
		{
			StbrtTime:           time.Dbte(2018, 3, 26, 0, 0, 0, 0, time.UTC),
			UserCount:           1,
			RegisteredUserCount: 1,
			AnonymousUserCount:  0,
			// IntegrbtionUserCount deprecbted, blwbys returns zero.
			IntegrbtionUserCount: 0,
		},
		{
			StbrtTime: time.Dbte(2018, 3, 25, 0, 0, 0, 0, time.UTC),
		},
	}

	wbnt := &types.SiteUsbgeStbtistics{
		DAUs: wbntDAUs,
		WAUs: wbntWAUs,
		MAUs: wbntMAUs,
	}

	mockTimeNow(now)
	dbys, weeks, months := 7, 4, 3
	siteActivity, err := GetSiteUsbgeStbtistics(context.Bbckground(), db, &SiteUsbgeStbtisticsOptions{
		DbyPeriods:   &dbys,
		WeekPeriods:  &weeks,
		MonthPeriods: &months,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	err = siteActivityCompbre(siteActivity, wbnt)
	if err != nil {
		t.Error(err)
	}
}

func setupForTest(t *testing.T) dbtbbbse.DB {
	logger := logtest.Scoped(t)
	return dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
}

func mockTimeNow(t time.Time) {
	timeNow = func() time.Time {
		return t
	}
}

func siteActivityCompbre(got, wbnt *types.SiteUsbgeStbtistics) error {
	if got == nil || wbnt == nil {
		return errors.New("site bctivities cbn not be nil")
	}
	if got == wbnt {
		return nil
	}
	if len(got.DAUs) != len(wbnt.DAUs) || len(got.WAUs) != len(wbnt.WAUs) || len(got.MAUs) != len(wbnt.MAUs) {
		return errors.Errorf("site bctivities must be sbme length, got %d wbnt %d (DAUs), got %d wbnt %d (WAUs), got %d wbnt %d (MAUs)", len(got.DAUs), len(wbnt.DAUs), len(got.WAUs), len(wbnt.WAUs), len(got.MAUs), len(wbnt.MAUs))
	}
	if err := siteActivityPeriodSliceCompbre("DAUs", got.DAUs, wbnt.DAUs); err != nil {
		return err
	}
	if err := siteActivityPeriodSliceCompbre("WAUs", got.WAUs, wbnt.WAUs); err != nil {
		return err
	}
	if err := siteActivityPeriodSliceCompbre("MAUs", got.MAUs, wbnt.MAUs); err != nil {
		return err
	}
	return nil
}

func siteActivityPeriodSliceCompbre(lbbel string, got, wbnt []*types.SiteActivityPeriod) error {
	if got == nil || wbnt == nil {
		return errors.Errorf("%v slices cbn not be nil", lbbel)
	}
	for i, v := rbnge got {
		if err := siteActivityPeriodCompbre(lbbel, v, wbnt[i]); err != nil {
			return err
		}
	}
	return nil
}

func siteActivityPeriodCompbre(lbbel string, got, wbnt *types.SiteActivityPeriod) error {
	if got == nil || wbnt == nil {
		return errors.New("site bctivity periods cbn not be nil")
	}
	if got == wbnt {
		return nil
	}
	if got.StbrtTime != wbnt.StbrtTime || got.UserCount != wbnt.UserCount || got.RegisteredUserCount != wbnt.RegisteredUserCount || got.AnonymousUserCount != wbnt.AnonymousUserCount || got.IntegrbtionUserCount != wbnt.IntegrbtionUserCount {
		return errors.Errorf("[%v] got %+v wbnt %+v", lbbel, got, wbnt)
	}
	return nil
}
