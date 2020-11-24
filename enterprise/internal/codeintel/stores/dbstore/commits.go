package dbstore

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
	"github.com/sourcegraph/sourcegraph/internal/db/batch"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// scanCommitGraphView scans a commit graph view from the return value of `*Store.query`.
func scanCommitGraphView(rows *sql.Rows, queryErr error) (_ *CommitGraphView, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	commitGraphView := NewCommitGraphView()

	for rows.Next() {
		var meta UploadMeta
		var commit, token string
		var ancestorVisible, overwritten bool

		if err := rows.Scan(&meta.UploadID, &commit, &token, &meta.Flags, &ancestorVisible, &overwritten); err != nil {
			return nil, err
		}

		if ancestorVisible {
			meta.Flags |= FlagAncestorVisible
		}
		if overwritten {
			meta.Flags |= FlagOverwritten
		}

		commitGraphView.Add(meta, commit, token)
	}

	return commitGraphView, nil
}

// HasRepository determines if there is LSIF data for the given repository.
func (s *Store) HasRepository(ctx context.Context, repositoryID int) (_ bool, err error) {
	ctx, endObservation := s.operations.hasRepository.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	count, _, err := basestore.ScanFirstInt(s.Store.Query(ctx, sqlf.Sprintf(`
		SELECT COUNT(*)
		FROM lsif_uploads
		WHERE state != 'deleted' AND repository_id = %s
		LIMIT 1
	`, repositoryID)))

	return count > 0, err
}

// HasCommit determines if the given commit is known for the given repository.
func (s *Store) HasCommit(ctx context.Context, repositoryID int, commit string) (_ bool, err error) {
	ctx, endObservation := s.operations.hasCommit.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("commit", commit),
	}})
	defer endObservation(1, observation.Args{})

	count, _, err := basestore.ScanFirstInt(s.Store.Query(ctx, sqlf.Sprintf(`
		SELECT COUNT(*)
		FROM lsif_nearest_uploads
		WHERE repository_id = %s AND commit_bytea = %s AND NOT overwritten
		LIMIT 1
	`, repositoryID, dbutil.CommitBytea(commit))))

	return count > 0, err
}

// MarkRepositoryAsDirty marks the given repository's commit graph as out of date.
func (s *Store) MarkRepositoryAsDirty(ctx context.Context, repositoryID int) (err error) {
	ctx, endObservation := s.operations.markRepositoryAsDirty.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	return s.Store.Exec(
		ctx,
		sqlf.Sprintf(`
			INSERT INTO lsif_dirty_repositories (repository_id, dirty_token, update_token)
			VALUES (%s, 1, 0)
			ON CONFLICT (repository_id) DO UPDATE SET dirty_token = lsif_dirty_repositories.dirty_token + 1
		`, repositoryID),
	)
}

func scanIntPairs(rows *sql.Rows, queryErr error) (_ map[int]int, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	values := map[int]int{}
	for rows.Next() {
		var value1 int
		var value2 int
		if err := rows.Scan(&value1, &value2); err != nil {
			return nil, err
		}

		values[value1] = value2
	}

	return values, nil
}

// DirtyRepositories returns a map from repository identifiers to a dirty token for each repository whose commit
// graph is out of date. This token should be passed to CalculateVisibleUploads in order to unmark the repository.
func (s *Store) DirtyRepositories(ctx context.Context) (_ map[int]int, err error) {
	ctx, endObservation := s.operations.dirtyRepositories.With(ctx, &err, observation.Args{LogFields: []log.Field{}})
	defer endObservation(1, observation.Args{})

	return scanIntPairs(s.Store.Query(ctx, sqlf.Sprintf(`SELECT repository_id, dirty_token FROM lsif_dirty_repositories WHERE dirty_token > update_token`)))
}

// CalculateVisibleUploads uses the given commit graph and the tip commit of the default branch to determine the set
// of LSIF uploads that are visible for each commit, and the set of uploads which are visible at the tip. The decorated
// commit graph is serialized to Postgres for use by find closest dumps queries.
//
// If dirtyToken is supplied, the repository will be unmarked when the supplied token does matches the most recent
// token stored in the database, the flag will not be cleared as another request for update has come in since this
// token has been read.
func (s *Store) CalculateVisibleUploads(ctx context.Context, repositoryID int, graph map[string][]string, tipCommit string, dirtyToken int) (err error) {
	ctx, endObservation := s.operations.calculateVisibleUploads.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.Int("numKeys", len(graph)),
		log.String("tipCommit", tipCommit),
		log.Int("dirtyToken", dirtyToken),
	}})
	defer endObservation(1, observation.Args{})

	tx, err := s.transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Pull all queryable upload metadata known to this repository so we can correlate
	// it with the current  commit graph.
	commitGraphView, err := scanCommitGraphView(tx.Store.Query(ctx, sqlf.Sprintf(`
		SELECT id, commit, md5(root || ':' || indexer) as token, 0 as distance, true as ancestor_visible, false as overwritten
		FROM lsif_uploads
		WHERE state = 'completed' AND repository_id = %s
	`, repositoryID)))
	if err != nil {
		return err
	}

	// Determine which uploads are visible to which commits for this repository
	visibleUploads, err := calculateVisibleUploads(graph, commitGraphView)
	if err != nil {
		return err
	}

	// Clear all old visibility data for this repository
	for _, query := range []string{
		`DELETE FROM lsif_nearest_uploads WHERE repository_id = %s`,
		`DELETE FROM lsif_uploads_visible_at_tip WHERE repository_id = %s`,
	} {
		if err := tx.Store.Exec(ctx, sqlf.Sprintf(query, repositoryID)); err != nil {
			return err
		}
	}

	nearestUploadsInserter := batch.NewBatchInserter(
		ctx,
		s.Store.Handle().DB(),
		"lsif_nearest_uploads",
		"repository_id",
		"commit_bytea",
		"upload_id",
		"distance",
		"ancestor_visible",
		"overwritten",
	)
	for commit, uploads := range visibleUploads {
		for _, uploadMeta := range uploads {
			if err := nearestUploadsInserter.Insert(
				ctx,
				repositoryID,
				dbutil.CommitBytea(commit),
				uploadMeta.UploadID,
				uploadMeta.Flags&MaxDistance,
				(uploadMeta.Flags&FlagAncestorVisible) != 0,
				(uploadMeta.Flags&FlagOverwritten) != 0,
			); err != nil {
				return err
			}
		}
	}
	if err := nearestUploadsInserter.Flush(ctx); err != nil {
		return err
	}

	// Update which repositories are visible from the tip of the default branch. This
	// flag is used to determine which bundles for a repository we open during a global
	// find references query.
	uploadsVisibleAtTipInserter := batch.NewBatchInserter(ctx, s.Store.Handle().DB(), "lsif_uploads_visible_at_tip", "repository_id", "upload_id")

	for _, uploadMeta := range visibleUploads[tipCommit] {
		if err := uploadsVisibleAtTipInserter.Insert(ctx, repositoryID, uploadMeta.UploadID); err != nil {
			return err
		}
	}
	if err := uploadsVisibleAtTipInserter.Flush(ctx); err != nil {
		return err
	}

	if dirtyToken != 0 {
		// If the user requests us to clear a dirty token, set the updated_token value to
		// the dirty token if it wouldn't decrease the value. Dirty repositories are determined
		// by having a non-equal dirty and update token, and we want the most recent upload
		// token to win this write.
		if err := tx.Store.Exec(ctx, sqlf.Sprintf(
			`UPDATE lsif_dirty_repositories SET update_token = GREATEST(update_token, %s) WHERE repository_id = %s`,
			dirtyToken,
			repositoryID,
		)); err != nil {
			return err
		}
	}

	return nil
}
