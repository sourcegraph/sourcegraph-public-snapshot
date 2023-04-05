package shared

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type EmbeddingsStore struct {
	basestore.ShareableStore
	logger log.Logger
}

func WrapDB(s *EmbeddingsStore, getter getRepoEmbeddingIndexFn) getRepoEmbeddingIndexFn {
	return func(ctx context.Context, repoName api.RepoName) (*embeddings.RepoEmbeddingIndex, error) {
		// try the database first
		m, err := s.GetEmbeddingsByName(ctx, repoName)
		if err != nil && !errcode.IsNotFound(err) {
			s.logger.Error("failed to fetch embeddings", log.Error(err))
		}
		if err == nil && m != nil {
			return m, nil
		}
		m, err = getter(ctx, repoName)
		if err != nil {
			return nil, err
		}
		// This should probably be made async as it will take time on larger repos.
		if m != nil {
			if err := s.UpdateEmbeddings(ctx, repoName, m); err != nil {
				s.logger.Error("failed to update embeddings", log.Error(err))
			}
		}
		return m, err
	}
}

const updateEmbeddingsFmtstr = `

`

func (s *EmbeddingsStore) UpdateEmbeddings(
	ctx context.Context,
	repoName api.RepoName,
	data *embeddings.RepoEmbeddingIndex,
) error {
	codeMetadata, err := json.Marshal(data.CodeIndex.RowMetadata)
	if err != nil {
		return errors.Wrap(err, "code metadata did not json-serialize")
	}
	textMetadata, err := json.Marshal(data.TextIndex.RowMetadata)
	if err != nil {
		return errors.Wrap(err, "text metadata did not json-serialize")
	}
	q := sqlf.Sprintf(`
        UPDATE repo_embeddings
        SET
            code_index = %s,
            code_column_dimension = %d,
            code_row_metadata = %s,
            code_ranks = %s,
            text_index = %s,
            text_column_dimension = %d,
            text_row_metadata = %s,
            text_ranks = %s
        WHERE repo_id = (
            SELECT id
            FROM repo
            WHERE name = %s
        )
    `,
		pq.Array(data.CodeIndex.Embeddings),
		data.CodeIndex.ColumnDimension,
		codeMetadata,
		pq.Array(data.CodeIndex.Ranks),
		pq.Array(data.TextIndex.Embeddings),
		data.TextIndex.ColumnDimension,
		textMetadata,
		pq.Array(data.TextIndex.Ranks),
		repoName)
	_, err = s.Handle().ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	return err
}

// GetEmbeddingsByName returns the embeddings index for given repo
// or nil, error not found
func (s EmbeddingsStore) GetEmbeddingsByName(ctx context.Context, repoName api.RepoName) (*embeddings.RepoEmbeddingIndex, error) {
	q := sqlf.Sprintf(`
		SELECT
			code_index,
			code_column_dimension,
			code_row_metadata,
			code_ranks
			text_index,
			text_column_dimension,
			text_row_metadata,
			text_ranks
		FROM repo_embeddings
		WHERE repo_id = (
			SELECT id
			FROM repo
			WHERE name = %s
		)
	`, repoName)
	var index embeddings.RepoEmbeddingIndex
	var codeMetadata, textMetadata json.RawMessage
	err := s.Handle().QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(
		pq.Array(&index.CodeIndex.Embeddings),
		&index.CodeIndex.ColumnDimension,
		&codeMetadata,
		pq.Array(&index.CodeIndex.Ranks),
		pq.Array(&index.TextIndex),
		&index.TextIndex.ColumnDimension,
		&textMetadata,
		pq.Array(&index.TextIndex.Ranks))
	if err == sql.ErrNoRows {
		return nil, &embeddingsNotFoundErr{}
	}
	if err := json.Unmarshal(codeMetadata, &index.CodeIndex.RowMetadata); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal code row metadata")
	}
	if err := json.Unmarshal(textMetadata, &index.TextIndex.RowMetadata); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal text row metadata")
	}
	return &index, err
}

type embeddingsNotFoundErr struct{}

func (e embeddingsNotFoundErr) Error() string {
	return "embeddings not found"
}

func (e embeddingsNotFoundErr) NotFound() bool {
	return true
}
