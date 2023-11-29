package servegit

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

// TestEnsureExtSVC is a light integration test just to check we successfully
// insert into the DB.
func TestEnsureExtSVC(t *testing.T) {
	logger := logtest.Scoped(t)
	testDB := database.NewDB(logger, dbtest.NewDB(t))
	store := testDB.ExternalServices()

	err := doEnsureExtSVC(context.Background(), store, "http://test", "/fake")
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.GetByID(context.Background(), ExtSVCID)
	if err != nil {
		t.Fatal(err)
	}
}
