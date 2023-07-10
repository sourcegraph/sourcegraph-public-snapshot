package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type EmbeddingPluginFileStore interface {
	basestore.ShareableStore

	// Create inserts the given embedding plugin file into the database.
	Create(ctx context.Context, filePath string, contents []byte, embeddingPluginID int32) (*types.EmbeddingPluginFile, error)
	// Get returns the embedding plugin file matching the given ID provided. If no such record exists,
	// a EmbeddingPluginFileNotFoundErr is returned.
	Get(ctx context.Context, id int32) (*types.EmbeddingPluginFile, error)
}

var embeddingPluginFileColumns = []*sqlf.Query{
	sqlf.Sprintf("embedding_plugin_files.id"),
	sqlf.Sprintf("embedding_plugin_files.file_path"),
	sqlf.Sprintf("embedding_plugin_files.contents"),
	sqlf.Sprintf("embedding_plugin_files.embedding_plugin_id"),
}

var embeddingPluginFileInsertColumns = []*sqlf.Query{
	sqlf.Sprintf("file_path"),
	sqlf.Sprintf("contents"),
	sqlf.Sprintf("embedding_plugin_id"),
}

func EmbeddingPluginFilesWith(other basestore.ShareableStore) EmbeddingPluginFileStore {
	return &embeddingPluginFileStore{Store: basestore.NewWithHandle(other.Handle())}
}

type EmbeddingPluginFileNotFoundErr struct {
	ID int32
}

func (e *EmbeddingPluginFileNotFoundErr) Error() string {
	return fmt.Sprintf("embedding plugin file with ID %d not found", e.ID)
}

func (e *EmbeddingPluginFileNotFoundErr) NotFound() bool {
	return true
}

type embeddingPluginFileStore struct {
	*basestore.Store
}

var _ EmbeddingPluginFileStore = &embeddingPluginFileStore{}

const embeddingPluginFileCreateQueryFmtStr = `
INSERT INTO
	embedding_plugin_files (%s)
	VALUES (
		%s,
		%s,
		%s
	)
	RETURNING %s
`

//	Create creates a new embedding plugin file.
//
// Returns the created embedding plugin file.
// Returns an error if there was an issue creating the embedding plugin file.
func (e *embeddingPluginFileStore) Create(ctx context.Context, filePath string, contents []byte, embeddingPluginID int32) (*types.EmbeddingPluginFile, error) {
	q := sqlf.Sprintf(
		embeddingPluginFileCreateQueryFmtStr,
		sqlf.Join(embeddingPluginFileInsertColumns, ", "),
		filePath,
		contents,
		embeddingPluginID,
		// Returning
		sqlf.Join(embeddingPluginFileColumns, ", "),
	)
	return scanEmbeddingPluginFile(e.QueryRow(ctx, q))
}

func scanEmbeddingPluginFile(s dbutil.Scanner) (value *types.EmbeddingPluginFile, err error) {
	err = s.Scan(&value.ID, &value.FilePath, &value.Contents, &value.EmbeddingPluginID)
	return
}

const getEmbeddingPlugingFileFmtStr = `
SELECT %s FROM embedding_plugin_files
WHERE %s
LIMIT 1;
`

//	Get returns the embedding plugin file with the given ID.
//
// Returns the embedding plugin file with the given ID.
// Returns EmbeddingPluginFileNotFoundErr if no embedding plugin file exists with the given ID.
// Returns an error if there was an issue querying the database.
func (e *embeddingPluginFileStore) Get(ctx context.Context, id int32) (*types.EmbeddingPluginFile, error) {
	q := sqlf.Sprintf(
		getEmbeddingPlugingFileFmtStr,
		sqlf.Join(embeddingPluginFileColumns, ", "),
		sqlf.Sprintf("id = %d", id),
	)

	epf, err := scanEmbeddingPluginFile(e.QueryRow(ctx, q))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &EmbeddingPluginFileNotFoundErr{ID: id}
		}
		return nil, errors.Wrap(err, "scanning embedding plugin file")
	}

	return epf, nil
}
