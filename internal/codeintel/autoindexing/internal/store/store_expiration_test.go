package store

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestExpireFailedRecords(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	insertRepo(t, db, 50, "")

	// TODO - insert indexes

	if err := store.ExpireFailedRecords(context.Background(), time.Hour, time.Now()); err != nil {
		t.Fatalf("unexpected error expiring failed records: %s", err)
	}

	// TODO - assertions
}
