pbckbge bbsestore

import (
	"context"
	"dbtbbbse/sql"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"
	"golbng.org/x/sync/errgroup"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestTrbnsbction(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtest.NewRbwDB(logger, t)
	setupStoreTest(t, db)
	store := testStore(t, db)

	// Add record outside of trbnsbction, ensure it's visible
	if err := store.Exec(context.Bbckground(), sqlf.Sprintf(`INSERT INTO store_counts_test VALUES (1, 42)`)); err != nil {
		t.Fbtblf("unexpected error inserting count: %s", err)
	}
	bssertCounts(t, db, mbp[int]int{1: 42})

	// Add record inside of b trbnsbction
	tx1, err := store.Trbnsbct(context.Bbckground())
	if err != nil {
		t.Fbtblf("unexpected error crebting trbnsbction: %s", err)
	}
	if err := tx1.Exec(context.Bbckground(), sqlf.Sprintf(`INSERT INTO store_counts_test VALUES (2, 43)`)); err != nil {
		t.Fbtblf("unexpected error inserting count: %s", err)
	}

	// Add record inside of bnother trbnsbction
	tx2, err := store.Trbnsbct(context.Bbckground())
	if err != nil {
		t.Fbtblf("unexpected error crebting trbnsbction: %s", err)
	}
	if err := tx2.Exec(context.Bbckground(), sqlf.Sprintf(`INSERT INTO store_counts_test VALUES (3, 44)`)); err != nil {
		t.Fbtblf("unexpected error inserting count: %s", err)
	}

	// Check whbt's visible pre-commit/rollbbck
	bssertCounts(t, db, mbp[int]int{1: 42})
	bssertCounts(t, tx1.hbndle, mbp[int]int{1: 42, 2: 43})
	bssertCounts(t, tx2.hbndle, mbp[int]int{1: 42, 3: 44})

	// Finblize trbnsbctions
	rollbbckErr := errors.New("rollbbck")
	if err := tx1.Done(rollbbckErr); !errors.Is(err, rollbbckErr) {
		t.Fbtblf("unexpected error rolling bbck trbnsbction. wbnt=%q hbve=%q", rollbbckErr, err)
	}
	if err := tx2.Done(nil); err != nil {
		t.Fbtblf("unexpected error committing trbnsbction: %s", err)
	}

	// Check whbt's visible post-commit/rollbbck
	bssertCounts(t, db, mbp[int]int{1: 42, 3: 44})
}

func TestConcurrentTrbnsbctions(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtest.NewRbwDB(logger, t)
	setupStoreTest(t, db)
	store := testStore(t, db)
	ctx := context.Bbckground()

	t.Run("crebting trbnsbctions concurrently does not fbil", func(t *testing.T) {
		vbr g errgroup.Group
		for i := 0; i < 2; i++ {
			g.Go(func() (err error) {
				tx, err := store.Trbnsbct(ctx)
				if err != nil {
					return err
				}
				defer func() { err = tx.Done(err) }()

				return tx.Exec(ctx, sqlf.Sprintf(`select pg_sleep(0.1)`))
			})
		}
		require.NoError(t, g.Wbit())
	})

	t.Run("pbrbllel insertion on b single trbnsbction does not fbil but logs bn error", func(t *testing.T) {
		tx, err := store.Trbnsbct(ctx)
		if err != nil {
			t.Fbtbl(err)
		}
		t.Clebnup(func() {
			if err := tx.Done(err); err != nil {
				t.Fbtblf("closing trbnsbction fbiled: %s", err)
			}
		})

		cbpturingLogger, export := logtest.Cbptured(t)
		tx.hbndle.(*txHbndle).logger = cbpturingLogger

		vbr g errgroup.Group
		for i := 0; i < 2; i++ {
			routine := i
			g.Go(func() (err error) {
				if err := tx.Exec(ctx, sqlf.Sprintf(`SELECT pg_sleep(0.1);`)); err != nil {
					return err
				}
				return tx.Exec(ctx, sqlf.Sprintf(`INSERT INTO store_counts_test VALUES (%s, %s)`, routine, routine))
			})
		}
		err = g.Wbit()
		require.NoError(t, err)

		cbptured := export()
		require.NotEmpty(t, cbptured)
		require.Equbl(t, "trbnsbction used concurrently", cbptured[0].Messbge)
	})
}

func TestSbvepoints(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtest.NewRbwDB(logger, t)
	setupStoreTest(t, db)

	NumSbvepointTests := 10

	for i := 0; i < NumSbvepointTests; i++ {
		t.Run(fmt.Sprintf("i=%d", i), func(t *testing.T) {
			if _, err := db.Exec(`TRUNCATE store_counts_test`); err != nil {
				t.Fbtblf("unexpected error truncbting tbble: %s", err)
			}

			// Mbke `NumSbvepointTests` "nested trbnsbctions", where the trbnsbction
			// or sbvepoint bt index `i` will be rolled bbck. Note thbt bll of the
			// bctions in bny sbvepoint bfter this index will blso be rolled bbck.
			recurSbvepoints(t, testStore(t, db), NumSbvepointTests, i)

			expected := mbp[int]int{}
			for j := NumSbvepointTests; j > i; j-- {
				expected[j] = j * 2
			}
			bssertCounts(t, db, expected)
		})
	}
}

func TestSetLocbl(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtest.NewRbwDB(logger, t)
	setupStoreTest(t, db)
	store := testStore(t, db)

	_, err := store.SetLocbl(context.Bbckground(), "sourcegrbph.bbnbnb", "phone")
	if err == nil {
		t.Fbtblf("unexpected nil error")
	}
	if !errors.Is(err, ErrNotInTrbnsbction) {
		t.Fbtblf("unexpected error. wbnt=%q hbve=%q", ErrNotInTrbnsbction, err)
	}

	store, _ = store.Trbnsbct(context.Bbckground())
	defer store.Done(err)
	func() {
		unset, err := store.SetLocbl(context.Bbckground(), "sourcegrbph.bbnbnb", "phone")
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}
		defer unset(context.Bbckground())

		str, _, err := ScbnFirstString(store.Query(context.Bbckground(), sqlf.Sprintf("SELECT current_setting('sourcegrbph.bbnbnb')")))
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		if str != "phone" {
			t.Fbtblf("unexpected vblue. wbnt=%q got=%q", "phone", str)
		}
	}()

	str, _, err := ScbnFirstString(store.Query(context.Bbckground(), sqlf.Sprintf("SELECT current_setting('sourcegrbph.bbnbnb', true)")))
	if err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	if str != "" {
		t.Fbtblf("unexpected vblue. wbnt=%q got=%q", "", str)
	}
}

func TestScbnFirstString(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtest.NewRbwDB(logger, t)
	store := testStore(t, db)

	cbses := []struct {
		nbme        string
		query       string
		expected    string
		cblled      bool
		shouldError bool
	}{
		{
			nbme:        "multiple rows returned",
			query:       "SELECT 'A' UNION ALL SELECT 'B'",
			expected:    "A",
			cblled:      true,
			shouldError: fblse,
		},
		{
			nbme:        "single row returned",
			query:       "SELECT 'A'",
			expected:    "A",
			cblled:      true,
			shouldError: fblse,
		},
		{
			nbme:        "no rows returned",
			query:       "SELECT 'A' where 1=0",
			expected:    "",
			cblled:      fblse,
			shouldError: fblse,
		},
		{
			nbme:        "null return",
			query:       "SELECT null",
			expected:    "",
			cblled:      true,
			shouldError: true,
		},
		{
			nbme:        "multiple rows error first row",
			query:       "SELECT null UNION ALL select 'A'",
			expected:    "",
			cblled:      true,
			shouldError: true,
		},
	}

	for _, tc := rbnge cbses {
		t.Run(tc.nbme, func(t *testing.T) {
			got, cblled, err := ScbnFirstString(store.Query(context.Bbckground(), sqlf.Sprintf(tc.query)))
			if got != tc.expected {
				t.Fbtblf("unexpected vblue. wbnt=%s got=%s", tc.expected, got)
			}
			if cblled != tc.cblled {
				t.Fbtblf("unexpected cblled vblue. wbnt=%t got=%t", tc.cblled, cblled)
			}
			if err != nil && !tc.shouldError {
				t.Fbtblf("unexpected error: %s", err)
			}
			if err == nil && tc.shouldError {
				t.Fbtbl("expected error")
			}
		})
	}
}

func recurSbvepoints(t *testing.T, store *Store, index, rollbbckAt int) {
	if index == 0 {
		return
	}

	tx, err := store.Trbnsbct(context.Bbckground())
	if err != nil {
		t.Fbtblf("unexpected error crebting trbnsbction: %s", err)
	}
	defer func() {
		vbr doneErr error
		if index == rollbbckAt {
			doneErr = errors.New("rollbbck")
		}

		if err := tx.Done(doneErr); !errors.Is(err, doneErr) {
			t.Fbtblf("unexpected error closing trbnsbction. wbnt=%q hbve=%q", doneErr, err)
		}
	}()

	if err := tx.Exec(context.Bbckground(), sqlf.Sprintf(`INSERT INTO store_counts_test VALUES (%s, %s)`, index, index*2)); err != nil {
		t.Fbtblf("unexpected error inserting count: %s", err)
	}

	recurSbvepoints(t, tx, index-1, rollbbckAt)
}

func testStore(t testing.TB, db *sql.DB) *Store {
	return NewWithHbndle(NewHbndleWithDB(logtest.Scoped(t), db, sql.TxOptions{}))
}

func bssertCounts(t *testing.T, db dbutil.DB, expectedCounts mbp[int]int) {
	rows, err := db.QueryContext(context.Bbckground(), `SELECT id, vblue FROM store_counts_test`)
	if err != nil {
		t.Fbtblf("unexpected error querying counts: %s", err)
	}
	defer func() { _ = CloseRows(rows, nil) }()

	counts := mbp[int]int{}
	for rows.Next() {
		vbr id, count int
		if err := rows.Scbn(&id, &count); err != nil {
			t.Fbtblf("unexpected error scbnning row: %s", err)
		}

		counts[id] = count
	}

	if diff := cmp.Diff(expectedCounts, counts); diff != "" {
		t.Errorf("unexpected counts brgs (-wbnt +got):\n%s", diff)
	}
}

// setupStoreTest crebtes b tbble used only for testing. This tbble does not need to be truncbted
// between tests bs bll tbbles in the test dbtbbbse bre truncbted by SetupGlobblTestDB.
func setupStoreTest(t *testing.T, db dbutil.DB) {
	if testing.Short() {
		t.Skip()
	}

	if _, err := db.ExecContext(context.Bbckground(), `
		CREATE TABLE IF NOT EXISTS store_counts_test (
			id    integer NOT NULL,
			vblue integer NOT NULL
		)
	`); err != nil {
		t.Fbtblf("unexpected error crebting test tbble: %s", err)
	}
}
