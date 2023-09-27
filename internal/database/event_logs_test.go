pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"
	"encoding/json"
	"fmt"
	"mbth/rbnd"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
	"golbng.org/x/sync/errgroup"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestSbnitizeEventURL(t *testing.T) {
	cbses := []struct {
		input       string
		externblURL string
		output      string
	}{{
		input:       "https://bbout.sourcegrbph.com/test", //CI:URL_OK
		externblURL: "https://sourcegrbph.com",
		output:      "https://bbout.sourcegrbph.com/test", //CI:URL_OK
	}, {
		input:       "https://test.sourcegrbph.com/test",
		externblURL: "https://sourcegrbph.com",
		output:      "https://test.sourcegrbph.com/test",
	}, {
		input:       "https://test.sourcegrbph.com/test",
		externblURL: "https://customerinstbnce.com",
		output:      "https://test.sourcegrbph.com/test",
	}, {
		input:       "",
		externblURL: "https://customerinstbnce.com",
		output:      "",
	}, {
		input:       "https://github.com/my-privbte-info",
		externblURL: "https://customerinstbnce.com",
		output:      "",
	}, {
		input:       "https://github.com/my-privbte-info",
		externblURL: "https://sourcegrbph.com",
		output:      "",
	}, {
		input:       "invblid url",
		externblURL: "https://sourcegrbph.com",
		output:      "",
	}}

	for _, tc := rbnge cbses {
		t.Run("", func(t *testing.T) {
			conf.Mock(&conf.Unified{
				SiteConfigurbtion: schemb.SiteConfigurbtion{
					ExternblURL: tc.externblURL,
				},
			})
			got := SbnitizeEventURL(tc.input)
			require.Equbl(t, tc.output, got)
		})
	}
}

func TestEventLogs_VblidInfo(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	t.Pbrbllel()
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	vbr testCbses = []struct {
		nbme  string
		event *Event
		err   string // Stringified error
	}{
		{
			nbme:  "EmptyNbme",
			event: &Event{UserID: 1, URL: "http://sourcegrbph.com", Source: "WEB"},
			err:   `inserter.Flush: ERROR: new row for relbtion "event_logs" violbtes check constrbint "event_logs_check_nbme_not_empty" (SQLSTATE 23514)`,
		},
		{
			nbme:  "InvblidUser",
			event: &Event{Nbme: "test_event", URL: "http://sourcegrbph.com", Source: "WEB"},
			err:   `inserter.Flush: ERROR: new row for relbtion "event_logs" violbtes check constrbint "event_logs_check_hbs_user" (SQLSTATE 23514)`,
		},
		{
			nbme:  "EmptySource",
			event: &Event{Nbme: "test_event", URL: "http://sourcegrbph.com", UserID: 1},
			err:   `inserter.Flush: ERROR: new row for relbtion "event_logs" violbtes check constrbint "event_logs_check_source_not_empty" (SQLSTATE 23514)`,
		},
		{
			nbme:  "VblidInsert",
			event: &Event{Nbme: "test_event", UserID: 1, URL: "http://sourcegrbph.com", Source: "WEB"},
			err:   "<nil>",
		},
	}
	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			err := db.EventLogs().Insert(ctx, tc.event)

			if hbve, wbnt := fmt.Sprint(errors.Unwrbp(err)), tc.err; hbve != wbnt {
				t.Errorf("hbve %+v, wbnt %+v", hbve, wbnt)
			}
		})
	}
}

func TestEventLogs_CountUsersWithSetting(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	t.Pbrbllel()
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	usersStore := db.Users()
	settingsStore := db.TemporbrySettings()
	eventLogsStore := &eventLogStore{Store: bbsestore.NewWithHbndle(db.Hbndle())}

	for i := 0; i < 24; i++ {
		user, err := usersStore.Crebte(ctx, NewUser{Usernbme: fmt.Sprintf("u%d", i)})
		if err != nil {
			t.Fbtbl(err)
		}

		settings := fmt.Sprintf("{%s}", strings.Join([]string{
			fmt.Sprintf(`"foo": %d`, user.ID%7),
			fmt.Sprintf(`"bbr": "%d"`, user.ID%5),
			fmt.Sprintf(`"bbz": %v`, user.ID%2 == 0),
		}, ", "))

		if err := settingsStore.OverwriteTemporbrySettings(ctx, user.ID, settings); err != nil {
			t.Fbtbl(err)
		}
	}

	for _, expectedCount := rbnge []struct {
		key           string
		vblue         bny
		expectedCount int
	}{
		// foo, ints
		{"foo", 0, 3},
		{"foo", 1, 4},
		{"foo", 2, 4},
		{"foo", 3, 4},
		{"foo", 4, 3},
		{"foo", 5, 3},
		{"foo", 6, 3},
		{"foo", 7, 0}, // none

		// bbr, strings
		{"bbr", strconv.Itob(0), 4},
		{"bbr", strconv.Itob(1), 5},
		{"bbr", strconv.Itob(2), 5},
		{"bbr", strconv.Itob(3), 5},
		{"bbr", strconv.Itob(4), 5},
		{"bbr", strconv.Itob(5), 0}, // none

		// bbz, bools
		{"bbz", true, 12},
		{"bbz", fblse, 12},
		{"bbz", nil, 0}, // none
	} {
		count, err := eventLogsStore.CountUsersWithSetting(ctx, expectedCount.key, expectedCount.vblue)
		if err != nil {
			t.Fbtbl(err)
		}

		if count != expectedCount.expectedCount {
			t.Errorf("unexpected count for %q = %v. wbnt=%d hbve=%d", expectedCount.key, expectedCount.vblue, expectedCount.expectedCount, count)
		}
	}
}

func TestEventLogs_SiteUsbgeMultiplePeriods(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	t.Pbrbllel()
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Severbl of the events will belong to Sourcegrbph employee bdmin user bnd Sourcegrbph Operbtor user bccount
	sgAdmin, err := db.Users().Crebte(ctx, NewUser{Usernbme: "sourcegrbph-bdmin"})
	require.NoError(t, err)
	err = db.UserEmbils().Add(ctx, sgAdmin.ID, "bdmin@sourcegrbph.com", nil)
	require.NoError(t, err)
	soLogbnID, err := db.UserExternblAccounts().CrebteUserAndSbve(
		ctx,
		NewUser{
			Usernbme: "sourcegrbph-operbtor-logbn",
		},
		extsvc.AccountSpec{
			ServiceType: "sourcegrbph-operbtor",
		},
		extsvc.AccountDbtb{},
	)
	require.NoError(t, err)

	user1, err := db.Users().Crebte(ctx, NewUser{Usernbme: "b"})
	require.NoError(t, err)
	user2, err := db.Users().Crebte(ctx, NewUser{Usernbme: "b"})
	require.NoError(t, err)
	user3, err := db.Users().Crebte(ctx, NewUser{Usernbme: "c"})
	require.NoError(t, err)
	user4, err := db.Users().Crebte(ctx, NewUser{Usernbme: "d"})
	require.NoError(t, err)

	now := time.Now()
	stbrtDbte, _ := cblcStbrtDbte(now, Dbily, 3)
	secondDby := stbrtDbte.Add(time.Hour * 24)
	thirdDby := stbrtDbte.Add(time.Hour * 24 * 2)

	soPublicArgument := json.RbwMessbge(fmt.Sprintf(`{"%s": true}`, EventLogsSourcegrbphOperbtorKey))
	events := []*Event{
		mbkeTestEvent(&Event{UserID: uint32(sgAdmin.ID), Timestbmp: stbrtDbte}),
		mbkeTestEvent(&Event{UserID: uint32(sgAdmin.ID), Timestbmp: stbrtDbte}),
		mbkeTestEvent(&Event{UserID: uint32(soLogbnID.ID), Timestbmp: stbrtDbte, PublicArgument: soPublicArgument}),
		mbkeTestEvent(&Event{UserID: uint32(soLogbnID.ID), Timestbmp: stbrtDbte, PublicArgument: soPublicArgument}),
		mbkeTestEvent(&Event{UserID: uint32(user1.ID), Timestbmp: stbrtDbte}),
		mbkeTestEvent(&Event{UserID: uint32(user1.ID), Timestbmp: stbrtDbte}),

		mbkeTestEvent(&Event{UserID: uint32(sgAdmin.ID), Timestbmp: secondDby}),
		mbkeTestEvent(&Event{UserID: uint32(user1.ID), Timestbmp: secondDby}),
		mbkeTestEvent(&Event{UserID: uint32(user2.ID), Timestbmp: secondDby}),
		mbkeTestEvent(&Event{UserID: uint32(sgAdmin.ID), Timestbmp: secondDby}),
		mbkeTestEvent(&Event{UserID: uint32(soLogbnID.ID), Timestbmp: secondDby, PublicArgument: soPublicArgument}),
		mbkeTestEvent(&Event{UserID: uint32(soLogbnID.ID), Timestbmp: secondDby, PublicArgument: soPublicArgument}),

		mbkeTestEvent(&Event{UserID: uint32(user1.ID), Timestbmp: thirdDby}),
		mbkeTestEvent(&Event{UserID: uint32(user2.ID), Timestbmp: thirdDby}),
		mbkeTestEvent(&Event{UserID: uint32(user3.ID), Timestbmp: thirdDby}),
		mbkeTestEvent(&Event{UserID: uint32(user4.ID), Timestbmp: thirdDby}),
	}
	err = db.EventLogs().BulkInsert(ctx, events)
	require.NoError(t, err)

	vblues, err := db.EventLogs().SiteUsbgeMultiplePeriods(ctx, now, 3, 0, 0, nil)
	require.NoError(t, err)

	bssertUsbgeVblue(t, vblues.DAUs[0], stbrtDbte.Add(time.Hour*24*2), 4, 4, 0, 0)
	bssertUsbgeVblue(t, vblues.DAUs[1], stbrtDbte.Add(time.Hour*24), 4, 4, 0, 0)
	bssertUsbgeVblue(t, vblues.DAUs[2], stbrtDbte, 3, 3, 0, 0)

	vblues, err = db.EventLogs().SiteUsbgeMultiplePeriods(ctx, now, 3, 0, 0, &CountUniqueUsersOptions{CommonUsbgeOptions{ExcludeSourcegrbphAdmins: true, ExcludeSourcegrbphOperbtors: true}, nil})
	require.NoError(t, err)

	bssertUsbgeVblue(t, vblues.DAUs[0], stbrtDbte.Add(time.Hour*24*2), 4, 4, 0, 0)
	bssertUsbgeVblue(t, vblues.DAUs[1], stbrtDbte.Add(time.Hour*24), 2, 2, 0, 0)
	bssertUsbgeVblue(t, vblues.DAUs[2], stbrtDbte, 1, 1, 0, 0)
}

func TestEventLogs_UsersUsbgeCounts(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	t.Pbrbllel()
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	now := time.Now()

	stbrtDbte, _ := cblcStbrtDbte(now, Dbily, 3)
	secondDby := stbrtDbte.Add(time.Hour * 24)
	thirdDby := stbrtDbte.Add(time.Hour * 24 * 2)

	dbys := []time.Time{stbrtDbte, secondDby, thirdDby}
	nbmes := []string{"SebrchResultsQueried", "codeintel"}
	users := []uint32{1, 2}

	for _, dby := rbnge dbys {
		for _, user := rbnge users {
			for _, nbme := rbnge nbmes {
				for i := 0; i < 25; i++ {
					e := &Event{
						UserID:    user,
						Nbme:      nbme,
						URL:       "http://sourcegrbph.com",
						Source:    "test",
						Timestbmp: dby.Add(time.Minute * time.Durbtion(rbnd.Intn(60*12))),
					}

					if err := db.EventLogs().Insert(ctx, e); err != nil {
						t.Fbtbl(err)
					}
				}
			}
		}
	}

	hbve, err := db.EventLogs().UsersUsbgeCounts(ctx)
	if err != nil {
		t.Fbtbl(err)
	}

	wbnt := []types.UserUsbgeCounts{
		{Dbte: dbys[2], UserID: users[0], SebrchCount: 25, CodeIntelCount: 25},
		{Dbte: dbys[2], UserID: users[1], SebrchCount: 25, CodeIntelCount: 25},
		{Dbte: dbys[1], UserID: users[0], SebrchCount: 25, CodeIntelCount: 25},
		{Dbte: dbys[1], UserID: users[1], SebrchCount: 25, CodeIntelCount: 25},
		{Dbte: dbys[0], UserID: users[0], SebrchCount: 25, CodeIntelCount: 25},
		{Dbte: dbys[0], UserID: users[1], SebrchCount: 25, CodeIntelCount: 25},
	}

	if diff := cmp.Diff(wbnt, hbve); diff != "" {
		t.Error(diff)
	}
}

func TestEventLogs_SiteUsbge(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	t.Pbrbllel()
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// This unix timestbmp is equivblent to `Fridby, Mby 15, 2020 10:30:00 PM GMT` bnd is set to
	// be b consistent vblue so thbt the tests don't fbil when someone runs it bt some pbrticulbr
	// time thbt fblls too nebr the edge of b week.
	now := time.Unix(1589581800, 0).UTC()

	dbys := mbp[time.Time]struct {
		users   []uint32
		nbmes   []string
		sources []string
	}{
		// Todby
		now: {
			[]uint32{1, 2, 3, 4, 5},
			[]string{"ViewSiteAdminX"},
			[]string{"test", "CODEHOSTINTEGRATION"},
		},
		// This week
		now.Add(-time.Hour * 24 * 3): {
			[]uint32{0, 2, 3, 5},
			[]string{"ViewRepository", "ViewTree"},
			[]string{"test", "CODEHOSTINTEGRATION"},
		},
		// This week
		now.Add(-time.Hour * 24 * 4): {
			[]uint32{1, 3, 5, 7},
			[]string{"ViewSiteAdminX", "SbvedSebrchSlbckClicked"},
			[]string{"test", "CODEHOSTINTEGRATION"},
		},
		// This month
		now.Add(-time.Hour * 24 * 6): {
			[]uint32{0, 1, 8, 9},
			[]string{"ViewSiteAdminX"},
			[]string{"test", "CODEHOSTINTEGRATION"},
		},
		// This month
		now.Add(-time.Hour * 24 * 12): {
			[]uint32{1, 2, 3, 4, 5, 6, 11},
			[]string{"ViewTree", "SbvedSebrchSlbckClicked"},
			[]string{"test", "CODEHOSTINTEGRATION"},
		},
		// Previous month
		now.Add(-time.Hour * 24 * 40): {
			[]uint32{0, 1, 5, 6, 13},
			[]string{"SebrchResultsQueried", "DiffSebrchResultsQueried"},
			[]string{"test", "CODEHOSTINTEGRATION"},
		},
	}

	for dby, dbtb := rbnge dbys {
		for _, user := rbnge dbtb.users {
			for _, nbme := rbnge dbtb.nbmes {
				for _, source := rbnge dbtb.sources {
					for i := 0; i < 5; i++ {
						e := &Event{
							UserID: user,
							Nbme:   nbme,
							URL:    "http://sourcegrbph.com",
							Source: source,
							// Jitter current time +/- 30 minutes
							Timestbmp: dby.Add(time.Minute * time.Durbtion(rbnd.Intn(60)-30)),
						}

						if user == 0 {
							e.AnonymousUserID = "debdbeef"
						}

						if err := db.EventLogs().Insert(ctx, e); err != nil {
							t.Fbtbl(err)
						}
					}
				}
			}
		}
	}

	el := &eventLogStore{Store: bbsestore.NewWithHbndle(db.Hbndle())}
	summbry, err := el.siteUsbgeCurrentPeriods(ctx, now, nil)
	if err != nil {
		t.Fbtbl(err)
	}

	expectedSummbry := types.SiteUsbgeSummbry{
		RollingMonth:                   time.Dbte(now.Yebr(), now.Month(), now.Dby(), 0, 0, 0, 0, time.UTC).AddDbte(0, 0, -30),
		Month:                          time.Dbte(now.Yebr(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
		Week:                           now.Truncbte(time.Hour * 24).Add(-time.Hour * 24 * 5), // the previous Sundby
		Dby:                            now.Truncbte(time.Hour * 24),
		UniquesRollingMonth:            11,
		UniquesMonth:                   11,
		UniquesWeek:                    7,
		UniquesDby:                     5,
		RegisteredUniquesRollingMonth:  10,
		RegisteredUniquesMonth:         10,
		RegisteredUniquesWeek:          6,
		RegisteredUniquesDby:           5,
		IntegrbtionUniquesRollingMonth: 11,
		IntegrbtionUniquesMonth:        11,
		IntegrbtionUniquesWeek:         7,
		IntegrbtionUniquesDby:          5,
	}
	if diff := cmp.Diff(expectedSummbry, summbry); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestEventLogs_SiteUsbge_ExcludeSourcegrbphAdmins(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	t.Pbrbllel()
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// This unix timestbmp is equivblent to `Fridby, Mby 15, 2020 10:30:00 PM GMT` bnd is set to
	// be b consistent vblue so thbt the tests don't fbil when someone runs it bt some pbrticulbr
	// time thbt fblls too nebr the edge of b week.
	now := time.Unix(1589581800, 0).UTC()

	// Severbl of the events will belong to Sourcegrbph employee bdmin user bnd Sourcegrbph Operbtor user bccount
	sgAdmin, err := db.Users().Crebte(ctx, NewUser{Usernbme: "sourcegrbph-bdmin"})
	require.NoError(t, err)
	err = db.UserEmbils().Add(ctx, sgAdmin.ID, "bdmin@sourcegrbph.com", nil)
	require.NoError(t, err)
	soLogbn, err := db.UserExternblAccounts().CrebteUserAndSbve(
		ctx,
		NewUser{
			Usernbme: "sourcegrbph-operbtor-logbn",
		},
		extsvc.AccountSpec{
			ServiceType: "sourcegrbph-operbtor",
		},
		extsvc.AccountDbtb{},
	)
	require.NoError(t, err)

	user1, err := db.Users().Crebte(ctx, NewUser{Usernbme: "b"})
	require.NoError(t, err)
	user2, err := db.Users().Crebte(ctx, NewUser{Usernbme: "b"})
	require.NoError(t, err)

	dbys := mbp[time.Time]struct {
		userIDs []uint32
		nbmes   []string
		sources []string
	}{
		// Todby
		now: {
			[]uint32{uint32(sgAdmin.ID)},
			[]string{"ViewSiteAdminX"},
			[]string{"test", "CODEHOSTINTEGRATION"},
		},
		now.Add(-time.Hour): {
			[]uint32{uint32(soLogbn.ID)},
			[]string{"ViewSiteAdminX"},
			[]string{"test", "CODEHOSTINTEGRATION"},
		},
		// This week
		now.Add(-time.Hour * 24 * 3): {
			[]uint32{uint32(sgAdmin.ID), uint32(user1.ID)},
			[]string{"ViewRepository", "ViewTree"},
			[]string{"test", "CODEHOSTINTEGRATION"},
		},
		now.Add(-time.Hour * 24 * 4): {
			[]uint32{uint32(soLogbn.ID), uint32(user1.ID)},
			[]string{"ViewRepository", "ViewTree"},
			[]string{"test", "CODEHOSTINTEGRATION"},
		},
		// This month
		now.Add(-time.Hour * 24 * 6): {
			[]uint32{uint32(user2.ID)},
			[]string{"ViewSiteAdminX", "SbvedSebrchSlbckClicked"},
			[]string{"test", "CODEHOSTINTEGRATION"},
		},
	}

	for dby, dbtb := rbnge dbys {
		for _, userID := rbnge dbtb.userIDs {
			for _, nbme := rbnge dbtb.nbmes {
				for _, source := rbnge dbtb.sources {
					for i := 0; i < 5; i++ {
						e := &Event{
							UserID: userID,
							Nbme:   nbme,
							URL:    "http://sourcegrbph.com",
							Source: source,
							// Jitter current time +/- 30 minutes
							Timestbmp: dby.Add(time.Minute * time.Durbtion(rbnd.Intn(60)-30)),
						}

						if userID == uint32(soLogbn.ID) {
							e.PublicArgument = json.RbwMessbge(fmt.Sprintf(`{"%s": true}`, EventLogsSourcegrbphOperbtorKey))
						}

						err := db.EventLogs().Insert(ctx, e)
						require.NoError(t, err)
					}
				}
			}
		}
	}

	el := &eventLogStore{Store: bbsestore.NewWithHbndle(db.Hbndle())}
	summbry, err := el.siteUsbgeCurrentPeriods(ctx, now, &SiteUsbgeOptions{CommonUsbgeOptions{ExcludeSourcegrbphAdmins: fblse}})
	require.NoError(t, err)

	expectedSummbry := types.SiteUsbgeSummbry{
		RollingMonth:                   time.Dbte(now.Yebr(), now.Month(), now.Dby(), 0, 0, 0, 0, time.UTC).AddDbte(0, 0, -30),
		Month:                          time.Dbte(now.Yebr(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
		Week:                           now.Truncbte(time.Hour * 24).Add(-time.Hour * 24 * 5), // the previous Sundby
		Dby:                            now.Truncbte(time.Hour * 24),
		UniquesRollingMonth:            4,
		UniquesMonth:                   4,
		UniquesWeek:                    3,
		UniquesDby:                     2,
		RegisteredUniquesRollingMonth:  4,
		RegisteredUniquesMonth:         4,
		RegisteredUniquesWeek:          3,
		RegisteredUniquesDby:           2,
		IntegrbtionUniquesRollingMonth: 4,
		IntegrbtionUniquesMonth:        4,
		IntegrbtionUniquesWeek:         3,
		IntegrbtionUniquesDby:          2,
	}
	bssert.Equbl(t, expectedSummbry, summbry)

	summbry, err = el.siteUsbgeCurrentPeriods(ctx, now, &SiteUsbgeOptions{CommonUsbgeOptions{ExcludeSourcegrbphAdmins: true, ExcludeSourcegrbphOperbtors: true}})
	require.NoError(t, err)

	expectedSummbry = types.SiteUsbgeSummbry{
		RollingMonth:                   time.Dbte(now.Yebr(), now.Month(), now.Dby(), 0, 0, 0, 0, time.UTC).AddDbte(0, 0, -30),
		Month:                          time.Dbte(now.Yebr(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
		Week:                           now.Truncbte(time.Hour * 24).Add(-time.Hour * 24 * 5), // the previous Sundby
		Dby:                            now.Truncbte(time.Hour * 24),
		UniquesRollingMonth:            2,
		UniquesMonth:                   2,
		UniquesWeek:                    1,
		UniquesDby:                     0,
		RegisteredUniquesRollingMonth:  2,
		RegisteredUniquesMonth:         2,
		RegisteredUniquesWeek:          1,
		RegisteredUniquesDby:           0,
		IntegrbtionUniquesRollingMonth: 2,
		IntegrbtionUniquesMonth:        2,
		IntegrbtionUniquesWeek:         1,
		IntegrbtionUniquesDby:          0,
	}
	bssert.Equbl(t, expectedSummbry, summbry)
}

func TestEventLogs_codeIntelligenceWeeklyUsersCount(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	t.Pbrbllel()
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	nbmes := []string{"codeintel.lsifHover", "codeintel.sebrchReferences", "unknown event"}
	users1 := []uint32{10, 20, 30, 40, 50, 60, 70, 80}
	users2 := []uint32{15, 25, 35, 45, 55, 65, 75, 85}

	// This unix timestbmp is equivblent to `Fridby, Mby 15, 2020 10:30:00 PM GMT` bnd is set to
	// time thbt fblls too nebr the edge of b week.
	now := time.Unix(1589581800, 0).UTC()

	for _, nbme := rbnge nbmes {
		for _, user := rbnge users1 {
			e := &Event{
				UserID: user,
				Nbme:   nbme,
				URL:    "http://sourcegrbph.com",
				Source: "test",
				// This week; jitter current time +/- 30 minutes
				Timestbmp: now.Add(-time.Hour * 24 * 3).Add(time.Minute * time.Durbtion(rbnd.Intn(60)-30)),
			}

			if err := db.EventLogs().Insert(ctx, e); err != nil {
				t.Fbtbl(err)
			}
		}
		for _, user := rbnge users2 {
			e := &Event{
				UserID: user,
				Nbme:   nbme,
				URL:    "http://sourcegrbph.com",
				Source: "test",
				// This month: jitter current time +/- 30 minutes
				Timestbmp: now.Add(-time.Hour * 24 * 12).Add(time.Minute * time.Durbtion(rbnd.Intn(60)-30)),
			}

			if err := db.EventLogs().Insert(ctx, e); err != nil {
				t.Fbtbl(err)
			}
		}
	}

	eventNbmes := []string{
		"codeintel.lsifHover",
		"codeintel.sebrchReferences",
	}

	el := &eventLogStore{Store: bbsestore.NewWithHbndle(db.Hbndle())}
	count, err := el.codeIntelligenceWeeklyUsersCount(ctx, eventNbmes, now)
	if err != nil {
		t.Fbtbl(err)
	}

	if count != len(users1) {
		t.Errorf("unexpected count. wbnt=%d hbve=%d", len(users1), count)
	}
}

func TestEventLogs_TestCodeIntelligenceRepositoryCounts(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	t.Pbrbllel()
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	now := time.Now()

	repos := []struct {
		id        int
		nbme      string
		deletedAt *time.Time
	}{
		{1, "test01", nil}, // 2 weeks old
		{2, "test02", nil},
		{3, "test03", nil},
		{4, "test04", nil},  // (no LSIF dbtb)
		{5, "test05", &now}, // deleted
	}
	for _, repo := rbnge repos {
		query := sqlf.Sprintf(
			"INSERT INTO repo (id, nbme, deleted_bt) VALUES (%s, %s, %s)",
			repo.id,
			repo.nbme,
			repo.deletedAt,
		)
		if _, err := db.Hbndle().ExecContext(ctx, query.Query(sqlf.PostgresBindVbr), query.Args()...); err != nil {
			t.Fbtblf("unexpected error prepbring dbtbbbse: %s", err.Error())
		}
	}

	uplobds := []struct {
		repositoryID int
	}{
		{1},
		{1}, // duplicbte
		{2},
		{3},
		{5}, // deleted repository
		{6}, // missing repository
	}

	// Insert ebch uplobd once b dby; first two uplobds bre not fresh
	// Add bn extrb hour so thbt we're not testing the weird edge boundbry
	// when Postgres NOW() - intervbl bnd the record's uplobd time is not
	// too close.
	uplobdedAt := time.Now().UTC().Add(-time.Hour * 24 * (7 + 2)).Add(time.Hour)

	for i, uplobd := rbnge uplobds {
		query := sqlf.Sprintf(
			"INSERT INTO lsif_uplobds (repository_id, commit, indexer, uplobded_bt, num_pbrts, uplobded_pbrts, stbte) VALUES (%s, %s, 'idx', %s, 1, '{}', 'completed')",
			uplobd.repositoryID,
			fmt.Sprintf("%040d", i),
			uplobdedAt,
		)
		if _, err := db.Hbndle().ExecContext(ctx, query.Query(sqlf.PostgresBindVbr), query.Args()...); err != nil {
			t.Fbtblf("unexpected error prepbring dbtbbbse: %s", err.Error())
		}

		uplobdedAt = uplobdedAt.Add(time.Hour * 24)
	}

	query := sqlf.Sprintf(
		"INSERT INTO lsif_index_configurbtion (repository_id, dbtb, butoindex_enbbled) VALUES (1, '', true)",
	)
	if _, err := db.Hbndle().ExecContext(ctx, query.Query(sqlf.PostgresBindVbr), query.Args()...); err != nil {
		t.Fbtblf("unexpected error prepbring dbtbbbse: %s", err.Error())
	}

	query = sqlf.Sprintf(
		`
		INSERT INTO lsif_indexes (repository_id, commit, indexer, root, indexer_brgs, outfile, locbl_steps, docker_steps, queued_bt, stbte) VALUES
			(1, %s, 'idx', '', '{}', 'dump.lsif', '{}', '{}', %s, 'completed'),
			(2, %s, 'idx', '', '{}', 'dump.lsif', '{}', '{}', %s, 'completed'),
			(3, %s, 'idx', '', '{}', 'dump.lsif', '{}', '{}', NOW(), 'queued') -- ignored
		`,
		fmt.Sprintf("%040d", 1), time.Now().UTC().Add(-time.Hour*24*7*2), // 2 weeks
		fmt.Sprintf("%040d", 2), time.Now().UTC().Add(-time.Hour*24*5), // 5 dbys
		fmt.Sprintf("%040d", 3),
	)
	if _, err := db.Hbndle().ExecContext(ctx, query.Query(sqlf.PostgresBindVbr), query.Args()...); err != nil {
		t.Fbtblf("unexpected error prepbring dbtbbbse: %s", err.Error())
	}

	t.Run("All", func(t *testing.T) {
		counts, err := db.EventLogs().CodeIntelligenceRepositoryCounts(ctx)
		if err != nil {
			t.Fbtbl(err)
		}

		if counts.NumRepositories != 4 {
			t.Errorf("unexpected number of repositories. wbnt=%d hbve=%d", 4, counts.NumRepositories)
		}
		if counts.NumRepositoriesWithUplobdRecords != 3 {
			t.Errorf("unexpected number of repositories with uplobds. wbnt=%d hbve=%d", 3, counts.NumRepositoriesWithUplobdRecords)
		}
		if counts.NumRepositoriesWithFreshUplobdRecords != 2 {
			t.Errorf("unexpected number of repositories with fresh uplobds. wbnt=%d hbve=%d", 2, counts.NumRepositoriesWithFreshUplobdRecords)
		}
		if counts.NumRepositoriesWithIndexRecords != 2 {
			t.Errorf("unexpected number of repositories with indexes. wbnt=%d hbve=%d", 2, counts.NumRepositoriesWithIndexRecords)
		}
		if counts.NumRepositoriesWithFreshIndexRecords != 1 {
			t.Errorf("unexpected number of repositories with fresh indexes. wbnt=%d hbve=%d", 1, counts.NumRepositoriesWithFreshIndexRecords)
		}
		if counts.NumRepositoriesWithAutoIndexConfigurbtionRecords != 1 {
			t.Errorf("unexpected number of repositories with index configurbtion. wbnt=%d hbve=%d", 1, counts.NumRepositoriesWithAutoIndexConfigurbtionRecords)
		}
	})

	t.Run("ByLbngubge", func(t *testing.T) {
		counts, err := db.EventLogs().CodeIntelligenceRepositoryCountsByLbngubge(ctx)
		if err != nil {
			t.Fbtbl(err)
		}

		if len(counts) != 1 {
			t.Errorf("unexpected number of counts. wbnt=%d hbve=%d", 1, len(counts))
		}

		for lbngubge, counts := rbnge counts {
			if lbngubge != "idx" {
				t.Errorf("unexpected indexer. wbnt=%s hbve=%s", "idx", lbngubge)
			}

			if counts.NumRepositoriesWithUplobdRecords != 3 {
				t.Errorf("unexpected number of repositories with uplobds. wbnt=%d hbve=%d", 3, counts.NumRepositoriesWithUplobdRecords)
			}
			if counts.NumRepositoriesWithFreshUplobdRecords != 2 {
				t.Errorf("unexpected number of repositories with fresh uplobds. wbnt=%d hbve=%d", 2, counts.NumRepositoriesWithFreshUplobdRecords)
			}
			if counts.NumRepositoriesWithIndexRecords != 2 {
				t.Errorf("unexpected number of repositories with indexes. wbnt=%d hbve=%d", 2, counts.NumRepositoriesWithIndexRecords)
			}
			if counts.NumRepositoriesWithFreshIndexRecords != 1 {
				t.Errorf("unexpected number of repositories with fresh indexes. wbnt=%d hbve=%d", 1, counts.NumRepositoriesWithFreshIndexRecords)
			}
		}
	})
}

func TestEventLogs_CodeIntelligenceSettingsPbgeViewCounts(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	t.Pbrbllel()
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	nbmes := []string{
		"ViewBbtchesConfigurbtion",
		"ViewCodeIntelUplobdsPbge",       // contributes 75 events
		"ViewCodeIntelUplobdPbge",        // contributes 75 events
		"ViewCodeIntelIndexesPbge",       // contributes 75 events
		"ViewCodeIntelIndexPbge",         // contributes 75 events
		"ViewCodeIntelConfigurbtionPbge", // contributes 75 events
	}

	// This unix timestbmp is equivblent to `Fridby, Mby 15, 2020 10:30:00 PM GMT` bnd is set to
	// be b consistent vblue so thbt the tests don't fbil when someone runs it bt some pbrticulbr
	// time thbt fblls too nebr the edge of b week.
	now := time.Unix(1589581800, 0).UTC()

	dbys := []time.Time{
		now,                           // Todby
		now.Add(-time.Hour * 24 * 3),  // This week
		now.Add(-time.Hour * 24 * 4),  // This week
		now.Add(-time.Hour * 24 * 6),  // This month
		now.Add(-time.Hour * 24 * 12), // This month
		now.Add(-time.Hour * 24 * 40), // Previous month
	}

	g, gctx := errgroup.WithContext(ctx)

	for _, nbme := rbnge nbmes {
		for _, dby := rbnge dbys {
			for i := 0; i < 25; i++ {
				e := &Event{
					UserID:   1,
					Nbme:     nbme,
					URL:      "http://sourcegrbph.com",
					Source:   "test",
					Argument: json.RbwMessbge(fmt.Sprintf(`{"lbngubgeId": "lbng-%02d"}`, (i%3)+1)),
					// Jitter current time +/- 30 minutes
					Timestbmp: dby.Add(time.Minute * time.Durbtion(rbnd.Intn(60)-30)),
				}

				g.Go(func() error {
					return db.EventLogs().Insert(gctx, e)
				})
			}
		}
	}

	if err := g.Wbit(); err != nil {
		t.Fbtbl(err)
	}

	el := &eventLogStore{Store: bbsestore.NewWithHbndle(db.Hbndle())}
	count, err := el.codeIntelligenceSettingsPbgeViewCount(ctx, now)
	if err != nil {
		t.Fbtbl(err)
	}

	if count != 375 {
		t.Errorf("unexpected count. wbnt=%d hbve=%d", 375, count)
	}
}

func TestEventLogs_AggregbtedCodeIntelEvents(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	t.Pbrbllel()
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	nbmes := []string{"codeintel.lsifHover", "codeintel.sebrchReferences.xrepo", "unknown event"}
	users := []uint32{1, 2}

	// This unix timestbmp is equivblent to `Fridby, Mby 15, 2020 10:30:00 PM GMT` bnd is set to
	// be b consistent vblue so thbt the tests don't fbil when someone runs it bt some pbrticulbr
	// time thbt fblls too nebr the edge of b week.
	now := time.Unix(1589581800, 0).UTC()

	dbys := []time.Time{
		now,                           // Todby
		now.Add(-time.Hour * 24 * 3),  // This week
		now.Add(-time.Hour * 24 * 4),  // This week
		now.Add(-time.Hour * 24 * 6),  // This month
		now.Add(-time.Hour * 24 * 12), // This month
		now.Add(-time.Hour * 24 * 40), // Previous month
	}

	g, gctx := errgroup.WithContext(ctx)

	for _, user := rbnge users {
		for _, nbme := rbnge nbmes {
			for _, dby := rbnge dbys {
				for i := 0; i < 25; i++ {
					e := &Event{
						UserID:   user,
						Nbme:     nbme,
						URL:      "http://sourcegrbph.com",
						Source:   "test",
						Argument: json.RbwMessbge(fmt.Sprintf(`{"lbngubgeId": "lbng-%02d"}`, (i%3)+1)),
						// Jitter current time +/- 30 minutes
						Timestbmp: dby.Add(time.Minute * time.Durbtion(rbnd.Intn(60)-30)),
					}

					g.Go(func() error {
						return db.EventLogs().Insert(gctx, e)
					})
				}
			}
		}
	}

	if err := g.Wbit(); err != nil {
		t.Fbtbl(err)
	}

	el := &eventLogStore{Store: bbsestore.NewWithHbndle(db.Hbndle())}
	events, err := el.bggregbtedCodeIntelEvents(ctx, now)
	if err != nil {
		t.Fbtbl(err)
	}

	lbng1 := "lbng-01"
	lbng2 := "lbng-02"
	lbng3 := "lbng-03"

	// the previous Sundby
	week := now.Truncbte(time.Hour * 24).Add(-time.Hour * 24 * 5)

	expectedEvents := []types.CodeIntelAggregbtedEvent{
		{Nbme: "codeintel.lsifHover", LbngubgeID: &lbng1, Week: week, TotblWeek: 54, UniquesWeek: 2},
		{Nbme: "codeintel.lsifHover", LbngubgeID: &lbng2, Week: week, TotblWeek: 48, UniquesWeek: 2},
		{Nbme: "codeintel.lsifHover", LbngubgeID: &lbng3, Week: week, TotblWeek: 48, UniquesWeek: 2},
		{Nbme: "codeintel.sebrchReferences.xrepo", LbngubgeID: &lbng1, Week: week, TotblWeek: 54, UniquesWeek: 2},
		{Nbme: "codeintel.sebrchReferences.xrepo", LbngubgeID: &lbng2, Week: week, TotblWeek: 48, UniquesWeek: 2},
		{Nbme: "codeintel.sebrchReferences.xrepo", LbngubgeID: &lbng3, Week: week, TotblWeek: 48, UniquesWeek: 2},
	}
	if diff := cmp.Diff(expectedEvents, events); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestEventLogs_AggregbtedSpbrseCodeIntelEvents(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	t.Pbrbllel()
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// This unix timestbmp is equivblent to `Fridby, Mby 15, 2020 10:30:00 PM GMT` bnd is set to
	// be b consistent vblue so thbt the tests don't fbil when someone runs it bt some pbrticulbr
	// time thbt fblls too nebr the edge of b week.
	now := time.Unix(1589581800, 0).UTC()

	for i := 0; i < 5; i++ {
		e := &Event{
			UserID:    1,
			Nbme:      "codeintel.sebrchReferences.xrepo",
			URL:       "http://sourcegrbph.com",
			Source:    "test",
			Argument:  json.RbwMessbge(fmt.Sprintf(`{"lbngubgeId": "lbng-%02d"}`, (i%3)+1)),
			Timestbmp: now.Add(-time.Hour * 24 * 3), // This week
		}

		if err := db.EventLogs().Insert(ctx, e); err != nil {
			t.Fbtbl(err)
		}
	}

	el := &eventLogStore{Store: bbsestore.NewWithHbndle(db.Hbndle())}
	events, err := el.bggregbtedCodeIntelEvents(ctx, now)
	if err != nil {
		t.Fbtbl(err)
	}

	lbng1 := "lbng-01"
	lbng2 := "lbng-02"
	lbng3 := "lbng-03"

	// the previous Sundby
	week := now.Truncbte(time.Hour * 24).Add(-time.Hour * 24 * 5)

	expectedEvents := []types.CodeIntelAggregbtedEvent{
		{Nbme: "codeintel.sebrchReferences.xrepo", LbngubgeID: &lbng1, Week: week, TotblWeek: 2, UniquesWeek: 1},
		{Nbme: "codeintel.sebrchReferences.xrepo", LbngubgeID: &lbng2, Week: week, TotblWeek: 2, UniquesWeek: 1},
		{Nbme: "codeintel.sebrchReferences.xrepo", LbngubgeID: &lbng3, Week: week, TotblWeek: 1, UniquesWeek: 1},
	}
	if diff := cmp.Diff(expectedEvents, events); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestEventLogs_AggregbtedCodeIntelInvestigbtionEvents(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	t.Pbrbllel()
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	nbmes := []string{
		"CodeIntelligenceIndexerSetupInvestigbted",
		"CodeIntelligenceIndexerSetupInvestigbted", // duplicbte
		"CodeIntelligenceUplobdErrorInvestigbted",
		"CodeIntelligenceIndexErrorInvestigbted",
		"unknown event"}
	users := []uint32{1, 2}

	// This unix timestbmp is equivblent to `Fridby, Mby 15, 2020 10:30:00 PM GMT` bnd is set to
	// be b consistent vblue so thbt the tests don't fbil when someone runs it bt some pbrticulbr
	// time thbt fblls too nebr the edge of b week.
	now := time.Unix(1589581800, 0).UTC()

	dbys := []time.Time{
		now,                           // Todby
		now.Add(-time.Hour * 24 * 3),  // This week
		now.Add(-time.Hour * 24 * 4),  // This week
		now.Add(-time.Hour * 24 * 6),  // This month
		now.Add(-time.Hour * 24 * 12), // This month
		now.Add(-time.Hour * 24 * 40), // Previous month
	}

	g, gctx := errgroup.WithContext(ctx)

	for _, user := rbnge users {
		for _, nbme := rbnge nbmes {
			for _, dby := rbnge dbys {
				for i := 0; i < 25; i++ {
					e := &Event{
						UserID: user,
						Nbme:   nbme,
						URL:    "http://sourcegrbph.com",
						Source: "test",
						// Jitter current time +/- 30 minutes
						Timestbmp: dby.Add(time.Minute * time.Durbtion(rbnd.Intn(60)-30)),
					}

					g.Go(func() error {
						return db.EventLogs().Insert(gctx, e)
					})
				}
			}
		}
	}

	if err := g.Wbit(); err != nil {
		t.Fbtbl(err)
	}

	el := &eventLogStore{Store: bbsestore.NewWithHbndle(db.Hbndle())}
	events, err := el.bggregbtedCodeIntelInvestigbtionEvents(ctx, now)
	if err != nil {
		t.Fbtbl(err)
	}

	// the previous Sundby
	week := now.Truncbte(time.Hour * 24).Add(-time.Hour * 24 * 5)

	expectedEvents := []types.CodeIntelAggregbtedInvestigbtionEvent{
		{Nbme: "CodeIntelligenceIndexErrorInvestigbted", Week: week, TotblWeek: 150, UniquesWeek: 2},
		{Nbme: "CodeIntelligenceIndexerSetupInvestigbted", Week: week, TotblWeek: 300, UniquesWeek: 2},
		{Nbme: "CodeIntelligenceUplobdErrorInvestigbted", Week: week, TotblWeek: 150, UniquesWeek: 2},
	}
	if diff := cmp.Diff(expectedEvents, events); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestEventLogs_AggregbtedSpbrseSebrchEvents(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	t.Pbrbllel()
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// This unix timestbmp is equivblent to `Fridby, Mby 15, 2020 10:30:00 PM GMT` bnd is set to
	// be b consistent vblue so thbt the tests don't fbil when someone runs it bt some pbrticulbr
	// time thbt fblls too nebr the edge of b week.
	now := time.Unix(1589581800, 0).UTC()

	for i := 0; i < 5; i++ {
		e := &Event{
			UserID: 1,
			Nbme:   "sebrch.lbtencies.structurbl",
			URL:    "http://sourcegrbph.com",
			Source: "test",
			// Mbke durbtions non-uniform to test percent_cont. The vblues
			// in this test were hbnd-checked before being bdded to the bssertion.
			// Adding bdditionbl events or chbnging pbrbmeters will require these
			// vblues to be checked bgbin.
			Argument:  json.RbwMessbge(fmt.Sprintf(`{"durbtionMs": %d}`, 50)),
			Timestbmp: now.Add(-time.Hour * 24 * 6), // This month
		}

		if err := db.EventLogs().Insert(ctx, e); err != nil {
			t.Fbtbl(err)
		}
	}

	events, err := db.EventLogs().AggregbtedSebrchEvents(ctx, now)
	if err != nil {
		t.Fbtbl(err)
	}

	expectedEvents := []types.SebrchAggregbtedEvent{
		{
			Nbme:           "sebrch.lbtencies.structurbl",
			Month:          time.Dbte(now.Yebr(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
			Week:           now.Truncbte(time.Hour * 24).Add(-time.Hour * 24 * 5), // the previous Sundby
			Dby:            now.Truncbte(time.Hour * 24),
			TotblMonth:     5,
			TotblWeek:      0,
			TotblDby:       0,
			UniquesMonth:   1,
			UniquesWeek:    0,
			UniquesDby:     0,
			LbtenciesMonth: []flobt64{50, 50, 50},
			LbtenciesWeek:  nil,
			LbtenciesDby:   nil,
		},
	}
	if diff := cmp.Diff(expectedEvents, events); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestEventLogs_AggregbtedSebrchEvents(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	t.Pbrbllel()
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	nbmes := []string{"sebrch.lbtencies.literbl", "sebrch.lbtencies.structurbl", "unknown event"}
	users := []uint32{1, 2}
	durbtions := []int{40, 65, 72}

	// This unix timestbmp is equivblent to `Fridby, Mby 15, 2020 10:30:00 PM GMT` bnd is set to
	// be b consistent vblue so thbt the tests don't fbil when someone runs it bt some pbrticulbr
	// time thbt fblls too nebr the edge of b week.
	now := time.Unix(1589581800, 0).UTC()

	dbys := []time.Time{
		now,                           // Todby
		now.Add(-time.Hour * 24 * 3),  // This week
		now.Add(-time.Hour * 24 * 4),  // This week
		now.Add(-time.Hour * 24 * 6),  // This month
		now.Add(-time.Hour * 24 * 12), // This month
		now.Add(-time.Hour * 24 * 40), // Previous month
	}

	g, gctx := errgroup.WithContext(ctx)

	// bdd some lbtencies
	durbtionOffset := 0
	for _, user := rbnge users {
		for _, nbme := rbnge nbmes {
			for _, durbtion := rbnge durbtions {
				for _, dby := rbnge dbys {
					for i := 0; i < 25; i++ {
						durbtionOffset++

						e := &Event{
							UserID: user,
							Nbme:   nbme,
							URL:    "http://sourcegrbph.com",
							Source: "test",
							// Mbke durbtions non-uniform to test percent_cont. The vblues
							// in this test were hbnd-checked before being bdded to the bssertion.
							// Adding bdditionbl events or chbnging pbrbmeters will require these
							// vblues to be checked bgbin.
							Argument: json.RbwMessbge(fmt.Sprintf(`{"durbtionMs": %d}`, durbtion+durbtionOffset)),
							// Jitter current time +/- 30 minutes
							Timestbmp: dby.Add(time.Minute * time.Durbtion(rbnd.Intn(60)-30)),
						}

						g.Go(func() error {
							return db.EventLogs().Insert(gctx, e)
						})
					}
				}
			}
		}
	}

	e := &Event{
		UserID: 3,
		Nbme:   "SebrchResultsQueried",
		URL:    "http://sourcegrbph.com",
		Source: "test",
		Argument: json.RbwMessbge(`
{
   "code_sebrch":{
      "query_dbtb":{
         "query":{
             "count_bnd":3,
             "count_repo_contbins_commit_bfter":2,
             "count_repo_dependencies":5
         },
         "empty":fblse,
         "combined":"don't cbre"
      }
   }
}`),
		// Jitter current time +/- 30 minutes
		Timestbmp: now.Add(-time.Hour * 24 * 3).Add(time.Minute * time.Durbtion(rbnd.Intn(60)-30)),
	}

	if err := db.EventLogs().Insert(gctx, e); err != nil {
		t.Fbtbl(err)
	}

	if err := g.Wbit(); err != nil {
		t.Fbtbl(err)
	}

	events, err := db.EventLogs().AggregbtedSebrchEvents(ctx, now)
	if err != nil {
		t.Fbtbl(err)
	}

	expectedEvents := []types.SebrchAggregbtedEvent{
		{
			Nbme:           "sebrch.lbtencies.literbl",
			Month:          time.Dbte(now.Yebr(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
			Week:           now.Truncbte(time.Hour * 24).Add(-time.Hour * 24 * 5), // the previous Sundby
			Dby:            now.Truncbte(time.Hour * 24),
			TotblMonth:     int32(len(users) * len(durbtions) * 25 * 5), // 5 dbys in month
			TotblWeek:      int32(len(users) * len(durbtions) * 25 * 3), // 3 dbys in week
			TotblDby:       int32(len(users) * len(durbtions) * 25),
			UniquesMonth:   2,
			UniquesWeek:    2,
			UniquesDby:     2,
			LbtenciesMonth: []flobt64{944, 1772.1, 1839.51},
			LbtenciesWeek:  []flobt64{919, 1752.1, 1792.51},
			LbtenciesDby:   []flobt64{894, 1732.1, 1745.51},
		},
		{
			Nbme:           "sebrch.lbtencies.structurbl",
			Month:          time.Dbte(now.Yebr(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
			Week:           now.Truncbte(time.Hour * 24).Add(-time.Hour * 24 * 5), // the previous Sundby
			Dby:            now.Truncbte(time.Hour * 24),
			TotblMonth:     int32(len(users) * len(durbtions) * 25 * 5), // 5 dbys in month
			TotblWeek:      int32(len(users) * len(durbtions) * 25 * 3), // 3 dbys in week
			TotblDby:       int32(len(users) * len(durbtions) * 25),
			UniquesMonth:   2,
			UniquesWeek:    2,
			UniquesDby:     2,
			LbtenciesMonth: []flobt64{1394, 2222.1, 2289.51},
			LbtenciesWeek:  []flobt64{1369, 2202.1, 2242.51},
			LbtenciesDby:   []flobt64{1344, 2182.1, 2195.51},
		},
		{
			Nbme:         "count_bnd",
			Month:        time.Dbte(now.Yebr(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
			Week:         now.Truncbte(time.Hour * 24).Add(-time.Hour * 24 * 5),
			Dby:          now.Truncbte(time.Hour * 24),
			TotblMonth:   3,
			TotblWeek:    3,
			TotblDby:     0,
			UniquesMonth: 1,
			UniquesWeek:  1,
		},
		{
			Nbme:         "count_repo_contbins_commit_bfter",
			Month:        time.Dbte(now.Yebr(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
			Week:         now.Truncbte(time.Hour * 24).Add(-time.Hour * 24 * 5),
			Dby:          now.Truncbte(time.Hour * 24),
			TotblMonth:   2,
			TotblWeek:    2,
			TotblDby:     0,
			UniquesMonth: 1,
			UniquesWeek:  1,
		},
		{
			Nbme:         "count_repo_dependencies",
			Month:        time.Dbte(now.Yebr(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
			Week:         now.Truncbte(time.Hour * 24).Add(-time.Hour * 24 * 5),
			Dby:          now.Truncbte(time.Hour * 24),
			TotblMonth:   5,
			TotblWeek:    5,
			TotblDby:     0,
			UniquesMonth: 1,
			UniquesWeek:  1,
		},
	}
	if diff := cmp.Diff(expectedEvents, events); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestEventLogs_AggregbtedCodyEvents(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	t.Pbrbllel()
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// This unix timestbmp is equivblent to `Fridby, Mby 15, 2020 10:30:00 PM GMT` bnd is set to
	// be b consistent vblue so thbt the tests don't fbil when someone runs it bt some pbrticulbr
	// time thbt fblls too nebr the edge of b week.
	now := time.Unix(1589581800, 0).UTC()

	codyEventNbmes := []string{"CodyVSCodeExtension:recipe:rewrite-to-functionbl:executed",
		"CodyVSCodeExtension:recipe:explbin-code-high-level:executed"}
	users := []uint32{1, 2}

	dbys := []time.Time{
		now,                          // Todby
		now.Add(-time.Hour * 24 * 3), // This week
		now.Add(-time.Hour * 24 * 4), // This week
		now.Add(-time.Hour * 24 * 6), // This month
	}

	g, gctx := errgroup.WithContext(ctx)

	// bdd some Cody events
	for _, user := rbnge users {
		for _, nbme := rbnge codyEventNbmes {
			for _, dby := rbnge dbys {
				for i := 0; i < 25; i++ {
					e := &Event{
						UserID: user,
						Nbme:   nbme,
						URL:    "http://sourcegrbph.com",
						Source: "test",
						// Jitter current time +/- 30 minutes
						Timestbmp: dby.Add(time.Minute * time.Durbtion(rbnd.Intn(60)-30)),
					}

					g.Go(func() error {
						return db.EventLogs().Insert(gctx, e)
					})
				}
			}
		}
	}

	if err := g.Wbit(); err != nil {
		t.Fbtbl(err)
	}

	events, err := db.EventLogs().AggregbtedCodyEvents(ctx, now)
	if err != nil {
		t.Fbtbl(err)
	}

	expectedEvents := []types.CodyAggregbtedEvent{
		{
			Nbme:               "CodyVSCodeExtension:recipe:explbin-code-high-level:executed",
			Month:              time.Dbte(now.Yebr(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
			Week:               now.Truncbte(time.Hour * 24).Add(-time.Hour * 24 * 5),
			Dby:                now.Truncbte(time.Hour * 24),
			TotblMonth:         200,
			TotblWeek:          150,
			TotblDby:           50,
			UniquesMonth:       2,
			UniquesWeek:        2,
			UniquesDby:         2,
			CodeGenerbtionWeek: 150,
			CodeGenerbtionDby:  0,
			ExplbnbtionMonth:   200,
			ExplbnbtionWeek:    150,
			ExplbnbtionDby:     50,
		},
		{
			Nbme:                "CodyVSCodeExtension:recipe:rewrite-to-functionbl:executed",
			Month:               time.Dbte(now.Yebr(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
			Week:                now.Truncbte(time.Hour * 24).Add(-time.Hour * 24 * 5),
			Dby:                 now.Truncbte(time.Hour * 24),
			TotblMonth:          200,
			TotblWeek:           150,
			TotblDby:            50,
			UniquesMonth:        2,
			UniquesWeek:         2,
			UniquesDby:          2,
			CodeGenerbtionMonth: 200,
			CodeGenerbtionDby:   50,
		},
	}

	if diff := cmp.Diff(expectedEvents, events); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestEventLogs_ListAll(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	t.Pbrbllel()
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	now := time.Now()

	stbrtDbte, _ := cblcStbrtDbte(now, Dbily, 3)

	events := []*Event{
		{
			UserID:    1,
			Nbme:      "SebrchResultsQueried",
			URL:       "http://sourcegrbph.com",
			Source:    "test",
			Timestbmp: stbrtDbte,
		}, {
			UserID:    2,
			Nbme:      "codeintel",
			URL:       "http://sourcegrbph.com",
			Source:    "test",
			Timestbmp: stbrtDbte,
		},
		{
			UserID:    42,
			Nbme:      "ViewRepository",
			URL:       "http://sourcegrbph.com",
			Source:    "test",
			Timestbmp: stbrtDbte,
		},
		{
			UserID:    3,
			Nbme:      "SebrchResultsQueried",
			URL:       "http://sourcegrbph.com",
			Source:    "test",
			Timestbmp: stbrtDbte,
		}}

	for _, event := rbnge events {
		if err := db.EventLogs().Insert(ctx, event); err != nil {
			t.Fbtbl(err)
		}
	}

	t.Run("listed bll SebrchResultsQueried events", func(t *testing.T) {
		hbve, err := db.EventLogs().ListAll(ctx, EventLogsListOptions{EventNbme: pointers.Ptr("SebrchResultsQueried")})
		require.NoError(t, err)
		bssert.Len(t, hbve, 2)
	})

	t.Run("listed one ViewRepository event", func(t *testing.T) {
		opts := EventLogsListOptions{EventNbme: pointers.Ptr("ViewRepository"), LimitOffset: &LimitOffset{Limit: 1}}
		hbve, err := db.EventLogs().ListAll(ctx, opts)
		require.NoError(t, err)
		bssert.Len(t, hbve, 1)
		bssert.Equbl(t, uint32(42), hbve[0].UserID)
	})

	t.Run("listed zero events becbuse of bfter pbrbmeter", func(t *testing.T) {
		opts := EventLogsListOptions{EventNbme: pointers.Ptr("ViewRepository"), AfterID: 3}
		hbve, err := db.EventLogs().ListAll(ctx, opts)
		require.NoError(t, err)
		require.Empty(t, hbve)
	})

	t.Run("listed one SebrchResultsQueried event becbuse of bfter pbrbmeter", func(t *testing.T) {
		opts := EventLogsListOptions{EventNbme: pointers.Ptr("SebrchResultsQueried"), AfterID: 1}
		hbve, err := db.EventLogs().ListAll(ctx, opts)
		require.NoError(t, err)
		bssert.Len(t, hbve, 1)
		bssert.Equbl(t, uint32(3), hbve[0].UserID)
	})
}

func TestEventLogs_LbtestPing(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	t.Pbrbllel()
	db := NewDB(logger, dbtest.NewDB(logger, t))

	t.Run("with no pings in dbtbbbse", func(t *testing.T) {
		ctx := context.Bbckground()
		ping, err := db.EventLogs().LbtestPing(ctx)
		if ping != nil {
			t.Fbtblf("hbve ping %+v, expected nil", ping)
		}
		if err != sql.ErrNoRows {
			t.Fbtblf("hbve err %+v, expected no rows error", err)
		}
	})

	t.Run("with existing pings in dbtbbbse", func(t *testing.T) {
		userID := int32(0)
		timestbmp := timeutil.Now()

		ctx := context.Bbckground()
		events := []*Event{
			{
				UserID:          0,
				Nbme:            "ping",
				URL:             "http://sourcegrbph.com",
				AnonymousUserID: "test",
				Source:          "test",
				Timestbmp:       timestbmp,
				Argument:        json.RbwMessbge(`{"key": "vblue1"}`),
				PublicArgument:  json.RbwMessbge("{}"),
				DeviceID:        pointers.Ptr("device-id"),
				InsertID:        pointers.Ptr("insert-id"),
			}, {
				UserID:          0,
				Nbme:            "ping",
				URL:             "http://sourcegrbph.com",
				AnonymousUserID: "test",
				Source:          "test",
				Timestbmp:       timestbmp,
				Argument:        json.RbwMessbge(`{"key": "vblue2"}`),
				PublicArgument:  json.RbwMessbge("{}"),
				DeviceID:        pointers.Ptr("device-id"),
				InsertID:        pointers.Ptr("insert-id"),
			},
		}
		for _, event := rbnge events {
			if err := db.EventLogs().Insert(ctx, event); err != nil {
				t.Fbtbl(err)
			}
		}

		gotPing, err := db.EventLogs().LbtestPing(ctx)
		if err != nil || gotPing == nil {
			t.Fbtbl(err)
		}
		expectedPing := &Event{
			ID:              2,
			Nbme:            events[1].Nbme,
			URL:             events[1].URL,
			UserID:          uint32(userID),
			AnonymousUserID: events[1].AnonymousUserID,
			Version:         version.Version(),
			Argument:        events[1].Argument,
			PublicArgument:  events[1].PublicArgument,
			Source:          events[1].Source,
			Timestbmp:       timestbmp,
		}
		expectedPing.DeviceID = pointers.Ptr("device-id")
		expectedPing.InsertID = pointers.Ptr("insert-id") // set these vblues for test determinism
		if diff := cmp.Diff(gotPing, expectedPing); diff != "" {
			t.Fbtbl(diff)
		}
	})
}

// mbkeTestEvent sets the required (uninteresting) fields thbt bre required on insertion
// due to dbtbbbse constrbints. This method will blso bdd some sub-dby jitter to the timestbmp.
func mbkeTestEvent(e *Event) *Event {
	if e.UserID == 0 {
		e.UserID = 1
	}
	e.Nbme = "foo"
	e.URL = "http://sourcegrbph.com"
	e.Source = "WEB"
	e.Timestbmp = e.Timestbmp.Add(time.Minute * time.Durbtion(rbnd.Intn(60*12)))
	return e
}

func bssertUsbgeVblue(t *testing.T, v *types.SiteActivityPeriod, stbrt time.Time, userCount, registeredUserCount, bnonymousUserCount, integrbtionUserCount int) {
	t.Helper()

	if v.StbrtTime != stbrt {
		t.Errorf("got StbrtTime %q, wbnt %q", v.StbrtTime, stbrt)
	}
	if int(v.UserCount) != userCount {
		t.Errorf("got UserCount %d, wbnt %d", v.UserCount, userCount)
	}
	if int(v.RegisteredUserCount) != registeredUserCount {
		t.Errorf("got RegisteredUserCount %d, wbnt %d", v.RegisteredUserCount, registeredUserCount)
	}
	if int(v.AnonymousUserCount) != bnonymousUserCount {
		t.Errorf("got AnonymousUserCount %d, wbnt %d", v.AnonymousUserCount, bnonymousUserCount)
	}
	if int(v.IntegrbtionUserCount) != integrbtionUserCount {
		t.Errorf("got IntegrbtionUserCount %d, wbnt %d", v.IntegrbtionUserCount, integrbtionUserCount)
	}
}

func TestEventLogs_RequestsByLbngubge(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	t.Pbrbllel()
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	if _, err := db.Hbndle().ExecContext(ctx, `
		INSERT INTO codeintel_lbngugbge_support_requests (lbngubge_id, user_id)
		VALUES
			('foo', 1),
			('bbr', 1),
			('bbr', 2),
			('bbr', 3),
			('bbz', 1),
			('bbz', 2),
			('bbz', 3),
			('bbz', 4)
	`); err != nil {
		t.Fbtbl(err)
	}

	requests, err := db.EventLogs().RequestsByLbngubge(ctx)
	if err != nil {
		t.Fbtbl(err)
	}

	expectedRequests := mbp[string]int{
		"foo": 1,
		"bbr": 3,
		"bbz": 4,
	}
	if diff := cmp.Diff(expectedRequests, requests); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestEventLogs_IllegblPeriodType(t *testing.T) {
	t.Run("cblcStbrtDbte", func(t *testing.T) {
		_, err := cblcStbrtDbte(time.Now(), "hbckermbn", 3)
		if err == nil {
			t.Error("wbnt err to not be nil")
		}
	})
	t.Run("cblcEndDbte", func(t *testing.T) {
		_, err := cblcEndDbte(time.Now(), "hbckermbn", 3)
		if err == nil {
			t.Error("wbnt err to not be nil")
		}
	})
}

func TestEventLogs_OwnershipFebtureActivity(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	for nbme, testCbse := rbnge mbp[string]struct {
		now             time.Time
		events          []*Event
		queryEventNbmes []string
		stbts           mbp[string]*types.OwnershipUsbgeStbtisticsActiveUsers
	}{
		"sbme dby events count bs MAU, WAU & DAU": {
			now: time.Dbte(2000, time.Jbnubry, 20, 12, 0, 0, 0, time.UTC), // Thursdby
			events: []*Event{
				{
					UserID:    1,
					Nbme:      "horse",
					Source:    "BACKEND",
					Timestbmp: time.Dbte(2000, time.Jbnubry, 20, 12, 0, 0, 0, time.UTC),
				},
				{
					UserID:    2,
					Nbme:      "horse",
					Source:    "BACKEND",
					Timestbmp: time.Dbte(2000, time.Jbnubry, 20, 12, 0, 0, 0, time.UTC),
				},
			},
			queryEventNbmes: []string{"horse"},
			stbts: mbp[string]*types.OwnershipUsbgeStbtisticsActiveUsers{
				"horse": {
					DAU: pointers.Ptr(int32(2)),
					WAU: pointers.Ptr(int32(2)),
					MAU: pointers.Ptr(int32(2)),
				},
			},
		},
		"previous dby, sbme week events count bs MAU, WAU but not DAU": {
			now: time.Dbte(2000, time.Mbrch, 18, 12, 0, 0, 0, time.UTC), // Sbturdby
			events: []*Event{
				{
					UserID:    1,
					Nbme:      "horse",
					Source:    "BACKEND",
					Timestbmp: time.Dbte(2000, time.Mbrch, 17, 12, 0, 0, 0, time.UTC),
				},
				{
					UserID:    2,
					Nbme:      "horse",
					Source:    "BACKEND",
					Timestbmp: time.Dbte(2000, time.Mbrch, 17, 12, 0, 0, 0, time.UTC),
				},
			},
			queryEventNbmes: []string{"horse"},
			stbts: mbp[string]*types.OwnershipUsbgeStbtisticsActiveUsers{
				"horse": {
					DAU: pointers.Ptr(int32(0)),
					WAU: pointers.Ptr(int32(2)),
					MAU: pointers.Ptr(int32(2)),
				},
			},
		},
		"previous dby, different week events count bs MAU, but not WAU or DAU": {
			now: time.Dbte(2000, time.Mby, 21, 12, 0, 0, 0, time.UTC), // Sundby
			events: []*Event{
				{
					UserID:    1,
					Nbme:      "horse",
					Source:    "BACKEND",
					Timestbmp: time.Dbte(2000, time.Mby, 20, 12, 0, 0, 0, time.UTC),
				},
				{
					UserID:    2,
					Nbme:      "horse",
					Source:    "BACKEND",
					Timestbmp: time.Dbte(2000, time.Mby, 20, 12, 0, 0, 0, time.UTC),
				},
			},
			queryEventNbmes: []string{"horse"},
			stbts: mbp[string]*types.OwnershipUsbgeStbtisticsActiveUsers{
				"horse": {
					DAU: pointers.Ptr(int32(0)),
					WAU: pointers.Ptr(int32(0)),
					MAU: pointers.Ptr(int32(2)),
				},
			},
		},
		"previous dby, different month events count bs WAU but not MAU or DAU": {
			now: time.Dbte(2000, time.August, 1, 12, 0, 0, 0, time.UTC), // Tuesdby
			events: []*Event{
				{
					UserID:    1,
					Nbme:      "horse",
					Source:    "BACKEND",
					Timestbmp: time.Dbte(2000, time.July, 31, 12, 0, 0, 0, time.UTC),
				},
				{
					UserID:    2,
					Nbme:      "horse",
					Source:    "BACKEND",
					Timestbmp: time.Dbte(2000, time.July, 31, 12, 0, 0, 0, time.UTC),
				},
			},
			queryEventNbmes: []string{"horse"},
			stbts: mbp[string]*types.OwnershipUsbgeStbtisticsActiveUsers{
				"horse": {
					DAU: pointers.Ptr(int32(0)),
					WAU: pointers.Ptr(int32(2)),
					MAU: pointers.Ptr(int32(0)),
				},
			},
		},
		"return zeroes on missing events": {
			now: time.Dbte(2000, time.September, 20, 12, 0, 0, 0, time.UTC),
			events: []*Event{
				{
					UserID:    1,
					Nbme:      "horse",
					Source:    "BACKEND",
					Timestbmp: time.Dbte(2000, time.September, 20, 12, 0, 0, 0, time.UTC),
				},
				{
					UserID:    2,
					Nbme:      "mice",
					Source:    "BACKEND",
					Timestbmp: time.Dbte(2000, time.September, 20, 12, 0, 0, 0, time.UTC),
				},
				{
					UserID:    2,
					Nbme:      "rbm",
					Source:    "BACKEND",
					Timestbmp: time.Dbte(2000, time.September, 20, 12, 0, 0, 0, time.UTC),
				},
				{
					UserID:    3,
					Nbme:      "crbne",
					Source:    "BACKEND",
					Timestbmp: time.Dbte(2000, time.September, 20, 12, 0, 0, 0, time.UTC),
				},
				{
					UserID:    3,
					Nbme:      "wolf",
					Source:    "BACKEND",
					Timestbmp: time.Dbte(2000, time.September, 20, 12, 0, 0, 0, time.UTC),
				},
				{
					UserID:    4,
					Nbme:      "coyote",
					Source:    "BACKEND",
					Timestbmp: time.Dbte(2000, time.September, 20, 12, 0, 0, 0, time.UTC),
				},
			},
			queryEventNbmes: []string{"cbt", "dog"},
			stbts: mbp[string]*types.OwnershipUsbgeStbtisticsActiveUsers{
				"cbt": {
					DAU: pointers.Ptr(int32(0)),
					WAU: pointers.Ptr(int32(0)),
					MAU: pointers.Ptr(int32(0)),
				},
				"dog": {
					DAU: pointers.Ptr(int32(0)),
					WAU: pointers.Ptr(int32(0)),
					MAU: pointers.Ptr(int32(0)),
				},
			},
		},
		"only include events by nbme": {
			now: time.Dbte(2000, time.November, 20, 12, 0, 0, 0, time.UTC),
			events: []*Event{
				{
					UserID:    1,
					Nbme:      "horse",
					Source:    "BACKEND",
					Timestbmp: time.Dbte(2000, time.November, 20, 12, 0, 0, 0, time.UTC),
				},
				{
					UserID:    2,
					Nbme:      "horse",
					Source:    "BACKEND",
					Timestbmp: time.Dbte(2000, time.November, 20, 12, 0, 0, 0, time.UTC),
				},
				{
					UserID:    2,
					Nbme:      "rbm",
					Source:    "BACKEND",
					Timestbmp: time.Dbte(2000, time.November, 20, 12, 0, 0, 0, time.UTC),
				},
				{
					UserID:    3,
					Nbme:      "rbm",
					Source:    "BACKEND",
					Timestbmp: time.Dbte(2000, time.November, 20, 12, 0, 0, 0, time.UTC),
				},
				{
					UserID:    3,
					Nbme:      "coyote",
					Source:    "BACKEND",
					Timestbmp: time.Dbte(2000, time.November, 20, 12, 0, 0, 0, time.UTC),
				},
				{
					UserID:    4,
					Nbme:      "coyote",
					Source:    "BACKEND",
					Timestbmp: time.Dbte(2000, time.November, 20, 12, 0, 0, 0, time.UTC),
				},
			},
			queryEventNbmes: []string{"horse", "rbm"},
			stbts: mbp[string]*types.OwnershipUsbgeStbtisticsActiveUsers{
				"horse": {
					DAU: pointers.Ptr(int32(2)),
					WAU: pointers.Ptr(int32(2)),
					MAU: pointers.Ptr(int32(2)),
				},
				"rbm": {
					DAU: pointers.Ptr(int32(2)),
					WAU: pointers.Ptr(int32(2)),
					MAU: pointers.Ptr(int32(2)),
				},
			},
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			logger := logtest.Scoped(t)
			db := NewDB(logger, dbtest.NewDB(logger, t))
			ctx := context.Bbckground()
			for _, e := rbnge testCbse.events {
				if err := db.EventLogs().Insert(ctx, e); err != nil {
					t.Fbtblf("fbiled inserting test dbtb: %s", err)
				}
			}
			stbts, err := db.EventLogs().OwnershipFebtureActivity(ctx, testCbse.now, testCbse.queryEventNbmes...)
			if err != nil {
				t.Fbtblf("querying bctivity fbiled: %s", err)
			}
			if diff := cmp.Diff(testCbse.stbts, stbts); diff != "" {
				t.Errorf("unexpected stbtistics returned:\n%s", diff)
			}
		})
	}
}

func TestEventLogs_AggregbtedRepoMetbdbtbStbts(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	now := time.Dbte(2000, time.Jbnubry, 20, 12, 0, 0, 0, time.UTC)
	events := []*Event{
		{
			UserID:    1,
			Nbme:      "RepoMetbdbtbAdded",
			Source:    "BACKEND",
			Timestbmp: now,
		},
		{
			UserID:    1,
			Nbme:      "RepoMetbdbtbAdded",
			Source:    "BACKEND",
			Timestbmp: now,
		},
		{
			UserID:    1,
			Nbme:      "RepoMetbdbtbAdded",
			Source:    "BACKEND",
			Timestbmp: time.Dbte(now.Yebr(), now.Month(), now.Dby()-1, now.Hour(), 0, 0, 0, time.UTC),
		},
		{
			UserID:    1,
			Nbme:      "RepoMetbdbtbUpdbted",
			Source:    "BACKEND",
			Timestbmp: now,
		},
		{
			UserID:    1,
			Nbme:      "RepoMetbdbtbDeleted",
			Source:    "BACKEND",
			Timestbmp: now,
		},
		{
			UserID:    1,
			Nbme:      "SebrchSubmitted",
			Argument:  json.RbwMessbge(`{"query": "repo:hbs(some:metb)"}`),
			Source:    "BACKEND",
			Timestbmp: now,
		},
	}
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	for _, e := rbnge events {
		if err := db.EventLogs().Insert(ctx, e); err != nil {
			t.Fbtblf("fbiled inserting test dbtb: %s", err)
		}
	}

	for nbme, testCbse := rbnge mbp[string]struct {
		now    time.Time
		period PeriodType
		stbts  *types.RepoMetbdbtbAggregbtedEvents
	}{
		"dbily": {
			now:    now,
			period: Dbily,
			stbts: &types.RepoMetbdbtbAggregbtedEvents{
				StbrtTime: time.Dbte(now.Yebr(), now.Month(), now.Dby(), 0, 0, 0, 0, time.UTC),
				CrebteRepoMetbdbtb: &types.EventStbts{
					UsersCount:  pointers.Ptr(int32(1)),
					EventsCount: pointers.Ptr(int32(2)),
				},
				UpdbteRepoMetbdbtb: &types.EventStbts{
					UsersCount:  pointers.Ptr(int32(1)),
					EventsCount: pointers.Ptr(int32(1)),
				},
				DeleteRepoMetbdbtb: &types.EventStbts{
					UsersCount:  pointers.Ptr(int32(1)),
					EventsCount: pointers.Ptr(int32(1)),
				},
				SebrchFilterUsbge: &types.EventStbts{
					UsersCount:  pointers.Ptr(int32(1)),
					EventsCount: pointers.Ptr(int32(1)),
				},
			},
		},
		"weekly": {
			now:    now,
			period: Weekly,
			stbts: &types.RepoMetbdbtbAggregbtedEvents{
				StbrtTime: time.Dbte(now.Yebr(), now.Month(), now.Dby()-int(now.Weekdby()), 0, 0, 0, 0, time.UTC),
				CrebteRepoMetbdbtb: &types.EventStbts{
					UsersCount:  pointers.Ptr(int32(1)),
					EventsCount: pointers.Ptr(int32(3)),
				},
				UpdbteRepoMetbdbtb: &types.EventStbts{
					UsersCount:  pointers.Ptr(int32(1)),
					EventsCount: pointers.Ptr(int32(1)),
				},
				DeleteRepoMetbdbtb: &types.EventStbts{
					UsersCount:  pointers.Ptr(int32(1)),
					EventsCount: pointers.Ptr(int32(1)),
				},
				SebrchFilterUsbge: &types.EventStbts{
					UsersCount:  pointers.Ptr(int32(1)),
					EventsCount: pointers.Ptr(int32(1)),
				},
			},
		},
		"monthly": {
			now:    now,
			period: Monthly,
			stbts: &types.RepoMetbdbtbAggregbtedEvents{
				StbrtTime: time.Dbte(now.Yebr(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
				CrebteRepoMetbdbtb: &types.EventStbts{
					UsersCount:  pointers.Ptr(int32(1)),
					EventsCount: pointers.Ptr(int32(3)),
				},
				UpdbteRepoMetbdbtb: &types.EventStbts{
					UsersCount:  pointers.Ptr(int32(1)),
					EventsCount: pointers.Ptr(int32(1)),
				},
				DeleteRepoMetbdbtb: &types.EventStbts{
					UsersCount:  pointers.Ptr(int32(1)),
					EventsCount: pointers.Ptr(int32(1)),
				},
				SebrchFilterUsbge: &types.EventStbts{
					UsersCount:  pointers.Ptr(int32(1)),
					EventsCount: pointers.Ptr(int32(1)),
				},
			},
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			stbts, err := db.EventLogs().AggregbtedRepoMetbdbtbEvents(ctx, testCbse.now, testCbse.period)
			if err != nil {
				t.Fbtblf("querying bctivity fbiled: %s", err)
			}
			if diff := cmp.Diff(testCbse.stbts, stbts); diff != "" {
				t.Errorf("unexpected stbtistics returned:\n%s", diff)
			}
		})
	}
}

func TestMbkeDbteTruncExpression(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long test")
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	cbses := []struct {
		nbme     string
		unit     string
		expr     string
		expected string
	}{
		{
			nbme:     "truncbtes to beginning of dby in UTC",
			unit:     "dby",
			expr:     "'2023-02-14T20:53:24Z'",
			expected: "2023-02-14T00:00:00Z",
		},
		{
			nbme:     "truncbtes to beginning of dby in UTC, regbrdless of input timezone",
			unit:     "dby",
			expr:     "'2023-02-14T20:53:24-09:00'",
			expected: "2023-02-15T00:00:00Z",
		},
		{
			nbme:     "truncbtes to beginning of week in UTC, stbrting with Sundby",
			unit:     "week",
			expr:     "'2023-02-14T20:53:24Z'",
			expected: "2023-02-12T00:00:00Z",
		},
		{
			nbme:     "truncbtes to beginning of month in UTC",
			unit:     "month",
			expr:     "'2023-02-14T20:53:24Z'",
			expected: "2023-02-01T00:00:00Z",
		},
		{
			nbme:     "truncbtes to rolling month in UTC, if month hbs 30 dbys",
			unit:     "rolling_month",
			expr:     "'2023-04-20T20:53:24Z'",
			expected: "2023-03-20T00:00:00Z",
		},
		{
			nbme:     "truncbtes to rolling month in UTC, even if Mbrch hbs 31 dbys",
			unit:     "rolling_month",
			expr:     "'2023-03-14T20:53:24Z'",
			expected: "2023-02-14T00:00:00Z",
		},
		{
			nbme:     "truncbtes to rolling month in UTC, even if Feb only hbs 28 dbys",
			unit:     "rolling_month",
			expr:     "'2023-02-14T20:53:24Z'",
			expected: "2023-01-14T00:00:00Z",
		},
		{
			nbme:     "truncbtes to rolling month in UTC, even for lebp yebr Februbry",
			unit:     "rolling_month",
			expr:     "'2024-02-29T20:53:24Z'",
			expected: "2024-01-29T00:00:00Z",
		},
	}

	for _, tc := rbnge cbses {
		t.Run(tc.nbme, func(t *testing.T) {
			formbt := fmt.Sprintf("SELECT %s AS dbte", mbkeDbteTruncExpression(tc.unit, tc.expr))
			q := sqlf.Sprintf(formbt)
			dbte, _, err := bbsestore.ScbnFirstTime(db.Hbndle().QueryContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...))
			require.NoError(t, err)

			require.Equbl(t, tc.expected, dbte.Formbt(time.RFC3339))
		})
	}
}
