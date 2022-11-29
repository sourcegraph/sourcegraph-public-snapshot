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

	q := createRepoMetadataQuery(meta)
	return s.query(ctx, q, func(sc dbutil.Scanner) error {
		return scanRepoMetadata(meta, sc)
	})
}

const createRepoMetadataQueryFmtstr = `
INSERT INTO batch_changes_repo_metadata (
  repo_id,
  created_at,
  updated_at,
  ignored
)
VALUES
  (%s, %s, %s, %s)
RETURNING
  %s
`

func createRepoMetadataQuery(meta *btypes.RepoMetadata) *sqlf.Query {
	return sqlf.Sprintf(
		createRepoMetadataQueryFmtstr,
		meta.ID,
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

	q := getRepoMetadataQuery(repoID)

	m := btypes.RepoMetadata{}
	err = s.query(ctx, q, func(sc dbutil.Scanner) (err error) {
		return scanRepoMetadata(&m, sc)
	})
	if err != nil {
		return nil, err
	}

	if m.ID == 0 {
		return nil, ErrNoResults
	}

	return &m, nil
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

type ListRepoMetadataOpts struct {
	CursorOpts
}

func (s *Store) ListReposMissingMetadata(ctx context.Context, opts ListRepoMetadataOpts) (repoIDs []api.RepoID, cursor int64, err error) {
	ctx, _, endObservation := s.operations.listReposMissingMetadata.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := listReposMissingMetadataQuery(opts)
	ids := []api.RepoID{}
	s.query(ctx, q, func(sc dbutil.Scanner) error {
		var id api.RepoID
		if err := sc.Scan(&id); err != nil {
			return err
		}
		ids = append(ids, id)
		return nil
	})

	return CursorIntResultset(opts.CursorOpts, ids, nil)
}

const listReposMissingMetadataQueryFmtstr = `
SELECT
  id
FROM
  repo
WHERE
  NOT EXISTS (
    SELECT
      repo_id
    FROM
      batch_changes_repo_metadata
    WHERE
      batch_changes_repo_metadata.repo_id = repo.id
  )
  AND %s
ORDER BY
  id ASC
%s -- LIMIT
`

func listReposMissingMetadataQuery(opts ListRepoMetadataOpts) *sqlf.Query {
	return sqlf.Sprintf(
		listReposMissingMetadataQueryFmtstr,
		opts.WhereDB("id", CursorDirectionAscending),
		opts.LimitDB(),
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
