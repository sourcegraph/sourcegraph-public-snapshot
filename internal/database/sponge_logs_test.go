package database

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestSponge(t *testing.T) {
	ctx := actor.WithInternalActor(context.Background())
	logger := logtest.NoOp(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	t.Run("save and read", func(t *testing.T) {
		spongeLog := SpongeLog{
			ID:          uuid.New(),
			Text:        "example test log",
			Interpreter: "test",
		}
		require.NoError(t, db.SpongeLogs().Save(ctx, spongeLog))
		gotLog, err := db.SpongeLogs().ByID(ctx, spongeLog.ID)
		require.NoError(t, err)
		require.Equal(t, spongeLog, gotLog)
	})
}
