package store

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestDerivativeGraphKey(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := New(observation.TestContextTB(t), db)

	if _, _, err := DerivativeGraphKey(ctx, store); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}
