package database

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type EmbeddingsJobsStore interface {
	basestore.ShareableStore

	GetEmbeddingsJob(ctx context.Context, repo api.RepoID) (*EmbeddingsJob, error)
}

type embeddingsJobsStore struct {
	*basestore.Store
}

func EmbeddingsJobsStoreWith(other basestore.ShareableStore) EmbeddingsJobsStore {
	return &embeddingsJobsStore{Store: basestore.NewWithHandle(other.Handle())}
}

type EmbeddingsJob struct {
	Revision api.CommitID
}

const getEmbeddingsJobQueryFmtstr = `
SELECT revision
FROM repo_embedding_jobs
WHERE repo_id = %s
AND state = 'completed'
AND failure_message IS NULL
ORDER BY finished_at DESC
LIMIT 1;
`

func (s *embeddingsJobsStore) GetEmbeddingsJob(ctx context.Context, repo api.RepoID) (*EmbeddingsJob, error) {
	return scanEmbeddingsJob(s.QueryRow(ctx, sqlf.Sprintf(getEmbeddingsJobQueryFmtstr, repo)))
}

func scanEmbeddingsJob(sc dbutil.Scanner) (*EmbeddingsJob, error) {
	var (
		revision api.CommitID
	)

	if err := sc.Scan(
		&revision,
	); err != nil {
		return nil, err
	}

	return &EmbeddingsJob{
		Revision: revision,
	}, nil
}
