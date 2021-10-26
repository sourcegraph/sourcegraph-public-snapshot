package lsifstore

import (
	"os"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

const testBundleID = 39162

func populateTestStore(t testing.TB) *Store {
	contents, err := os.ReadFile("./testdata/lsif-go@ad3507cb.sql")
	if err != nil {
		t.Fatalf("unexpected error reading testdata: %s", err)
	}

	db := dbtest.NewDB(t)

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("unexpected error starting transaction: %s", err)
	}
	defer func() {
		if err := tx.Commit(); err != nil {
			t.Fatalf("unexpected error finishing transaction: %s", err)
		}
	}()

	for _, line := range strings.Split(string(contents), "\n") {
		if line == "" || strings.HasPrefix(line, "---") {
			continue
		}

		if _, err := tx.Exec(line); err != nil {
			t.Fatalf("unexpected error loading database data: %s", err)
		}
	}

	return NewStore(db, &observation.TestContext)
}
