package store

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *Store) GetRepoStatus(ctx context.Context, repoID api.RepoID, commit string) (rs *btypes.RepoStatus, err error) {
	ctx, _, endObservation := s.operations.getRepoStatus.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int32("RepoID", int32(repoID)),
		log.String("Commit", commit),
	}})
	defer endObservation(1, observation.Args{})

	return getRecord(ctx, s, getRepoStatusQuery(repoID, commit), scanRepoStatus)
}

const getRepoStatusQueryFmtstr = `
SELECT
  %s
FROM
  batch_repo_status
INNER JOIN
  repo
ON
  repo.id = batch_repo_status.repo_id
WHERE
  repo_id = %s
  AND commit = %s
`

func getRepoStatusQuery(repoID api.RepoID, commit string) *sqlf.Query {
	return sqlf.Sprintf(
		getRepoStatusQueryFmtstr,
		sqlf.Join(repoStatusColumns, ","),
		repoID,
		commit,
	)
}

func (s *Store) UpsertRepoStatus(ctx context.Context, rs *btypes.RepoStatus) (err error) {
	ctx, _, endObservation := s.operations.upsertRepoStatus.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return createOrUpdateRecord(ctx, s, upsertRepoStatusQuery(rs), scanRepoStatus, rs)
}

const upsertRepoStatusQueryFmtstr = `
INSERT INTO batch_repo_status
  (%s)
VALUES
  (%s, %s, %s)
ON CONFLICT (repo_id) DO UPDATE SET
  (commit, ignored) = (%s, %s)
RETURNING
  %s
`

func upsertRepoStatusQuery(rs *btypes.RepoStatus) *sqlf.Query {
	columns := sqlf.Join(repoStatusColumns, ",")
	return sqlf.Sprintf(
		upsertRepoStatusQueryFmtstr,
		columns,
		rs.RepoID, rs.Commit, rs.Ignored,
		rs.Commit, rs.Ignored,
		columns,
	)
}

var repoStatusColumns = []*sqlf.Query{
	sqlf.Sprintf("repo_id"),
	sqlf.Sprintf("commit"),
	sqlf.Sprintf("ignored"),
}

func scanRepoStatus(rs *btypes.RepoStatus, sc dbutil.Scanner) error {
	return sc.Scan(
		&rs.RepoID,
		&rs.Commit,
		&rs.Ignored,
	)
}
