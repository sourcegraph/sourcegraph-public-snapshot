package testing

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

func TruncateTables(t *testing.T, db *sql.DB, tables ...string) {
	t.Helper()

	_, err := db.ExecContext(context.Background(), "TRUNCATE "+strings.Join(tables, ", ")+" RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatal(err)
	}
}
