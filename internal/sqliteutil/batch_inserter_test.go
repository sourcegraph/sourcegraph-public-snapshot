package sqliteutil

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
)

func init() {
	SetLocalLibpath()
	MustRegisterSqlite3WithPcre()
}

func TestBatchInserter(t *testing.T) {
	ctx := context.Background()

	var expectedValues [][]interface{}
	for i := 0; i < 1000; i++ {
		expectedValues = append(expectedValues, []interface{}{i, i + 1, i + 2, i + 3, i + 4})
	}

	withTestDB(t, func(db *sqlx.DB) error {
		inserter := NewBatchInserter(db, "test", "col1", "col2", "col3", "col4", "col5")
		for _, values := range expectedValues {
			if err := inserter.Insert(ctx, values...); err != nil {
				return err
			}
		}

		if err := inserter.Flush(ctx); err != nil {
			return err
		}

		rows, err := db.Query("SELECT col1, col2, col3, col4, col5 from test")
		if err != nil {
			return err
		}
		defer rows.Close()

		var values [][]interface{}
		for rows.Next() {
			var v1, v2, v3, v4, v5 int
			if err := rows.Scan(&v1, &v2, &v3, &v4, &v5); err != nil {
				return err
			}

			values = append(values, []interface{}{v1, v2, v3, v4, v5})
		}

		if diff := cmp.Diff(expectedValues, values); diff != "" {
			t.Errorf("unexpected table contents (-want +got):\n%s", diff)
		}

		return nil
	})
}
func BenchmarkSQLiteInsertion(b *testing.B) {
	var expectedValues [][]interface{}
	for i := 0; i < b.N; i++ {
		expectedValues = append(expectedValues, []interface{}{i, i + 1, i + 2, i + 3, i + 4})
	}

	withTestDB(b, func(db *sqlx.DB) error {
		b.ResetTimer()

		for _, values := range expectedValues {
			if _, err := db.Exec("INSERT INTO test (col1, col2, col3, col4, col5) VALUES (?, ?, ?, ?, ?)", values...); err != nil {
				return err
			}
		}

		return nil
	})
}

func BenchmarkSQLiteInsertionInTransaction(b *testing.B) {
	ctx := context.Background()

	var expectedValues [][]interface{}
	for i := 0; i < b.N; i++ {
		expectedValues = append(expectedValues, []interface{}{i, i + 1, i + 2, i + 3, i + 4})
	}

	withTestDB(b, func(db *sqlx.DB) error {
		b.ResetTimer()

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		for _, values := range expectedValues {
			if _, err := tx.Exec("INSERT INTO test (col1, col2, col3, col4, col5) VALUES (?, ?, ?, ?, ?)", values...); err != nil {
				return err
			}
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		return nil
	})
}

func BenchmarkSQLiteInsertionWithBatchInserter(b *testing.B) {
	ctx := context.Background()

	var expectedValues [][]interface{}
	for i := 0; i < b.N; i++ {
		expectedValues = append(expectedValues, []interface{}{i, i + 1, i + 2, i + 3, i + 4})
	}

	withTestDB(b, func(db *sqlx.DB) error {
		b.ResetTimer()

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		inserter := NewBatchInserter(tx, "test", "col1", "col2", "col3", "col4", "col5")
		for _, values := range expectedValues {
			if err := inserter.Insert(ctx, values...); err != nil {
				return err
			}
		}

		if err := inserter.Flush(ctx); err != nil {
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		return nil
	})
}

func withTestDB(t testing.TB, test func(db *sqlx.DB) error) {
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("unexpected error creating temp directory: %s", err)
	}
	defer os.RemoveAll(tempDir)

	db, err := sqlx.Open("sqlite3_with_pcre", filepath.Join(tempDir, "batch.db"))
	if err != nil {
		t.Fatalf("unexpected error opening database: %s", err)
	}

	createTableQuery := `
		CREATE TABLE test (
			id integer primary key not null,
			col1 integer not null,
			col2 integer not null,
			col3 integer not null,
			col4 integer not null,
			col5 integer not null
		)
	`
	_, err1 := db.Exec(createTableQuery)
	_, err2 := db.Exec("PRAGMA synchronous = OFF")
	_, err3 := db.Exec("PRAGMA journal_mode = OFF")

	for _, err := range []error{err1, err2, err3} {
		if err != nil {
			t.Fatalf("unexpected error setting up database: %s", err)
		}
	}

	if err := test(db); err != nil {
		t.Fatalf("unexpected error running test: %s", err)
	}
}
