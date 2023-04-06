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
FROM
	embedding_versions ev
INNER JOIN repo AS r ON ev.repo_id = r.id
WHERE r.name = %s
LIMIT 1
`

func (s EmbeddingsStore) HasEmbeddings(ctx context.Context, repoName api.RepoName) (string, error) {
	q := sqlf.Sprintf(
		embeddingsExistFmtstr,
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

const upsertVersionIDFmtstr = `
	WITH ins AS (
		INSERT INTO embedding_versions (repo_id, revision)
		SELECT id, $2
		FROM repo
		WHERE repo.name = $1
		ON CONFLICT (repo_id, revision) DO NOTHING
		RETURNING embedding_versions.id
	)
	SELECT id FROM ins
	UNION ALL
	SELECT ev.id
	FROM embedding_versions AS ev
	INNER JOIN repo AS r ON ev.repo_id = r.id
	WHERE r.name = $1 AND ev.revision = $2
	LIMIT 1;
`

const insertEmbeddingFmtstr = `
	INSERT INTO %s (version_id, embedding, file_name, start_line, end_line, rank)
	VALUES         ($1,         $2,        $3,        $4,         $5,       $6)
`

// inefficient, this should use bulk inserter instead
func insertEmbeddings(ctx context.Context, tableName string, tx *basestore.Store, versionID int32, revision api.CommitID, embeddings []float32, meta embeddings.RepoEmbeddingRowMetadata, rank float32) error {
	q := fmt.Sprintf(insertEmbeddingFmtstr, tableName)
	_, err := tx.Handle().ExecContext(ctx, q, versionID, fmtVector(embeddings), meta.FileName, meta.StartLine, meta.EndLine, rank)
	return err
}

func (s EmbeddingsStore) UpdateEmbeddings(
	ctx context.Context,
	e *embeddings.RepoEmbeddingIndex,
) error {
	var versionID int32
	if err := s.Handle().QueryRowContext(ctx, upsertVersionIDFmtstr, e.RepoName, e.Revision).Scan(&versionID); err != nil {
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
				if err := insertEmbeddings(ctx, tableName, tx, versionID, e.Revision, index.Embeddings[i*index.ColumnDimension:(i+1)*index.ColumnDimension], r, index.Ranks[i]); err != nil {
					return errors.Wrapf(err, "failed to insert embeddings %s", tableName)
				}
			}
		}
		return nil
	})
}

const findVersionIdFmtstr = `
	SELECT ev.id
	FROM embedding_versions AS ev
	INNER JOIN repo AS r ON ev.repo_id = r.id
	WHERE r.name = $1
	AND ev.revision = $2
	LIMIT 1
`

const embeddingsQueryFmtstr = `
SELECT v.file_name, v.start_line, v.end_line
FROM %s AS v
WHERE v.version_id = $1
ORDER BY v.embedding <=> $2::vector
LIMIT $3
`

func (s EmbeddingsStore) QueryEmbeddings(
	ctx context.Context,
	repoName api.RepoName,
	revision string,
	tableName string, // either text_embeddings or code_embeddings HACK HACK HACK
	query []float32,
	n int32,
) ([]embeddings.RepoEmbeddingRowMetadata, error) {
	var versionID int32
	if err := s.Handle().QueryRowContext(ctx, findVersionIdFmtstr, repoName, revision).Scan(&versionID); err != nil {
		return nil, errors.Wrap(err, "failed to find repo x version")
	}
	q := fmt.Sprintf(embeddingsQueryFmtstr, tableName)
	rs, err := s.Handle().QueryContext(ctx, q, versionID, fmtVector(query), n)
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
