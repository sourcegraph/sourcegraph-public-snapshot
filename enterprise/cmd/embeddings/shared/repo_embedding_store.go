package shared

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type EmbeddingsStore struct {
	*basestore.Store
	logger log.Logger
}

const embeddingsExistFmtstr = `
	SELECT revision
	FROM (
		SELECT revision
		FROM text_embeddings AS te
		INNER JOIN repo AS r ON te.repo_id = r.id
		WHERE r.name = %s

		UNION ALL

		SELECT revision
		FROM code_embeddings AS ce
		INNER JOIN repo AS r ON ce.repo_id = r.id
		WHERE r.name = %s
	) AS revisions
	LIMIT 1
`

func (s EmbeddingsStore) HasEmbeddings(ctx context.Context, repoName api.RepoName) (string, error) {
	q := sqlf.Sprintf(
		embeddingsExistFmtstr,
		repoName,
		repoName,
	)
	var rev string
	err := s.Handle().QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&rev)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return rev, nil
}

const insertEmbeddingFmtstr = `
	WITH ins_vector AS (
		INSERT INTO embedding_vectors (repo_id, embedding)
		VALUES ($1, $2::vector)
		RETURNING id AS embedding_id
	)
	INSERT INTO %s (embedding_id, repo_id, revision, file_name, start_line, end_line, rank)
	SELECT          embedding_id, $1,      $3,       $4,        $5,         $6,       $7
	FROM ins_vector
`

// inefficient, this should use bulk inserter instead
func insertEmbeddings(ctx context.Context, tableName string, tx *basestore.Store, repoID api.RepoID, revision api.CommitID, embeddings []float32, meta embeddings.RepoEmbeddingRowMetadata, rank float32) error {
	q := fmt.Sprintf(insertEmbeddingFmtstr, tableName)
	_, err := tx.Handle().ExecContext(ctx, q, repoID, fmtVector(embeddings), revision, meta.FileName, meta.StartLine, meta.EndLine, rank)
	return err
}

func (s EmbeddingsStore) UpdateEmbeddings(
	ctx context.Context,
	e *embeddings.RepoEmbeddingIndex,
) error {
	var repoID int32
	if err := s.Handle().QueryRowContext(ctx, "SELECT id FROM repo WHERE name = $1", e.RepoName).Scan(&repoID); err != nil {
		return errors.Wrapf(err, "cannot find repo '%s'", e.RepoName)
	}
	return s.WithTransact(ctx, func(tx *basestore.Store) error {
		// TODO drop embeddings before inserting
		// TODO bulk insert is more efficient than going row by row
		for tableName, index := range map[string]embeddings.EmbeddingIndex{
			"code_embeddings": e.CodeIndex,
			"text_embeddings": e.TextIndex,
		} {
			for i, r := range index.RowMetadata {
				if err := insertEmbeddings(ctx, tableName, tx, api.RepoID(repoID), e.Revision, index.Embeddings[i*index.ColumnDimension:(i+1)*index.ColumnDimension], r, index.Ranks[i]); err != nil {
					return errors.Wrapf(err, "failed to insert embeddings %s", tableName)
				}
			}
		}
		return nil
	})
}

const embeddingsQueryFmtstr = `
	SELECT m.file_name, m.start_line, m.end_line
	FROM embedding_vectors AS v
	INNER JOIN %s AS m ON v.id = m.embedding_id
	INNER JOIN repo AS r ON r.id = m.repo_id
	WHERE r.name = $1
	AND m.revision = $2
	ORDER BY v.embedding <=> $3::vector
	LIMIT $4
`

func (s EmbeddingsStore) QueryEmbeddings(
	ctx context.Context,
	repoName api.RepoName,
	revision string,
	tableName string, // either text_embeddings or code_embeddings HACK HACK HACK
	query []float32,
	n int32,
) ([]embeddings.RepoEmbeddingRowMetadata, error) {
	q := fmt.Sprintf(embeddingsQueryFmtstr, tableName)
	rs, err := s.Handle().QueryContext(ctx, q, repoName, revision, fmtVector(query), n)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	var ms []embeddings.RepoEmbeddingRowMetadata
	for rs.Next() {
		var m embeddings.RepoEmbeddingRowMetadata
		if err := rs.Scan(&m.FileName, &m.StartLine, &m.EndLine); err != nil {
			return nil, err
		}
		ms = append(ms, m)
	}
	return ms, nil
}

func fmtVector(fs []float32) string {
	var b bytes.Buffer
	fmt.Fprint(&b, "[")
	var notFirst bool
	for _, f := range fs {
		if notFirst {
			fmt.Fprint(&b, ",")
		}
		fmt.Fprintf(&b, "%.9f", f)
		notFirst = true
	}
	fmt.Fprint(&b, "]")
	return b.String()
}

type embeddingsNotFoundErr struct{}

func (e embeddingsNotFoundErr) Error() string {
	return "embeddings not found"
}

func (e embeddingsNotFoundErr) NotFound() bool {
	return true
}
