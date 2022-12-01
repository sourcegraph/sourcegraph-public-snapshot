package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (s *Store) GetRepoMetadata(ctx context.Context, repoID api.RepoID) (meta *btypes.RepoMetadata, err error) {
	ctx, _, endObservation := s.operations.getRepoMetadata.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("ID", int(repoID)),
	}})
	defer endObservation(1, observation.Args{})

	repoAuthzConds, err := database.AuthzQueryConds(ctx, database.NewDBWith(s.logger, s))
	if err != nil {
		return nil, errors.Wrap(err, "GetRepoMetadata generating authz query conds")
	}

	return getRecord(ctx, s, getRepoMetadataQuery(repoID, repoAuthzConds), scanRepoMetadata)
}

const getRepoMetadataQueryFmtstr = `
SELECT
  %s
FROM
  batch_changes_repo_metadata
WHERE
  repo_id = %s
  AND EXISTS (
    SELECT 1
    FROM
      repo
    WHERE
      repo.id = batch_changes_repo_metadata.repo_id
      AND %s -- authz
  )
`

func getRepoMetadataQuery(repoID api.RepoID, repoAuthzConds *sqlf.Query) *sqlf.Query {
	return sqlf.Sprintf(
		getRepoMetadataQueryFmtstr,
		sqlf.Join(repoMetadataColumns, ","),
		repoID,
		repoAuthzConds,
	)
}

func (s *Store) ListRepoIDsMissingMetadata(ctx context.Context, opts CursorOpts) (repoIDs []api.RepoID, cursor int64, err error) {
	ctx, _, endObservation := s.operations.listRepoIDsMissingMetadata.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	repoAuthzConds, err := database.AuthzQueryConds(ctx, database.NewDBWith(s.logger, s))
	if err != nil {
		return nil, 0, errors.Wrap(err, "ListRepoIDsMissingMetadata generating authz query conds")
	}

	q := listRepoIDsMissingMetadataQuery(opts, repoAuthzConds)
	ids := []api.RepoID{}
	err = s.query(ctx, q, func(sc dbutil.Scanner) error {
		var id api.RepoID
		if err := sc.Scan(&id); err != nil {
			return err
		}
		ids = append(ids, id)
		return nil
	})

	return CursorIntResultset(opts, ids, err)
}

const listRepoIDsMissingMetadataQueryFmtstr = `
SELECT
  id
FROM
  repo
WHERE
  NOT EXISTS (
    SELECT 1
    FROM
      batch_changes_repo_metadata
    WHERE
      batch_changes_repo_metadata.repo_id = repo.id
  )
  AND %s -- cursor where clause
  AND %s -- authz
ORDER BY id ASC
%s -- LIMIT
`

func listRepoIDsMissingMetadataQuery(opts CursorOpts, repoAuthzConds *sqlf.Query) *sqlf.Query {
	return sqlf.Sprintf(
		listRepoIDsMissingMetadataQueryFmtstr,
		opts.WhereDB("id", CursorDirectionAscending),
		repoAuthzConds,
		opts.LimitDB(),
	)
}

func (s *Store) ListReposWithOutdatedMetadata(ctx context.Context, opts CursorOpts) (metas []*btypes.RepoMetadata, cursor int64, err error) {
	ctx, _, endObservation := s.operations.listReposWithOutdatedMetadata.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	repoAuthzConds, err := database.AuthzQueryConds(ctx, database.NewDBWith(s.logger, s))
	if err != nil {
		return nil, 0, errors.Wrap(err, "ListReposWithOutdatedMetadata generating authz query conds")
	}

	return listRecords(ctx, s, listReposWithOutdatedMetadataQuery(opts, repoAuthzConds), opts, scanRepoMetadata)
}

const listReposWithOutdatedMetadataQueryFmtstr = `
SELECT
  %s
FROM
  batch_changes_repo_metadata
WHERE
  EXISTS (
    SELECT 1
    FROM
      repo
    WHERE
      repo.id = batch_changes_repo_metadata.repo_id
      AND (
        (repo.updated_at IS NOT NULL AND repo.updated_at > batch_changes_repo_metadata.updated_at)
        OR
        (repo.created_at > batch_changes_repo_metadata.updated_at)
      )
      AND %s -- authz
  )
  AND %s -- cursor where clause
ORDER BY repo_id ASC
%s -- LIMIT
`

func listReposWithOutdatedMetadataQuery(opts CursorOpts, repoAuthzConds *sqlf.Query) *sqlf.Query {
	return sqlf.Sprintf(
		listReposWithOutdatedMetadataQueryFmtstr,
		sqlf.Join(repoMetadataColumns, ","),
		repoAuthzConds,
		opts.WhereDB("id", CursorDirectionAscending),
		opts.LimitDB(),
	)
}

func (s *Store) UpsertRepoMetadata(ctx context.Context, meta *btypes.RepoMetadata) (err error) {
	ctx, _, endObservation := s.operations.upsertRepoMetadata.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("ID", int(meta.RepoID)),
	}})
	defer endObservation(1, observation.Args{})

	now := s.now()
	meta.UpdatedAt = now
	return createOrUpdateRecord(ctx, s, upsertRepoMetadataQuery(meta, now), scanRepoMetadata, meta)
}

const upsertRepoMetadataQueryFmtstr = `
INSERT INTO batch_changes_repo_metadata
  (%s)
VALUES
  (%s, %s, %s, %s)
ON CONFLICT (repo_id) DO UPDATE SET
  (updated_at, ignored) = (%s, %s)
RETURNING
  %s
`

func upsertRepoMetadataQuery(meta *btypes.RepoMetadata, now time.Time) *sqlf.Query {
	columns := sqlf.Join(repoMetadataColumns, ",")

	return sqlf.Sprintf(
		upsertRepoMetadataQueryFmtstr,
		columns,
		meta.RepoID,
		now,
		meta.UpdatedAt,
		meta.Ignored,
		meta.UpdatedAt,
		meta.Ignored,
		columns,
	)
}

func scanRepoMetadata(meta *btypes.RepoMetadata, sc dbutil.Scanner) error {
	return sc.Scan(
		&meta.RepoID,
		&meta.CreatedAt,
		&meta.UpdatedAt,
		&meta.Ignored,
	)
}

var repoMetadataColumns = []*sqlf.Query{
	sqlf.Sprintf("repo_id"),
	sqlf.Sprintf("created_at"),
	sqlf.Sprintf("updated_at"),
	sqlf.Sprintf("ignored"),
}
