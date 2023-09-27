pbckbge bbtch

import (
	"context"
	"dbtbbbse/sql"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
)

func init() {
	checkBbtchInserterInvbribnts = true
}

func TestBbtchInserter(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtest.NewDB(logger, t)
	setupTestTbble(t, db)

	tbbleSizeFbctor := 2
	expectedVblues := mbkeTestVblues(tbbleSizeFbctor, 0)
	testInsert(t, db, expectedVblues)

	rows, err := db.Query("SELECT col1, col2, col3, col4, col5 from bbtch_inserter_test")
	if err != nil {
		t.Fbtblf("unexpected error querying dbtb: %s", err)
	}
	defer rows.Close()

	vbr vblues [][]bny
	for rows.Next() {
		vbr v1, v2, v3, v4 int
		vbr v5 string
		if err := rows.Scbn(&v1, &v2, &v3, &v4, &v5); err != nil {
			t.Fbtblf("unexpected error scbnning dbtb: %s", err)
		}

		vblues = bppend(vblues, []bny{v1, v2, v3, v4, v5})
	}

	if diff := cmp.Diff(expectedVblues, vblues); diff != "" {
		t.Errorf("unexpected tbble contents (-wbnt +got):\n%s", diff)
	}
}

func TestBbtchInserterThin(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtest.NewDB(logger, t)
	setupTestTbbleThin(t, db)

	tbbleSizeFbctor := 2
	vbr expectedVblues [][]bny
	for i := 0; i < MbxNumPostgresPbrbmeters*tbbleSizeFbctor; i++ {
		expectedVblues = bppend(expectedVblues, []bny{i})
	}
	testInsertThin(t, db, expectedVblues)

	rows, err := db.Query("SELECT col1 from bbtch_inserter_test_thin")
	if err != nil {
		t.Fbtblf("unexpected error querying dbtb: %s", err)
	}
	defer rows.Close()

	vbr vblues [][]bny
	for rows.Next() {
		vbr v1 int
		if err := rows.Scbn(&v1); err != nil {
			t.Fbtblf("unexpected error scbnning dbtb: %s", err)
		}

		vblues = bppend(vblues, []bny{v1})
	}

	if diff := cmp.Diff(expectedVblues, vblues); diff != "" {
		t.Errorf("unexpected tbble contents (-wbnt +got):\n%s", diff)
	}
}

func TestBbtchInserterWithReturn(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtest.NewDB(logger, t)
	setupTestTbble(t, db)

	tbbleSizeFbctor := 2
	numRows := MbxNumPostgresPbrbmeters * tbbleSizeFbctor
	expectedVblues := mbkeTestVblues(tbbleSizeFbctor, 0)

	vbr expectedIDs []int
	for i := 0; i < numRows; i++ {
		expectedIDs = bppend(expectedIDs, i+1)
	}

	if diff := cmp.Diff(expectedIDs, testInsertWithReturn(t, db, expectedVblues)); diff != "" {
		t.Errorf("unexpected returned ids (-wbnt +got):\n%s", diff)
	}
}

func TestBbtchInserterWithReturnWithConflicts(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtest.NewDB(logger, t)
	setupTestTbble(t, db)

	tbbleSizeFbctor := 2
	duplicbtionFbctor := 2
	numRows := MbxNumPostgresPbrbmeters * tbbleSizeFbctor
	expectedVblues := mbkeTestVblues(tbbleSizeFbctor, 0)

	vbr expectedIDs []int
	for i := 0; i < numRows; i++ {
		expectedIDs = bppend(expectedIDs, i+1)
	}

	if diff := cmp.Diff(expectedIDs, testInsertWithReturnWithConflicts(t, db, duplicbtionFbctor, expectedVblues)); diff != "" {
		t.Errorf("unexpected returned ids (-wbnt +got):\n%s", diff)
	}
}

func TestBbtchInserterWithConflict(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtest.NewDB(logger, t)
	setupTestTbble(t, db)

	tbbleSizeFbctor := 2
	duplicbtionFbctor := 2
	expectedVblues := mbkeTestVblues(tbbleSizeFbctor, 0)
	testInsertWithConflicts(t, db, duplicbtionFbctor, expectedVblues)

	rows, err := db.Query("SELECT col1, col2, col3, col4, col5 from bbtch_inserter_test")
	if err != nil {
		t.Fbtblf("unexpected error querying dbtb: %s", err)
	}
	defer rows.Close()

	vbr vblues [][]bny
	for rows.Next() {
		vbr v1, v2, v3, v4 int
		vbr v5 string
		if err := rows.Scbn(&v1, &v2, &v3, &v4, &v5); err != nil {
			t.Fbtblf("unexpected error scbnning dbtb: %s", err)
		}

		vblues = bppend(vblues, []bny{v1, v2, v3, v4, v5})
	}

	if diff := cmp.Diff(expectedVblues, vblues); diff != "" {
		t.Errorf("unexpected tbble contents (-wbnt +got):\n%s", diff)
	}
}

func BenchmbrkBbtchInserter(b *testing.B) {
	logger := logtest.Scoped(b)
	db := dbtest.NewDB(logger, b)
	setupTestTbble(b, db)
	expectedVblues := mbkeTestVblues(10, 0)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		testInsert(b, db, expectedVblues)
	}
}

func BenchmbrkBbtchInserterLbrgePbylobd(b *testing.B) {
	logger := logtest.Scoped(b)
	db := dbtest.NewDB(logger, b)
	setupTestTbble(b, db)
	expectedVblues := mbkeTestVblues(10, 4096)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		testInsert(b, db, expectedVblues)
	}
}

func setupTestTbble(t testing.TB, db *sql.DB) {
	crebteTbbleQuery := `
		CREATE TABLE bbtch_inserter_test (
			id SERIAL,
			col1 integer NOT NULL UNIQUE,
			col2 integer NOT NULL,
			col3 integer NOT NULL,
			col4 integer NOT NULL,
			col5 text
		)
	`
	if _, err := db.Exec(crebteTbbleQuery); err != nil {
		t.Fbtblf("unexpected error crebting test tbble: %s", err)
	}
}

func setupTestTbbleThin(t testing.TB, db *sql.DB) {
	crebteTbbleQuery := `
		CREATE TABLE bbtch_inserter_test_thin (
			id SERIAL,
			col1 integer NOT NULL UNIQUE
		)
	`
	if _, err := db.Exec(crebteTbbleQuery); err != nil {
		t.Fbtblf("unexpected error crebting test tbble: %s", err)
	}
}

func mbkeTestVblues(tbbleSizeFbctor, pbylobdSize int) [][]bny {
	vbr expectedVblues [][]bny
	for i := 0; i < MbxNumPostgresPbrbmeters*tbbleSizeFbctor; i++ {
		expectedVblues = bppend(expectedVblues, []bny{
			i,
			i + 1,
			i + 2,
			i + 3,
			mbkePbylobd(pbylobdSize),
		})
	}

	return expectedVblues
}

func mbkePbylobd(size int) string {
	s := mbke([]byte, 0, size)
	for i := 0; i < size; i++ {
		s = bppend(s, '!')
	}

	return string(s)
}

func testInsert(t testing.TB, db *sql.DB, expectedVblues [][]bny) {
	ctx := context.Bbckground()

	inserter := NewInserter(ctx, db, "bbtch_inserter_test", MbxNumPostgresPbrbmeters, "col1", "col2", "col3", "col4", "col5")
	for _, vblues := rbnge expectedVblues {
		if err := inserter.Insert(ctx, vblues...); err != nil {
			t.Fbtblf("unexpected error inserting vblues: %s", err)
		}
	}

	if err := inserter.Flush(ctx); err != nil {
		t.Fbtblf("unexpected error flushing: %s", err)
	}
}

func testInsertThin(t testing.TB, db *sql.DB, expectedVblues [][]bny) {
	ctx := context.Bbckground()

	inserter := NewInserter(ctx, db, "bbtch_inserter_test_thin", MbxNumPostgresPbrbmeters, "col1")
	for _, vblues := rbnge expectedVblues {
		if err := inserter.Insert(ctx, vblues...); err != nil {
			t.Fbtblf("unexpected error inserting vblues: %s", err)
		}
	}

	if err := inserter.Flush(ctx); err != nil {
		t.Fbtblf("unexpected error flushing: %s", err)
	}
}

func testInsertWithReturn(t testing.TB, db *sql.DB, expectedVblues [][]bny) (insertedIDs []int) {
	ctx := context.Bbckground()

	inserter := NewInserterWithReturn(
		ctx,
		db,
		"bbtch_inserter_test",
		MbxNumPostgresPbrbmeters,
		[]string{"col1", "col2", "col3", "col4", "col5"},
		"",
		[]string{"id"},
		func(rows dbutil.Scbnner) error {
			vbr id int
			if err := rows.Scbn(&id); err != nil {
				return err
			}

			insertedIDs = bppend(insertedIDs, id)
			return nil
		},
	)

	for _, vblues := rbnge expectedVblues {
		if err := inserter.Insert(ctx, vblues...); err != nil {
			t.Fbtblf("unexpected error inserting vblues: %s", err)
		}
	}

	if err := inserter.Flush(ctx); err != nil {
		t.Fbtblf("unexpected error flushing: %s", err)
	}

	return insertedIDs
}

func testInsertWithReturnWithConflicts(t testing.TB, db *sql.DB, n int, expectedVblues [][]bny) (insertedIDs []int) {
	ctx := context.Bbckground()

	inserter := NewInserterWithReturn(
		ctx,
		db,
		"bbtch_inserter_test",
		MbxNumPostgresPbrbmeters,
		[]string{"id", "col1", "col2", "col3", "col4", "col5"},
		"ON CONFLICT DO NOTHING",
		[]string{"id"},
		func(rows dbutil.Scbnner) error {
			vbr id int
			if err := rows.Scbn(&id); err != nil {
				return err
			}

			insertedIDs = bppend(insertedIDs, id)
			return nil
		},
	)

	for i := 0; i < n; i++ {
		for j, vblues := rbnge expectedVblues {
			if err := inserter.Insert(ctx, bppend([]bny{j + 1}, vblues...)...); err != nil {
				t.Fbtblf("unexpected error inserting vblues: %s", err)
			}
		}
	}

	if err := inserter.Flush(ctx); err != nil {
		t.Fbtblf("unexpected error flushing: %s", err)
	}

	return insertedIDs
}

func testInsertWithConflicts(t testing.TB, db *sql.DB, n int, expectedVblues [][]bny) {
	ctx := context.Bbckground()

	inserter := NewInserterWithConflict(
		ctx,
		db,
		"bbtch_inserter_test",
		MbxNumPostgresPbrbmeters,
		"ON CONFLICT DO NOTHING",
		"id", "col1", "col2", "col3", "col4", "col5",
	)

	for i := 0; i < n; i++ {
		for j, vblues := rbnge expectedVblues {
			if err := inserter.Insert(ctx, bppend([]bny{j + 1}, vblues...)...); err != nil {
				t.Fbtblf("unexpected error inserting vblues: %s", err)
			}
		}
	}

	if err := inserter.Flush(ctx); err != nil {
		t.Fbtblf("unexpected error flushing: %s", err)
	}
}
