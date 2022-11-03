package store

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestReconcileCandidates(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)
	ctx := context.Background()

	// TODO - setup test

	_, err := store.ReconcileCandidates(ctx, 50)
	if err != nil {
		t.Fatalf("failed to get candidate IDs for reconciliation: %s", err)
	}

	// TODO - assert
	t.Fail()
}
