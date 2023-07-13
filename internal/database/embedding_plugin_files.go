package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const getEmbeddingPlugingFileFmtStr = `
SELECT %s FROM embedding_plugin_files
WHERE %s
LIMIT 1;
`

type EmbeddingPluginFilesStore interface {
	basestore.ShareableStore

	// Create inserts the given embedding plugin file into the database.
	Create(ctx context.Context, filePath string, contents string, embeddingPluginID int32) (*types.EmbeddingPluginFile, error)

	Get(context.Context, PluginFilesListOpts) (*types.EmbeddingPluginFile, error)
	GetByPluginID(context.Context, int32) ([]*types.EmbeddingPluginFile, error)
}

type embeddingPluginFilesStore struct {
	logger log.Logger
	*basestore.Store
}

func EmbeddingPluginFilesWith(logger log.Logger, other basestore.ShareableStore) EmbeddingPluginFilesStore {
	return &embeddingPluginFilesStore{
		logger: logger,
		Store:  basestore.NewWithHandle(other.Handle()),
	}
}

var listEmbeddingPluginsFilesSQL = `
SELECT id, file_path, contents, embedding_plugin_id FROM embedding_plugin_files
WHERE %s;
`

var listEmbeddingPluginFilesByPluginIDSQL = `
SELECT id, file_path, contents, embedding_plugin_id FROM embedding_plugin_files
WHERE embedding_plugin_id = %d
`

func scanEmbeddingPluginFiles(logger log.Logger, rows *sql.Rows, p *types.EmbeddingPluginFile) error {
	return rows.Scan(
		&p.ID,
		&p.FilePath,
		&p.Contents,
		&p.EmbeddingPluginID,
	)
}

type PluginFileNotFoundErr struct {
	ID       int32
	FilePath string
}

type PluginFilesListOpts struct {
	ByID *struct {
		ID int32
	}
	ByPlugin *struct {
		EmbeddingPluginID int32
		FilePath          string
	}
}

func (e *PluginFileNotFoundErr) Error() string {
	if e.FilePath != "" {
		return fmt.Sprintf("plugin file not found: file_path=%q", e.FilePath)
	}
	if e.ID != 0 {
		return fmt.Sprintf("plugin file not found: id=%d", e.ID)
	}
	return "plugin file not found"
}

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

//	Create creates a new embedding plugin file.
//
// Returns the created embedding plugin file.
// Returns an error if there was an issue creating the embedding plugin file.
func (s *embeddingPluginFilesStore) Create(ctx context.Context, filePath string, contents string, embeddingPluginID int32) (*types.EmbeddingPluginFile, error) {
	q := sqlf.Sprintf(
		embeddingPluginFileCreateQueryFmtStr,
		sqlf.Join(embeddingPluginFileInsertColumns, ", "),
		filePath,
		contents,
		embeddingPluginID,
		// Returning
		sqlf.Join(embeddingPluginFileColumns, ", "),
	)
	ret := &types.EmbeddingPluginFile{}
	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}

	scanEmbeddingPluginFiles(s.logger, rows, ret)
	return ret, nil
}

func (s *embeddingPluginFilesStore) Get(ctx context.Context, opts PluginFilesListOpts) (f *types.EmbeddingPluginFile, err error) {
	tr, ctx := trace.New(ctx, "embedding_plugin_files.Get")
	defer tr.FinishWithErr(&err)

	whereClause := ""

	if opts.ByID != nil {
		whereClause = fmt.Sprintf("id = %d", opts.ByID.ID)
	} else if opts.ByPlugin != nil {
		whereClause = fmt.Sprintf("embedding_plugin_id = %d AND file_path = '%s'", opts.ByPlugin.EmbeddingPluginID, opts.ByPlugin.FilePath)
	} else {
		return nil, errors.New("either ByID or ByPlugin should be set")
	}

	q := sqlf.Sprintf(fmt.Sprintf(listEmbeddingPluginsFilesSQL, whereClause))
	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, &PluginFileNotFoundErr{}
	}

	f = &types.EmbeddingPluginFile{}
	if err := scanEmbeddingPluginFiles(s.logger, rows, f); err != nil {
		return nil, err
	}

	return f, nil
}

func (s *embeddingPluginFilesStore) GetByPluginID(ctx context.Context, pluginID int32) (fs []*types.EmbeddingPluginFile, err error) {
	tr, ctx := trace.New(ctx, "embedding_plugin_files.GetByPluginID")
	defer tr.FinishWithErr(&err)

	q := sqlf.Sprintf(listEmbeddingPluginFilesByPluginIDSQL, pluginID)
	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		pluginFile := &types.EmbeddingPluginFile{}
		if err := scanEmbeddingPluginFiles(s.logger, rows, pluginFile); err != nil {
			return nil, err
		}
		fs = append(fs, pluginFile)
	}

	return fs, nil
}
