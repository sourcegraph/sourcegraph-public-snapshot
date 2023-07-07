package database

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestEmbeddingPluginFilesGet(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	store := db.EmbeddingPluginFiles()

	t.Run("non-existent file", func(t *testing.T) {
		fileID := int32(100)
		f, err := store.Get(ctx, fileID)
		require.Nil(t, f)
		require.Error(t, err, "embedding plugin file with ID %d not found", fileID)
	})
}
