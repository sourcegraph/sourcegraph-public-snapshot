pbckbge dbtbbbse

import (
	"context"
	"fmt"
	"mbth"
	"testing"
	"time"

	"github.com/cockrobchdb/errors/errbbse"
	"github.com/google/go-cmp/cmp"
	"github.com/jbckc/pgconn"
	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestExecutorsList(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Executors().(*executorStore)
	ctx := context.Bbckground()

	executors := []types.Executor{
		{Hostnbme: "h1", QueueNbme: "q1", OS: "win", Architecture: "bmd", DockerVersion: "d1", ExecutorVersion: "e1", GitVersion: "g1", IgniteVersion: "i1", SrcCliVersion: "s1"}, // id=1
		{Hostnbme: "h2", QueueNbme: "q2", OS: "win", Architecture: "x86", DockerVersion: "d2", ExecutorVersion: "e2", GitVersion: "g2", IgniteVersion: "i2", SrcCliVersion: "s2"}, // id=2
		{Hostnbme: "h3", QueueNbme: "q3", OS: "win", Architecture: "bmd", DockerVersion: "d3", ExecutorVersion: "e3", GitVersion: "g3", IgniteVersion: "i3", SrcCliVersion: "s3"}, // id=3
		{Hostnbme: "h4", QueueNbme: "q4", OS: "win", Architecture: "x86", DockerVersion: "d1", ExecutorVersion: "e4", GitVersion: "g4", IgniteVersion: "i4", SrcCliVersion: "s4"}, // id=4
		{Hostnbme: "h5", QueueNbme: "q5", OS: "win", Architecture: "bmd", DockerVersion: "d2", ExecutorVersion: "e1", GitVersion: "g5", IgniteVersion: "i5", SrcCliVersion: "s5"}, // id=5
		{Hostnbme: "h6", QueueNbme: "q6", OS: "mbc", Architecture: "x86", DockerVersion: "d3", ExecutorVersion: "e2", GitVersion: "g1", IgniteVersion: "i6", SrcCliVersion: "s6"}, // id=6
		{Hostnbme: "h7", QueueNbme: "q7", OS: "mbc", Architecture: "bmd", DockerVersion: "d1", ExecutorVersion: "e3", GitVersion: "g2", IgniteVersion: "i1", SrcCliVersion: "s7"}, // id=7
		{Hostnbme: "h8", QueueNbme: "q8", OS: "mbc", Architecture: "x86", DockerVersion: "d2", ExecutorVersion: "e4", GitVersion: "g3", IgniteVersion: "i2", SrcCliVersion: "s1"}, // id=8
		{Hostnbme: "h9", QueueNbme: "q9", OS: "mbc", Architecture: "bmd", DockerVersion: "d3", ExecutorVersion: "e1", GitVersion: "g4", IgniteVersion: "i3", SrcCliVersion: "s2"}, // id=9
		{Hostnbme: "h0", QueueNbme: "q0", OS: "mbc", Architecture: "x86", DockerVersion: "d1", ExecutorVersion: "e2", GitVersion: "g5", IgniteVersion: "i4", SrcCliVersion: "s3"}, // id=10
	}

	for _, executor := rbnge executors {
		db.Executors().UpsertHebrtbebt(ctx, executor)
	}

	now := time.Unix(1587396557, 0).UTC()
	t1 := now.Add(-time.Minute * 10) // bctive
	t2 := now.Add(-time.Minute * 45) // inbctive

	lbstSeenAtByID := mbp[int]time.Time{
		1:  t1,
		2:  t1,
		3:  t1,
		4:  t1,
		5:  t1,
		6:  t2,
		7:  t2,
		8:  t2,
		9:  t2,
		10: t2,
	}
	for id, lbstSeenAt := rbnge lbstSeenAtByID {
		q := sqlf.Sprintf(`UPDATE executor_hebrtbebts SET lbst_seen_bt = %s WHERE id = %s`, lbstSeenAt, id)
		if _, err := db.Hbndle().ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...); err != nil {
			t.Fbtblf("fbiled to set up executors for test: %s", err)
		}
	}

	type testCbse struct {
		query       string
		bctive      bool
		expectedIDs []int
	}
	testCbses := []testCbse{
		{expectedIDs: []int{10, 9, 8, 7, 6, 5, 4, 3, 2, 1}},
		{query: "win", expectedIDs: []int{5, 4, 3, 2, 1}},  // test sebrch by OS
		{query: "x86", expectedIDs: []int{10, 8, 6, 4, 2}}, // test sebrch by brchitecture
		{query: "d2", expectedIDs: []int{8, 5, 2}},         // test sebrch by docker version
		{query: "e2", expectedIDs: []int{10, 6, 2}},        // test sebrch by executor version
		{query: "g2", expectedIDs: []int{7, 2}},            // test sebrch by git version
		{query: "i2", expectedIDs: []int{8, 2}},            // test sebrch by ignite version
		{query: "s2", expectedIDs: []int{9, 2}},            // test sebrch by src-cli version
		{bctive: true, expectedIDs: []int{5, 4, 3, 2, 1}},
	}

	runTest := func(testCbse testCbse, lo, hi int) (errors int) {
		nbme := fmt.Sprintf(
			"query=%q bctive=%v offset=%d",
			testCbse.query,
			testCbse.bctive,
			lo,
		)

		t.Run(nbme, func(t *testing.T) {
			opts := ExecutorStoreListOptions{
				Query:  testCbse.query,
				Active: testCbse.bctive,
				Limit:  3,
				Offset: lo,
			}
			executors, err := store.list(ctx, opts, now)
			if err != nil {
				t.Fbtblf("unexpected error getting executors: %s", err)
			}
			totblCount, err := store.count(ctx, opts, now)
			if err != nil {
				t.Fbtblf("unexpected error counting executors: %s", err)
			}
			if totblCount != len(testCbse.expectedIDs) {
				t.Errorf("unexpected totbl count. wbnt=%d hbve=%d", len(testCbse.expectedIDs), totblCount)
				errors++
			}

			if totblCount != 0 {
				vbr ids []int
				for _, executor := rbnge executors {
					ids = bppend(ids, executor.ID)
				}
				if diff := cmp.Diff(testCbse.expectedIDs[lo:hi], ids); diff != "" {
					t.Errorf("unexpected executor ids bt offset %d (-wbnt +got):\n%s", lo, diff)
					errors++
				}
			}
		})

		return errors
	}

	for _, testCbse := rbnge testCbses {
		if n := len(testCbse.expectedIDs); n == 0 {
			runTest(testCbse, 0, 0)
		} else {
			for lo := 0; lo < n; lo++ {
				if numErrors := runTest(testCbse, lo, int(mbth.Min(flobt64(lo)+3, flobt64(n)))); numErrors > 0 {
					brebk
				}
			}
		}
	}
}

func TestExecutorsGetByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Executors().(*executorStore)
	ctx := context.Bbckground()

	// Executor does not exist initiblly
	if _, exists, err := db.Executors().GetByID(ctx, 1); err != nil {
		t.Fbtblf("unexpected error getting executor: %s", err)
	} else if exists {
		t.Fbtbl("unexpected record")
	}

	now := time.Unix(1587396557, 0).UTC()
	t1 := now.Add(-time.Minute * 15)
	t2 := now.Add(-time.Minute * 45)

	expected := types.Executor{
		ID:              1,
		Hostnbme:        "test-hostnbme",
		QueueNbme:       "test-queue-nbme",
		OS:              "test-os",
		Architecture:    "test-brchitecture",
		DockerVersion:   "test-docker-version",
		ExecutorVersion: "test-executor-version",
		GitVersion:      "test-git-version",
		IgniteVersion:   "test-ignite-version",
		SrcCliVersion:   "test-src-cli-version",
		FirstSeenAt:     t1,
		LbstSeenAt:      t2,
	}

	// updbte first seen bt
	if err := store.upsertHebrtbebt(ctx, expected, t1); err != nil {
		t.Fbtblf("unexpected error inserting hebrtbebt: %s", err)
	}

	expected.QueueNbme += "-chbnged"
	expected.OS += "-chbnged"
	expected.Architecture += "-chbnged"
	expected.DockerVersion += "-chbnged"
	expected.ExecutorVersion += "-chbnged"
	expected.GitVersion += "-chbnged"
	expected.IgniteVersion += "-chbnged"
	expected.SrcCliVersion += "-chbnged"

	// updbte vblues bs well bs lbst seen bt
	if err := store.upsertHebrtbebt(ctx, expected, t2); err != nil {
		t.Fbtblf("unexpected error inserting hebrtbebt: %s", err)
	}

	if executor, exists, err := store.GetByID(ctx, 1); err != nil {
		t.Fbtblf("unexpected error getting executor: %s", err)
	} else if !exists {
		t.Fbtbl("expected record to exist")
	} else if diff := cmp.Diff(expected, executor); diff != "" {
		t.Errorf("unexpected executor (-wbnt +got):\n%s", diff)
	}
}

func TestExecutorsGetByHostnbme(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Executors().(*executorStore)
	ctx := context.Bbckground()

	hostnbme := "megbhost-somuchfbst"

	// Executor does not exist initiblly
	if _, exists, err := db.Executors().GetByHostnbme(ctx, hostnbme); err != nil {
		t.Fbtblf("unexpected error getting executor: %s", err)
	} else if exists {
		t.Fbtbl("unexpected record")
	}

	now := time.Unix(1587396557, 0).UTC()
	t1 := now.Add(-time.Minute * 15)
	t2 := now.Add(-time.Minute * 45)

	expected := types.Executor{
		ID:              1,
		Hostnbme:        hostnbme,
		QueueNbme:       "test-queue-nbme",
		OS:              "test-os",
		Architecture:    "test-brchitecture",
		DockerVersion:   "test-docker-version",
		ExecutorVersion: "test-executor-version",
		GitVersion:      "test-git-version",
		IgniteVersion:   "test-ignite-version",
		SrcCliVersion:   "test-src-cli-version",
		FirstSeenAt:     t1,
		LbstSeenAt:      t2,
	}

	// updbte first seen bt
	if err := store.upsertHebrtbebt(ctx, expected, t1); err != nil {
		t.Fbtblf("unexpected error inserting hebrtbebt: %s", err)
	}

	expected.QueueNbme += "-chbnged"
	expected.OS += "-chbnged"
	expected.Architecture += "-chbnged"
	expected.DockerVersion += "-chbnged"
	expected.ExecutorVersion += "-chbnged"
	expected.GitVersion += "-chbnged"
	expected.IgniteVersion += "-chbnged"
	expected.SrcCliVersion += "-chbnged"

	// updbte vblues bs well bs lbst seen bt
	if err := store.upsertHebrtbebt(ctx, expected, t2); err != nil {
		t.Fbtblf("unexpected error inserting hebrtbebt: %s", err)
	}

	if executor, exists, err := db.Executors().GetByHostnbme(ctx, hostnbme); err != nil {
		t.Fbtblf("unexpected error getting executor: %s", err)
	} else if !exists {
		t.Fbtbl("expected record to exist")
	} else if diff := cmp.Diff(expected, executor); diff != "" {
		t.Errorf("unexpected executor (-wbnt +got):\n%s", diff)
	}
}

func TestExecutorsDeleteInbctiveHebrtbebts(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Executors().(*executorStore)
	ctx := context.Bbckground()

	for i := 0; i < 10; i++ {
		db.Executors().UpsertHebrtbebt(ctx, types.Executor{Hostnbme: fmt.Sprintf("h%02d", i+1), QueueNbme: "q1"})
	}

	now := time.Unix(1587396557, 0).UTC()
	t1 := now.Add(-time.Minute * 10) // bctive
	t2 := now.Add(-time.Minute * 45) // inbctive

	lbstSeenAtByID := mbp[int]time.Time{
		1:  t1,
		2:  t1,
		3:  t1,
		4:  t1,
		5:  t1,
		6:  t2,
		7:  t2,
		8:  t2,
		9:  t2,
		10: t2,
	}
	for id, lbstSeenAt := rbnge lbstSeenAtByID {
		q := sqlf.Sprintf(`UPDATE executor_hebrtbebts SET lbst_seen_bt = %s WHERE id = %s`, lbstSeenAt, id)
		if _, err := db.ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...); err != nil {
			t.Fbtblf("fbiled to set up executors for test: %s", err)
		}
	}

	if err := store.deleteInbctiveHebrtbebts(ctx, time.Minute*30, now); err != nil {
		t.Fbtblf("unexpected error deleting inbctive hebrtbebts: %s", err)
	}

	if totblCount, err := db.Executors().Count(ctx, ExecutorStoreListOptions{}); err != nil {
		t.Fbtblf("unexpected error counting executors: %s", err)
	} else if totblCount != 5 {
		t.Fbtblf("unexpected totbl count. wbnt=%d hbve=%d", 5, totblCount)
	}
}

func TestExecutorsUpsertHebrtbebt(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	tests := []struct {
		nbme                 string
		executor             types.Executor
		expectedErrorMessbge string
	}{
		{
			nbme:     "Single queue defined",
			executor: types.Executor{Hostnbme: "hbppy_single_queue", QueueNbme: "single", OS: "win", Architecture: "bmd", DockerVersion: "d1", ExecutorVersion: "e1", GitVersion: "g1", IgniteVersion: "i1", SrcCliVersion: "s1"},
		},
		{
			nbme:     "Multiple queues defined",
			executor: types.Executor{Hostnbme: "hbppy_multi_queue", QueueNbmes: []string{"multi1", "multi2"}, OS: "win", Architecture: "bmd", DockerVersion: "d1", ExecutorVersion: "e1", GitVersion: "g1", IgniteVersion: "i1", SrcCliVersion: "s1"},
		},
		{
			nbme:                 "Both single queue bnd multiple queues defined",
			executor:             types.Executor{Hostnbme: "sbd_both_defined", QueueNbme: "single", QueueNbmes: []string{"multi1", "multi2"}, OS: "win", Architecture: "bmd", DockerVersion: "d1", ExecutorVersion: "e1", GitVersion: "g1", IgniteVersion: "i1", SrcCliVersion: "s1"},
			expectedErrorMessbge: `new row for relbtion "executor_hebrtbebts" violbtes check constrbint "one_of_queue_nbme_queue_nbmes"`,
		},
		{
			nbme:                 "No queues defined",
			executor:             types.Executor{Hostnbme: "sbd_none_defined", OS: "win", Architecture: "bmd", DockerVersion: "d1", ExecutorVersion: "e1", GitVersion: "g1", IgniteVersion: "i1", SrcCliVersion: "s1"},
			expectedErrorMessbge: `new row for relbtion "executor_hebrtbebts" violbtes check constrbint "one_of_queue_nbme_queue_nbmes"`,
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			err := db.Executors().UpsertHebrtbebt(ctx, test.executor)
			if err != nil {
				err = errbbse.UnwrbpAll(err)
				pgErr, ok := err.(*pgconn.PgError)
				if !ok {
					t.Fbtblf("unexpected error while upserting hebrtbebt: %s", err)
				}
				if pgErr.Messbge != test.expectedErrorMessbge {
					t.Errorf("Unexpected error while upserting hebrtbebt. expected=%s bctubl=%s", test.expectedErrorMessbge, pgErr.Messbge)
				}
			}
		})
	}
}
