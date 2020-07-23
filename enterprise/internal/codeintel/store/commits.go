package store

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
)

// scanUploadMeta scans upload metadata grouped by commit from the return value of `*store.query`.
func scanUploadMeta(rows *sql.Rows, queryErr error) (_ map[string][]UploadMeta, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = closeRows(rows, err) }()

	uploadMeta := map[string][]UploadMeta{}
	for rows.Next() {
		var uploadID int
		var commit string
		var root string
		var indexer string
		if err := rows.Scan(&uploadID, &commit, &root, &indexer); err != nil {
			return nil, err
		}

		uploadMeta[commit] = append(uploadMeta[commit], UploadMeta{
			UploadID: uploadID,
			Root:     root,
			Indexer:  indexer,
		})
	}

	return uploadMeta, nil
}

// HasRepository determines if there is LSIF data for the given repository.
func (s *store) HasRepository(ctx context.Context, repositoryID int) (bool, error) {
	count, _, err := scanFirstInt(s.query(ctx, sqlf.Sprintf(`
		SELECT COUNT(*)
		FROM lsif_uploads
		WHERE repository_id = %s
		LIMIT 1
	`, repositoryID)))

	return count > 0, err
}

// HasCommit determines if the given commit is known for the given repository.
func (s *store) HasCommit(ctx context.Context, repositoryID int, commit string) (bool, error) {
	count, _, err := scanFirstInt(s.query(ctx, sqlf.Sprintf(`
		SELECT COUNT(*)
		FROM lsif_nearest_uploads
		WHERE repository_id = %s and commit = %s
		LIMIT 1
	`, repositoryID, commit)))

	return count > 0, err
}

// MarkRepositoryAsDirty marks the given repository's commit graph as out of date.
func (s *store) MarkRepositoryAsDirty(ctx context.Context, repositoryID int) error {
	return s.queryForEffect(
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
	defer func() { err = closeRows(rows, err) }()

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
func (s *store) DirtyRepositories(ctx context.Context) (map[int]int, error) {
	return scanIntPairs(s.query(ctx, sqlf.Sprintf(`SELECT repository_id, dirty_token FROM lsif_dirty_repositories WHERE dirty_token > update_token`)))
}

// CalculateVisibleUploads uses the given commit graph and the tip commit of the default branch to determine the set
// of LSIF uploads that are visible for each commit, and the set of uploads which are visible at the tip. The decorated
// commit graph is serialized to Postgres for use by find closest dumps queries.
//
// If dirtyToken is supplied, the repository will be unmarked when the supplied token does matches the most recent
// token stored in the database, the flag will not be cleared as another request for update has come in since this
// token has been read.
func (s *store) CalculateVisibleUploads(ctx context.Context, repositoryID int, graph map[string][]string, tipCommit string, dirtyToken int) error {
	tx, err := s.transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Pull all queryable upload metadata known to this repository so we can correlate
	// it with the current  commit graph.
	uploadMeta, err := scanUploadMeta(tx.query(ctx, sqlf.Sprintf(`
		SELECT id, commit, root, indexer
		FROM lsif_uploads
		WHERE state = 'completed' AND repository_id = %s
	`, repositoryID)))
	if err != nil {
		return err
	}

	// Determine which uploads are visible to which commits for this repository
	visibleUploads, err := calculateVisibleUploads(graph, uploadMeta)
	if err != nil {
		return err
	}

	// Clear all old visibility data for this repository
	for _, query := range []string{
		`DELETE FROM lsif_nearest_uploads WHERE repository_id = %s`,
		`DELETE FROM lsif_uploads_visible_at_tip WHERE repository_id = %s`,
	} {
		if err := tx.queryForEffect(ctx, sqlf.Sprintf(query, repositoryID)); err != nil {
			return err
		}
	}

	n := 0
	for _, uploads := range visibleUploads {
		n += len(uploads)
	}
	nearestUploadsRows := make([]*sqlf.Query, 0, n)

	for commit, uploads := range visibleUploads {
		for _, uploadMeta := range uploads {
			nearestUploadsRows = append(nearestUploadsRows, sqlf.Sprintf(
				"(%s, %s, %s, %s)",
				repositoryID,
				commit,
				uploadMeta.UploadID,
				uploadMeta.Distance,
			))
		}
	}

	// Insert new data for this repository in batches - it's likely we'll exceed the maximum
	// number of placeholders per query so we need to break it into several queries below this
	// size.
	for _, batch := range batchQueries(nearestUploadsRows, MaxPostgresNumParameters/4) {
		if err := tx.queryForEffect(ctx, sqlf.Sprintf(
			`INSERT INTO lsif_nearest_uploads (repository_id, "commit", upload_id, distance) VALUES %s`,
			sqlf.Join(batch, ","),
		)); err != nil {
			return err
		}
	}

	visibleAtTipRows := make([]*sqlf.Query, 0, len(visibleUploads[tipCommit]))
	for _, uploadMeta := range visibleUploads[tipCommit] {
		visibleAtTipRows = append(visibleAtTipRows, sqlf.Sprintf("(%s, %s)", repositoryID, uploadMeta.UploadID))
	}

	// Update which repositories are visible from the tip of the default branch. This
	// flag is used to determine which bundles for a repository we open during a global
	// find references query.
	if len(visibleAtTipRows) > 0 {
		for _, batch := range batchQueries(visibleAtTipRows, MaxPostgresNumParameters/2) {
			if err := tx.queryForEffect(ctx, sqlf.Sprintf(
				`INSERT INTO lsif_uploads_visible_at_tip (repository_id, upload_id) VALUES %s`,
				sqlf.Join(batch, ","),
			)); err != nil {
				return err
			}
		}
	}

	if dirtyToken != 0 {
		// If the user requests us to clear a dirty token, set the updated_token value to
		// the dirty token if it wouldn't decrease the value. Dirty repositories are determined
		// by having a non-equal dirty and update token, and we want the most recent upload
		// token to win this write.
		if err := tx.queryForEffect(ctx, sqlf.Sprintf(
			`UPDATE lsif_dirty_repositories SET update_token = GREATEST(update_token, %s) WHERE repository_id = %s`,
			dirtyToken,
			repositoryID,
		)); err != nil {
			return err
		}
	}

	return nil
}

// MaxPostgresNumParameters is the maximum number of parameters per query that Postgres
// will allow. Exceeding this number of parameters will cause the query to be rejected
// by the server.
const MaxPostgresNumParameters = 65535

// batchQueries cuts the given query slice into batches of a maximum size. This function
// will allocate only the outer array to hold each batch, and the data for each batch
// will refer to the given slice.
func batchQueries(queries []*sqlf.Query, batchSize int) (batches [][]*sqlf.Query) {
	for len(queries) > 0 {
		if len(queries) > batchSize {
			batches = append(batches, queries[:batchSize])
			queries = queries[batchSize:]
		} else {
			batches = append(batches, queries)
			queries = nil
		}
	}

	return batches
}
