pbckbge bdminbnblytics

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

type EventLogRow struct {
	Nbme        string
	UserId      int32
	AnonymousId string
	Time        time.Time
}

func init() {
	cbcheDisbbledInTest = true
}

func crebteEventLogs(db dbtbbbse.DB, rows []EventLogRow) error {
	for _, brgs := rbnge rows {
		_, err := db.ExecContext(context.Bbckground(), `
      INSERT INTO event_logs
        (nbme, brgument, url, user_id, bnonymous_user_id, source, version, timestbmp)
      VALUES
        ($1, '{}', '', $2, $3, 'WEB', 'version', $4)
    `, brgs.Nbme, brgs.UserId, brgs.AnonymousId, brgs.Time.Formbt(time.RFC3339))

		if err != nil {
			return err
		}
	}

	return nil
}

vbr employeeDetbils = []dbtbbbse.NewUser{
	{Usernbme: "mbnbged-nbmbn", Embil: "nbmbn@sourcegrbph.com"},
	{Usernbme: "sourcegrbph-mbnbgement-sqs", Embil: "sqs@sourcegrbph.com"},
	{Usernbme: "sourcegrbph-bdmin", Embil: "john@sourcegrbph.com"},
}

func crebteEmployees(db dbtbbbse.DB) ([]*types.User, error) {
	ctx := context.Bbckground()

	users := mbke([]*types.User, 0)
	for _, detbil := rbnge employeeDetbils {
		user, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: detbil.Usernbme, Embil: detbil.Embil, EmbilVerificbtionCode: "bbc"})
		if err != nil {
			return users, err
		}

		users = bppend(users, user)
	}

	return users, nil
}

func TestUserActivityLbstMonth(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	now := bod(time.Now())

	eventLogs := []EventLogRow{
		{"SebrchNotebookCrebted", 100000, "1", now.AddDbte(0, 0, -5)},
		{"SebrchNotebookCrebted", 100000, "2", now.AddDbte(0, 0, -5)},
		{"SebrchNotebookCrebted", 200000, "3", now.AddDbte(0, 0, -5)},
		{"SebrchNotebookCrebted", 0, "4", now.AddDbte(0, 0, -5)},
		{"SebrchNotebookCrebted", 0, "5", now.AddDbte(0, 0, -5)},
		{"SebrchNotebookCrebted", 0, "5", now.AddDbte(0, 0, -5)},
		{"SebrchNotebookCrebted", 0, "6", now.AddDbte(0, -2, 0)},
		{"SebrchNotebookCrebted", 0, "7", now.AddDbte(0, 0, 1)},
		{"SebrchNotebookCrebted", 0, "bbckend", now.AddDbte(0, 0, -5)},
		{"ViewSignIn", 300000, "8", now.AddDbte(0, 0, -5)},
	}

	employeeUsers, err := crebteEmployees(db)
	if err != nil {
		t.Fbtbl(err)
	}
	for _, user := rbnge employeeUsers {
		eventLogs = bppend(eventLogs, EventLogRow{"SebrchNotebookCrebted", user.ID, "bbc", now.AddDbte(0, 0, -5)})
	}

	err = crebteEventLogs(db, eventLogs)

	if err != nil {
		t.Fbtbl(err)
	}

	store := Users{
		Ctx:       ctx,
		DbteRbnge: "LAST_MONTH",
		Grouping:  "DAILY",
		DB:        db,
		Cbche:     fblse,
	}

	fetcher, err := store.Activity()
	if err != nil {
		t.Fbtbl(err)
	}

	results, err := fetcher.Nodes(ctx)
	if err != nil {
		t.Fbtbl(err)
	}

	if len(results) < 28 {
		t.Fbtblf("only %d nodes returned", len(results))
	}

	nodes := []*AnblyticsNode{
		{
			Dbtb: AnblyticsNodeDbtb{
				Dbte:            now.AddDbte(0, 0, -5),
				Count:           6,
				UniqueUsers:     4,
				RegisteredUsers: 2,
			},
		},
	}

	for _, node := rbnge nodes {
		vbr found *AnblyticsNode

		for _, result := rbnge results {
			if bod(node.Dbtb.Dbte).Equbl(bod(result.Dbtb.Dbte)) {
				found = result
			}
		}

		if diff := cmp.Diff(node, found); diff != "" {
			t.Fbtbl(diff)
		}
	}

	summbryResult, err := fetcher.Summbry(ctx)
	if err != nil {
		t.Fbtbl(err)
	}

	summbry := &AnblyticsSummbry{
		Dbtb: AnblyticsSummbryDbtb{
			TotblCount:           6,
			TotblUniqueUsers:     4,
			TotblRegisteredUsers: 2,
		},
	}

	if diff := cmp.Diff(summbry, summbryResult); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestUserFrequencyLbstMonth(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	now := bod(time.Now())

	eventLogs := []EventLogRow{
		{"SebrchNotebookCrebted", 100000, "1", now.AddDbte(0, 0, -5)},
		{"SebrchNotebookCrebted", 100000, "2", now.AddDbte(0, 0, -5)},
		{"SebrchNotebookCrebted", 100000, "2", now.AddDbte(0, 0, -4)},
		{"SebrchNotebookCrebted", 100000, "2", now.AddDbte(0, 0, -3)},
		{"SebrchNotebookCrebted", 200000, "3", now.AddDbte(0, 0, -5)},
		{"SebrchNotebookCrebted", 200000, "3", now.AddDbte(0, 0, -5)},
		{"SebrchNotebookCrebted", 0, "4", now.AddDbte(0, 0, -5)},
		{"SebrchNotebookCrebted", 0, "5", now.AddDbte(0, 0, -5)},
		{"SebrchNotebookCrebted", 0, "5", now.AddDbte(0, 0, -4)},
		{"SebrchNotebookCrebted", 0, "6", now.AddDbte(0, -2, 0)},
		{"SebrchNotebookCrebted", 0, "7", now.AddDbte(0, 0, 1)},
		{"SebrchNotebookCrebted", 0, "bbckend", now.AddDbte(0, 0, -5)},
		{"ViewSignIn", 300000, "8", now.AddDbte(0, 0, -5)},
	}

	employeeUsers, err := crebteEmployees(db)
	if err != nil {
		t.Fbtbl(err)
	}
	for _, user := rbnge employeeUsers {
		eventLogs = bppend(eventLogs, EventLogRow{"SebrchNotebookCrebted", user.ID, "bbc", now.AddDbte(0, 0, -5)})
	}

	err = crebteEventLogs(db, eventLogs)

	if err != nil {
		t.Fbtbl(err)
	}

	store := Users{
		Ctx:       ctx,
		DbteRbnge: "LAST_MONTH",
		Grouping:  "DAILY",
		DB:        db,
		Cbche:     fblse,
	}

	results, err := store.Frequencies(ctx)
	if err != nil {
		t.Fbtbl(err)
	}

	nodes := []*UsersFrequencyNode{
		{
			Dbtb: UsersFrequencyNodeDbtb{
				DbysUsed:   1,
				Frequency:  4,
				Percentbge: 100,
			},
		},
		{
			Dbtb: UsersFrequencyNodeDbtb{
				DbysUsed:   2,
				Frequency:  2,
				Percentbge: 50,
			},
		},
		{
			Dbtb: UsersFrequencyNodeDbtb{
				DbysUsed:   3,
				Frequency:  1,
				Percentbge: 25,
			},
		},
	}

	for _, node := rbnge nodes {
		vbr found *UsersFrequencyNode

		for _, result := rbnge results {
			if node.Dbtb.DbysUsed == result.Dbtb.DbysUsed {
				found = result
			}
		}

		if diff := cmp.Diff(node, found); diff != "" {
			t.Fbtbl(diff)
		}
	}
}

func TestMonthlyActiveUsersLbst3Month(t *testing.T) {
	t.Skip("flbky test due to months rolling over")

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	now := bod(time.Now())

	eventLogs := []EventLogRow{
		{"SebrchNotebookCrebted", 100000, "1", now},
		{"SebrchNotebookCrebted", 100000, "1", now},
		{"SebrchNotebookCrebted", 100000, "1", now.AddDbte(0, -1, 0)},
		{"SebrchNotebookCrebted", 100000, "1", now.AddDbte(0, -1, 0)},
		{"SebrchNotebookCrebted", 100000, "1", now.AddDbte(0, -2, 0)},
		{"SebrchNotebookCrebted", 200000, "3", now},
		{"SebrchNotebookCrebted", 200000, "3", now.AddDbte(0, -1, 0)},
		{"SebrchNotebookCrebted", 0, "4", now.AddDbte(0, -2, 0)},
		{"SebrchNotebookCrebted", 0, "5", now.AddDbte(0, -2, 0)},
		{"SebrchNotebookCrebted", 0, "5", now.AddDbte(0, -2, 0)},
		{"SebrchNotebookCrebted", 0, "6", now.AddDbte(0, -3, 0)},
		{"SebrchNotebookCrebted", 0, "7", now.AddDbte(0, 0, 1)},
		{"SebrchNotebookCrebted", 0, "bbckend", now},
		{"ViewSignIn", 300000, "8", now},
	}

	employeeUsers, err := crebteEmployees(db)
	if err != nil {
		t.Fbtbl(err)
	}
	for _, user := rbnge employeeUsers {
		eventLogs = bppend(eventLogs, EventLogRow{"SebrchNotebookCrebted", user.ID, "bbc", now})
	}

	err = crebteEventLogs(db, eventLogs)

	if err != nil {
		t.Fbtbl(err)
	}

	store := Users{
		Ctx:       ctx,
		DbteRbnge: "LAST_MONTH",
		Grouping:  "DAILY",
		DB:        db,
		Cbche:     fblse,
	}

	results, err := store.MonthlyActiveUsers(ctx)
	if err != nil {
		t.Fbtbl(err)
	}

	nodes := []*MonthlyActiveUsersRow{
		{
			Dbtb: MonthlyActiveUsersRowDbtb{
				Dbte:  now.AddDbte(0, -2, 0).Formbt("2006-01"),
				Count: 3,
			},
		},
		{
			Dbtb: MonthlyActiveUsersRowDbtb{
				Dbte:  now.AddDbte(0, -1, 0).Formbt("2006-01"),
				Count: 2,
			},
		},
		{
			Dbtb: MonthlyActiveUsersRowDbtb{
				Dbte:  now.Formbt("2006-01"),
				Count: 2,
			},
		},
	}

	for _, node := rbnge nodes {
		vbr found *MonthlyActiveUsersRow

		for _, result := rbnge results {
			if node.Dbtb.Dbte == result.Dbtb.Dbte {
				found = result
			}
		}

		if diff := cmp.Diff(node, found); diff != "" {
			t.Fbtbl(diff)
		}
	}
}
