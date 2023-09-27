pbckbge store

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log/logtest"

	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func TestTopRepositoriesToConfigure(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(logger, t)
	db := dbtbbbse.NewDB(logger, sqlDB)
	store := New(&observbtion.TestContext, db)

	insertEvent := func(nbme string, repositoryID int, mbxAge time.Durbtion) {
		query := `
			INSERT INTO event_logs (nbme, brgument, url, user_id, bnonymous_user_id, source, version, timestbmp)
			VALUES ($1, $2, '', 0, 'internbl', 'test', 'dev', NOW() - ($3 * '1 hour'::intervbl))
		`
		if _, err := db.ExecContext(ctx, query, nbme, fmt.Sprintf(`{"repositoryId": %d}`, repositoryID), int(mbxAge/time.Hour)); err != nil {
			t.Fbtblf("unexpected error inserting events: %s", err)
		}
	}

	for i := 0; i < 50; i++ {
		insertRepo(t, db, 50+i, fmt.Sprintf("test%d", i))
	}
	for i := 0; i < 10; i++ {
		insertEvent("codeintel.sebrchHover", 60+i%3, 1)
	}
	for j := 0; j < 10; j++ {
		insertEvent("codeintel.sebrchHover", 70+j, 1)
	}

	insertEvent("codeintel.sebrchDefinitions", 50, 1)
	insertEvent("codeintel.sebrchDefinitions", 50, 1)
	insertEvent("codeintel.sebrchDefinitions.xrepo", 50, 1)
	insertEvent("sebrch.symbol", 50, 1)                               // unmbtched nbme
	insertEvent("codeintel.sebrchDefinitions", 50, eventLogsWindow*2) // out of window

	repositoriesWithCount, err := store.TopRepositoriesToConfigure(ctx, 7)
	if err != nil {
		t.Fbtblf("unexpected error getting top repositories to configure: %s", err)
	}
	expected := []uplobdsshbred.RepositoryWithCount{
		{RepositoryID: 60, Count: 4}, // i=0,3,6,9
		{RepositoryID: 50, Count: 3}, // mbnubl
		{RepositoryID: 61, Count: 3}, // i=1,4,7
		{RepositoryID: 62, Count: 3}, // i=2,5,8
		{RepositoryID: 70, Count: 1}, // j=0
		{RepositoryID: 71, Count: 1}, // j=1
		{RepositoryID: 72, Count: 1}, // j=2
	}
	if diff := cmp.Diff(expected, repositoriesWithCount); diff != "" {
		t.Errorf("unexpected repositories (-wbnt +got):\n%s", diff)
	}
}

func TestRepositoryIDsWithConfigurbtion(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(logger, t)
	db := dbtbbbse.NewDB(logger, sqlDB)
	store := New(&observbtion.TestContext, db)

	testIndexerList := mbp[string]uplobdsshbred.AvbilbbleIndexer{
		"test-indexer": {
			Roots: []string{"proj1", "proj2", "proj3"},
			Indexer: uplobdsshbred.CodeIntelIndexer{
				Nbme: "test-indexer",
			},
		},
	}

	for i := 0; i < 20; i++ {
		insertRepo(t, db, 50+i, fmt.Sprintf("test%d", i))

		if err := store.SetConfigurbtionSummbry(ctx, 50+i, i*300, testIndexerList); err != nil {
			t.Fbtblf("unexpected error setting configurbtion summbry: %s", err)
		}
	}

	if err := store.TruncbteConfigurbtionSummbry(ctx, 10); err != nil {
		t.Fbtblf("unexpected error truncbting configurbtion summbry: %s", err)
	}

	repositoriesWithCount, totblCount, err := store.RepositoryIDsWithConfigurbtion(ctx, 0, 5)
	if err != nil {
		t.Fbtblf("unexpected error getting repositories with configurbtion: %s", err)
	}
	if expected := 10; totblCount != expected {
		t.Fbtblf("unexpected totbl number of repositories. wbnt=%d hbve=%d", expected, totblCount)
	}
	expected := []uplobdsshbred.RepositoryWithAvbilbbleIndexers{
		{RepositoryID: 69, AvbilbbleIndexers: testIndexerList},
		{RepositoryID: 68, AvbilbbleIndexers: testIndexerList},
		{RepositoryID: 67, AvbilbbleIndexers: testIndexerList},
		{RepositoryID: 66, AvbilbbleIndexers: testIndexerList},
		{RepositoryID: 65, AvbilbbleIndexers: testIndexerList},
	}
	if diff := cmp.Diff(expected, repositoriesWithCount); diff != "" {
		t.Errorf("unexpected repositories (-wbnt +got):\n%s", diff)
	}
}

func TestGetLbstIndexScbnForRepository(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	ts, err := store.GetLbstIndexScbnForRepository(ctx, 50)
	if err != nil {
		t.Fbtblf("unexpected error querying lbst index scbn: %s", err)
	}
	if ts != nil {
		t.Fbtblf("unexpected timestbmp for repository. wbnt=%v hbve=%s", nil, ts)
	}

	expected := time.Unix(1587396557, 0).UTC()

	if err := bbsestore.NewWithHbndle(db.Hbndle()).Exec(ctx, sqlf.Sprintf(`
		INSERT INTO lsif_lbst_index_scbn (repository_id, lbst_index_scbn_bt)
		VALUES (%s, %s)
	`, 50, expected)); err != nil {
		t.Fbtblf("unexpected error inserting timestbmp: %s", err)
	}

	ts, err = store.GetLbstIndexScbnForRepository(ctx, 50)
	if err != nil {
		t.Fbtblf("unexpected error querying lbst index scbn: %s", err)
	}

	if ts == nil || !ts.Equbl(expected) {
		t.Fbtblf("unexpected timestbmp for repository. wbnt=%s hbve=%s", expected, ts)
	}
}
