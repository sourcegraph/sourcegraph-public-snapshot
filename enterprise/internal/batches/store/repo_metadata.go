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

func (s *Store) CreateRepoMetadata(ctx context.Context, meta *btypes.RepoMetadata) (err error) {
	ctx, _, endObservation := s.operations.createRepoMetadata.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	if meta.CreatedAt.IsZero() {
		meta.CreatedAt = s.now()
	}
	if meta.UpdatedAt.IsZero() {
		meta.UpdatedAt = meta.CreatedAt
	}

	return createOrUpdateRecord(ctx, s, createRepoMetadataQuery(meta), scanRepoMetadata, meta)
}

const createRepoMetadataQueryFmtstr = `
INSERT INTO batch_changes_repo_metadata
  (created_at, updated_at, ignored)
VALUES
  (%s, %s, %s)
RETURNING
  %s
`

func createRepoMetadataQuery(meta *btypes.RepoMetadata) *sqlf.Query {
	return sqlf.Sprintf(
		createRepoMetadataQueryFmtstr,
		meta.CreatedAt,
		meta.UpdatedAt,
		meta.Ignored,
		sqlf.Join(repoMetadataColumns, ","),
	)
}

func (s *Store) GetRepoMetadata(ctx context.Context, repoID api.RepoID) (meta *btypes.RepoMetadata, err error) {
	ctx, _, endObservation := s.operations.getRepoMetadata.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("ID", int(meta.ID)),
	}})
	defer endObservation(1, observation.Args{})

	return getRecord(ctx, s, getRepoMetadataQuery(repoID), scanRepoMetadata)
}

const getRepoMetadataQueryFmtstr = `
SELECT
  %s
FROM
  batch_changes_repo_metadata
WHERE
  repo_id = %s
`

func getRepoMetadataQuery(repoID api.RepoID) *sqlf.Query {
	return sqlf.Sprintf(
		getRepoMetadataQueryFmtstr,
		sqlf.Join(repoMetadataColumns, ","),
		repoID,
	)
}

func (s *Store) ListRepoIDsMissingMetadata(ctx context.Context, opts CursorOpts) (repoIDs []api.RepoID, cursor int64, err error) {
	ctx, _, endObservation := s.operations.listRepoIDsMissingMetadata.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := listRepoIDsMissingMetadataQuery(opts)
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
ORDER BY id ASC
%s -- LIMIT
`

func listRepoIDsMissingMetadataQuery(opts CursorOpts) *sqlf.Query {
	return sqlf.Sprintf(
		listRepoIDsMissingMetadataQueryFmtstr,
		opts.WhereDB("id", CursorDirectionAscending),
		opts.LimitDB(),
	)
}

func (s *Store) ListReposWithOutdatedMetadata(ctx context.Context, opts CursorOpts) (metas []*btypes.RepoMetadata, cursor int64, err error) {
	ctx, _, endObservation := s.operations.listReposWithOutdatedMetadata.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return listRecords(ctx, s, listReposWithOutdatedMetadataQuery(opts), opts, scanRepoMetadata)
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
      AND repo.updated_at > batch_changes_repo_metadata.updated_at
  )
  AND %s -- cursor where clause
ORDER BY id ASC
%s -- LIMIT
`

func listReposWithOutdatedMetadataQuery(opts CursorOpts) *sqlf.Query {
	return sqlf.Sprintf(
		listReposWithOutdatedMetadataQueryFmtstr,
		sqlf.Join(repoMetadataColumns, ","),
		opts.WhereDB("id", CursorDirectionAscending),
		opts.LimitDB(),
	)
}

func (s *Store) UpsertRepoMetadata(ctx context.Context, meta *btypes.RepoMetadata) (err error) {
	ctx, _, endObservation := s.operations.upsertRepoMetadata.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("ID", int(meta.ID)),
	}})
	defer endObservation(1, observation.Args{})

	meta.UpdatedAt = s.now()
	return createOrUpdateRecord(ctx, s, upsertRepoMetadataQuery(meta), scanRepoMetadata, meta)
}

const upsertRepoMetadataQueryFmtstr = `
INSERT INTO batch_changes_repo_metadata
  (created_at, updated_at, ignored)
VALUES
  (%s, %s, %s)
ON CONFLICT DO UPDATE SET
  (updated_at, ignored) = (%s, %s)
RETURNING
  %s
`

func upsertRepoMetadataQuery(meta *btypes.RepoMetadata) *sqlf.Query {
	return sqlf.Sprintf(
		upsertRepoMetadataQueryFmtstr,
		meta.CreatedAt,
		meta.UpdatedAt,
		meta.Ignored,
		meta.UpdatedAt,
		meta.Ignored,
		sqlf.Join(repoMetadataColumns, ","),
	)
}

func scanRepoMetadata(meta *btypes.RepoMetadata, sc dbutil.Scanner) error {
	return sc.Scan(
		&meta.ID,
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
