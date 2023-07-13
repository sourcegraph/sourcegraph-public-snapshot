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

	pluginID := int32(200)
	_, err := store.Create(ctx, "/internal/file.md", "test", pluginID)
	require.NoError(t, err)

	// t.Run("non-existent file", func(t *testing.T) {
	// 	fileID := int32(100)
	// 	epf, err := store.Get(ctx, fileID)
	// 	require.Nil(t, epf)
	// 	require.Error(t, err, "embedding plugin file with ID %d not found", fileID)
	// })

	// t.Run("existing file", func(t *testing.T) {
	// 	epf, err := store.Get(ctx, newFile.ID)
	// 	require.NoError(t, err)
	// 	require.Equal(t, newFile.ID, epf.ID)
	// 	require.Equal(t, newFile.FilePath, epf.FilePath)
	// 	require.Equal(t, newFile.Contents, epf.Contents)
	// 	require.Equal(t, newFile.EmbeddingPluginID, epf.EmbeddingPluginID)
	// })
}

func TestEmbeddingPluginFilesCreate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	store := db.EmbeddingPluginFiles()

	contents := `
	# Hello
This is a markdown file.`
	filePath := "/internal/memos/test.md"
	pluginID := int32(34)

	epf, err := store.Create(ctx, filePath, contents, pluginID)
	require.NoError(t, err)
	require.Equal(t, filePath, epf.FilePath)
	require.Equal(t, contents, epf.Contents)
	require.Equal(t, epf.EmbeddingPluginID, pluginID)
}
