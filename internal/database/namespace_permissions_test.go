package database

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestNamespacePermission(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)

	t.Run("Create", func(t *testing.T) {
		t.Parallel()
		db := NewDB(logger, dbtest.NewDB(logger, t))
		// _, userID, ctx := newTestUser(ctx, t, db)
	})
}
