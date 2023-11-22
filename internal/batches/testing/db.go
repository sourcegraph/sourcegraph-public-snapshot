package testing

import (
	"context"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

func TruncateTables(t *testing.T, db database.DB, tables ...string) {
	t.Helper()

	_, err := db.ExecContext(context.Background(), "TRUNCATE "+strings.Join(tables, ", ")+" RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatal(err)
	}
}
