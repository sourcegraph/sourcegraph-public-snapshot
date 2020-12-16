package dbstore

import (
	"bytes"
	"context"
	"database/sql"
	"strconv"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/commitgraph"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
	"github.com/sourcegraph/sourcegraph/internal/db/batch"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// scanCommitGraphView scans a commit graph view from the return value of `*Store.query`.
func scanCommitGraphView(rows *sql.Rows, queryErr error) (_ *commitgraph.CommitGraphView, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	commitGraphView := commitgraph.NewCommitGraphView()

	for rows.Next() {
		var meta commitgraph.UploadMeta
		var commit, token string

		if err := rows.Scan(&meta.UploadID, &commit, &token, &meta.Distance); err != nil {
			return nil, err
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

	count, _, err := basestore.ScanFirstInt(s.Store.Query(
		ctx,
		sqlf.Sprintf(`
			SELECT
				(SELECT COUNT(*) FROM lsif_nearest_uploads WHERE repository_id = %s AND commit_bytea = %s) +
				(SELECT COUNT(*) FROM lsif_nearest_uploads_links WHERE repository_id = %s AND commit_bytea = %s)
		`,
			repositoryID, dbutil.CommitBytea(commit),
			repositoryID, dbutil.CommitBytea(commit)),
	))

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

// CalculateVisibleUploads uses the given commit graph and the tip commit of the default branch to determine the
// set of LSIF uploads that are visible for each commit, and the set of uploads which are visible at the tip. The
// decorated commit graph is serialized to Postgres for use by find closest dumps queries.
//
// If dirtyToken is supplied, the repository will be unmarked when the supplied token does matches the most recent
// token stored in the database, the flag will not be cleared as another request for update has come in since this
// token has been read.
func (s *Store) CalculateVisibleUploads(ctx context.Context, repositoryID int, commitGraph *gitserver.CommitGraph, tipCommit string, dirtyToken int) (err error) {
	ctx, traceLog, endObservation := s.operations.calculateVisibleUploads.WithAndLogger(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", repositoryID),
			log.Int("numCommitGraphKeys", len(commitGraph.Order())),
			log.String("tipCommit", tipCommit),
			log.Int("dirtyToken", dirtyToken),
		},
	})
	defer endObservation(1, observation.Args{})

	tx, err := s.transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Pull all queryable upload metadata known to this repository so we can correlate
	// it with the current  commit graph.
	commitGraphView, err := scanCommitGraphView(tx.Store.Query(ctx, sqlf.Sprintf(`
		SELECT id, commit, md5(root || ':' || indexer) as token, 0 as distance
		FROM lsif_uploads
		WHERE state = 'completed' AND repository_id = %s
	`, repositoryID)))
	if err != nil {
		return err
	}
	traceLog(
		log.Int("numCommitGraphViewMetaKeys", len(commitGraphView.Meta)),
		log.Int("numCommitGraphViewTokenKeys", len(commitGraphView.Tokens)),
	)

	// Determine which uploads are visible to which commits for this repository
	graph := commitgraph.NewGraph(commitGraph, commitGraphView)

	// Clear all old visibility data for this repository
	for _, query := range []string{
		`DELETE FROM lsif_nearest_uploads WHERE repository_id = %s`,
		`DELETE FROM lsif_nearest_uploads_links WHERE repository_id = %s`,
		`DELETE FROM lsif_uploads_visible_at_tip WHERE repository_id = %s`,
	} {
		if err := tx.Store.Exec(ctx, sqlf.Sprintf(query, repositoryID)); err != nil {
			return err
		}
	}

	// Update the set of uploads that are visible from each commit for a given repository.
	nearestUploadsInserter := batch.NewBatchInserter(
		ctx,
		s.Store.Handle().DB(),
		"lsif_nearest_uploads",
		"repository_id",
		"commit_bytea",
		"uploads",
	)

	// Update the commits not inserted into the table above by adding links to a unique
	// ancestor and their relative distance in the graph. We use this as a cheap way to
	// reconstruct the full data set, which is multiplicative in the size of the commit
	// graph AND the number of unique roots.
	nearestUploadsLinksInserter := batch.NewBatchInserter(
		ctx,
		s.Store.Handle().DB(),
		"lsif_nearest_uploads_links",
		"repository_id",
		"commit_bytea",
		"ancestor_commit_bytea",
		"distance",
	)

	listSerializer := NewUploadMetaListSerializer()

	var numNearestUploadsRecords int
	var numNearestUploadsLinksRecords int

	for v := range graph.Stream() {
		if v.Uploads != nil {
			numNearestUploadsRecords++

			if err := nearestUploadsInserter.Insert(
				ctx,
				repositoryID,
				dbutil.CommitBytea(v.Uploads.Commit),
				listSerializer.Serialize(v.Uploads.Uploads),
			); err != nil {
				return err
			}
		}
		if v.Links != nil {
			numNearestUploadsLinksRecords++

			if err := nearestUploadsLinksInserter.Insert(
				ctx,
				repositoryID,
				dbutil.CommitBytea(v.Links.Commit),
				dbutil.CommitBytea(v.Links.AncestorCommit),
				v.Links.Distance,
			); err != nil {
				return err
			}
		}
	}
	if err := nearestUploadsInserter.Flush(ctx); err != nil {
		return err
	}
	if err := nearestUploadsLinksInserter.Flush(ctx); err != nil {
		return err
	}
	traceLog(
		log.Int("numNearestUploadsRecords", numNearestUploadsRecords),
		log.Int("numNearestUploadsLinksRecords", numNearestUploadsLinksRecords),
	)

	// Update which repositories are visible from the tip of the default branch. This
	// flag is used to determine which bundles for a repository we open during a global
	// find references query.
	uploadsVisibleAtTipInserter := batch.NewBatchInserter(
		ctx,
		s.Store.Handle().DB(),
		"lsif_uploads_visible_at_tip",
		"repository_id",
		"upload_id",
	)

	var numUploadsVisibleAtTipRecords int
	for _, uploadMeta := range graph.UploadsVisibleAtCommit(tipCommit) {
		numUploadsVisibleAtTipRecords++

		if err := uploadsVisibleAtTipInserter.Insert(ctx, repositoryID, uploadMeta.UploadID); err != nil {
			return err
		}
	}
	if err := uploadsVisibleAtTipInserter.Flush(ctx); err != nil {
		return err
	}
	traceLog(log.Int("numUploadsVisibleAtTipRecords", numUploadsVisibleAtTipRecords))

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

type uploadMetaListSerializer struct {
	buf     bytes.Buffer
	scratch []byte
}

func NewUploadMetaListSerializer() *uploadMetaListSerializer {
	return &uploadMetaListSerializer{
		scratch: make([]byte, 4),
	}
}

// Serialize returns a new byte slice with the given upload metadata values encoded
// as a JSON object (keys being the upload_id and values being the distance field).
//
// Our original attempt just built a map[int]int and passed it to the JSON package
// to be marshalled into a byte array. Unfortunately that puts reflection over the
// map value in the hot path for commit graph processing. We also can't avoid the
// reflection by passing a struct without changing the shape of the data persisted
// in the database.
//
// By serializing this value ourselves we minimize allocations. This change resulted
// in a 50% reduction of the memory required by BenchmarkCalculateVisibleUploads.
//
// This method is not safe for concurrent use.
func (s *uploadMetaListSerializer) Serialize(uploadMetas []commitgraph.UploadMeta) []byte {
	s.write(uploadMetas)
	return s.take()
}

func (s *uploadMetaListSerializer) write(uploadMetas []commitgraph.UploadMeta) {
	s.buf.WriteByte('{')
	for i, uploadMeta := range uploadMetas {
		if i > 0 {
			s.buf.WriteByte(',')
		}

		s.writeUploadMeta(uploadMeta)
	}
	s.buf.WriteByte('}')
}

func (s *uploadMetaListSerializer) writeUploadMeta(uploadMeta commitgraph.UploadMeta) {
	s.buf.WriteByte('"')
	s.writeInteger(uploadMeta.UploadID)
	s.buf.Write([]byte{'"', ':'})
	s.writeInteger(int(uploadMeta.Distance))
}

func (s *uploadMetaListSerializer) writeInteger(value int) {
	s.scratch = s.scratch[:0]
	s.scratch = strconv.AppendInt(s.scratch, int64(value), 10)
	s.buf.Write(s.scratch)
}

func (s *uploadMetaListSerializer) take() []byte {
	dest := make([]byte, s.buf.Len())
	copy(dest, s.buf.Bytes())
	s.buf.Reset()

	return dest
}
