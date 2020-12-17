package dbstore

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/commitgraph"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Dump is a subset of the lsif_uploads table (queried via the lsif_dumps_with_repository_name view)
// and stores only processed records.
type Dump struct {
	ID             int        `json:"id"`
	Commit         string     `json:"commit"`
	Root           string     `json:"root"`
	VisibleAtTip   bool       `json:"visibleAtTip"`
	UploadedAt     time.Time  `json:"uploadedAt"`
	State          string     `json:"state"`
	FailureMessage *string    `json:"failureMessage"`
	StartedAt      *time.Time `json:"startedAt"`
	FinishedAt     *time.Time `json:"finishedAt"`
	ProcessAfter   *time.Time `json:"processAfter"`
	NumResets      int        `json:"numResets"`
	NumFailures    int        `json:"numFailures"`
	RepositoryID   int        `json:"repositoryId"`
	RepositoryName string     `json:"repositoryName"`
	Indexer        string     `json:"indexer"`
}

// scanDumps scans a slice of dumps from the return value of `*Store.query`.
func scanDumps(rows *sql.Rows, queryErr error) (_ []Dump, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var dumps []Dump
	for rows.Next() {
		var dump Dump
		if err := rows.Scan(
			&dump.ID,
			&dump.Commit,
			&dump.Root,
			&dump.VisibleAtTip,
			&dump.UploadedAt,
			&dump.State,
			&dump.FailureMessage,
			&dump.StartedAt,
			&dump.FinishedAt,
			&dump.ProcessAfter,
			&dump.NumResets,
			&dump.NumFailures,
			&dump.RepositoryID,
			&dump.RepositoryName,
			&dump.Indexer,
		); err != nil {
			return nil, err
		}

		dumps = append(dumps, dump)
	}

	return dumps, nil
}

// scanFirstDump scans a slice of dumps from the return value of `*Store.query` and returns the first.
func scanFirstDump(rows *sql.Rows, err error) (Dump, bool, error) {
	dumps, err := scanDumps(rows, err)
	if err != nil || len(dumps) == 0 {
		return Dump{}, false, err
	}
	return dumps[0], true, nil
}

// GetDumpByID returns a dump by its identifier and boolean flag indicating its existence.
func (s *Store) GetDumpByID(ctx context.Context, id int) (_ Dump, _ bool, err error) {
	ctx, endObservation := s.operations.getDumpByID.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	return scanFirstDump(s.Store.Query(ctx, sqlf.Sprintf(`
		SELECT
			d.id,
			d.commit,
			d.root,
			`+visibleAtTipFragment+` AS visible_at_tip,
			d.uploaded_at,
			d.state,
			d.failure_message,
			d.started_at,
			d.finished_at,
			d.process_after,
			d.num_resets,
			d.num_failures,
			d.repository_id,
			d.repository_name,
			d.indexer
		FROM lsif_dumps_with_repository_name d WHERE d.id = %s
	`, id)))
}

const visibleAtTipFragment = `EXISTS (SELECT 1 FROM lsif_uploads_visible_at_tip WHERE repository_id = d.repository_id AND upload_id = d.id)`

// FindClosestDumps returns the set of dumps that can most accurately answer queries for the given repository, commit, path, and
// optional indexer. If rootMustEnclosePath is true, then only dumps with a root which is a prefix of path are returned. Otherwise,
// any dump with a root intersecting the given path is returned.
//
// This method should be used when the commit is known to exist in the lsif_nearest_uploads table. If it doesn't, then this method
// will return no dumps (as the input commit is not reachable from anything with an upload). The nearest uploads table must be
// refreshed before calling this method when the commit is unknown.
//
// Because refreshing the commit graph can be very expensive, we also provide FindClosestDumpsFromGraphFragment. That method should
// be used instead in low-latency paths. It should be supplied a small fragment of the commit graph that contains the input commit
// as well as a commit that is likely to exist in the lsif_nearest_uploads table. This is enough to propagate the correct upload
// visibility data down the graph fragment.
//
// The graph supplied to FindClosestDumpsFromGraphFragment will also determine whether or not it is possible to produce a partial set
// of visible uploads (ideally, we'd like to return the complete set of visible uploads, or fail). If the graph fragment is complete
// by depth (e.g. if the graph contains an ancestor at depth d, then the graph also contains all other ancestors up to depth d), then
// we get the ideal behavior. Only if we contain a partial row of ancestors will we return partial results.
func (s *Store) FindClosestDumps(ctx context.Context, repositoryID int, commit, path string, rootMustEnclosePath bool, indexer string) (dumps []Dump, err error) {
	ctx, endObservation := s.operations.findClosestDumps.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", repositoryID),
			log.String("commit", commit),
			log.String("path", path),
			log.Bool("rootMustEnclosePath", rootMustEnclosePath),
			log.String("indexer", indexer),
		},
	})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("numDumps", len(dumps)),
		}})
	}()

	conds := makeFindClosestDumpConditions(path, rootMustEnclosePath, indexer)

	return scanDumps(s.Store.Query(
		ctx,
		sqlf.Sprintf(`
			WITH visible_uploads AS (%s)
			SELECT
				d.id,
				d.commit,
				d.root,
				`+visibleAtTipFragment+` AS visible_at_tip,
				d.uploaded_at,
				d.state,
				d.failure_message,
				d.started_at,
				d.finished_at,
				d.process_after,
				d.num_resets,
				d.num_failures,
				d.repository_id,
				d.repository_name,
				d.indexer
			FROM visible_uploads vu
			JOIN lsif_dumps_with_repository_name d ON d.id = vu.upload_id
			WHERE %s
		`, makeVisibleUploadsQuery(repositoryID, commit), sqlf.Join(conds, " AND ")),
	))
}

// FindClosestDumpsFromGraphFragment returns the set of dumps that can most accurately answer queries for the given repository, commit,
// path, and optional indexer by only considering the given fragment of the full git graph. See FindClosestDumps for additional details.
func (s *Store) FindClosestDumpsFromGraphFragment(ctx context.Context, repositoryID int, commit, path string, rootMustEnclosePath bool, indexer string, commitGraph *gitserver.CommitGraph) (dumps []Dump, err error) {
	ctx, traceLog, endObservation := s.operations.findClosestDumpsFromGraphFragment.WithAndLogger(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", repositoryID),
			log.String("commit", commit),
			log.String("path", path),
			log.Bool("rootMustEnclosePath", rootMustEnclosePath),
			log.String("indexer", indexer),
			log.Int("numCommitGraphKeys", len(commitGraph.Order())),
		},
	})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("numDumps", len(dumps)),
		}})
	}()

	if len(commitGraph.Order()) == 0 {
		return nil, nil
	}

	commits := make([]string, 0, len(commitGraph.Graph()))
	for commit := range commitGraph.Graph() {
		commits = append(commits, commit)
	}

	commitGraphView, err := scanCommitGraphView(s.Store.Query(ctx, sqlf.Sprintf(`
		WITH visible_uploads AS (%s)
		SELECT
			vu.upload_id,
			encode(vu.commit_bytea, 'hex'),
			md5(u.root || ':' || u.indexer) as token,
			vu.distance
		FROM visible_uploads vu
		JOIN lsif_uploads u ON u.id = vu.upload_id
	`, makeVisibleUploadCandidatesQuery(repositoryID, commits...))))
	if err != nil {
		return nil, err
	}
	traceLog(
		log.Int("numCommitGraphViewMetaKeys", len(commitGraphView.Meta)),
		log.Int("numCommitGraphViewTokenKeys", len(commitGraphView.Tokens)),
	)

	var ids []*sqlf.Query
	for _, uploadMeta := range commitgraph.NewGraph(commitGraph, commitGraphView).UploadsVisibleAtCommit(commit) {
		ids = append(ids, sqlf.Sprintf("%d", uploadMeta.UploadID))
	}
	if len(ids) == 0 {
		return nil, nil
	}

	conds := makeFindClosestDumpConditions(path, rootMustEnclosePath, indexer)

	return scanDumps(s.Store.Query(
		ctx,
		sqlf.Sprintf(`
			SELECT
				d.id,
				d.commit,
				d.root,
				`+visibleAtTipFragment+` AS visible_at_tip,
				d.uploaded_at,
				d.state,
				d.failure_message,
				d.started_at,
				d.finished_at,
				d.process_after,
				d.num_resets,
				d.num_failures,
				d.repository_id,
				d.repository_name,
				d.indexer
			FROM lsif_dumps_with_repository_name d
			WHERE d.id IN (%s) AND %s
		`, sqlf.Join(ids, ","), sqlf.Join(conds, " AND ")),
	))
}

// makeVisibleUploadCandidatesQuery returns a SQL query returning the set of uploads
// visible from the given commits. This is done by looking at each commit's row in the
// lsif_nearest_uploads, and the (adjusted) set of uploads visible from each commit's
// nearest ancestor according to data compressed in the links table.
//
// NB: A commit should be present in at most one of these tables.
func makeVisibleUploadCandidatesQuery(repositoryID int, commits ...string) *sqlf.Query {
	if len(commits) == 0 {
		panic("No commits supplied to makeVisibleUploadCandidatesQuery.")
	}

	commitQueries := make([]*sqlf.Query, 0, len(commits))
	for _, commit := range commits {
		commitQueries = append(commitQueries, sqlf.Sprintf("%s", dbutil.CommitBytea(commit)))
	}

	return sqlf.Sprintf(`
		SELECT
			nu.repository_id,
			upload_id::integer,
			nu.commit_bytea,
			u_distance::text::integer as distance
		FROM lsif_nearest_uploads nu
		CROSS JOIN jsonb_each(nu.uploads) as u(upload_id, u_distance)
		WHERE nu.repository_id = %s AND nu.commit_bytea IN (%s)
		UNION (
			SELECT
				nu.repository_id,
				upload_id::integer,
				ul.commit_bytea,
				u_distance::text::integer + ul.distance as distance
			FROM lsif_nearest_uploads_links ul
			JOIN lsif_nearest_uploads nu ON nu.repository_id = ul.repository_id AND nu.commit_bytea = ul.ancestor_commit_bytea
			CROSS JOIN jsonb_each(nu.uploads) as u(upload_id, u_distance)
			WHERE nu.repository_id = %s AND ul.commit_bytea IN (%s)
		)
	`,
		repositoryID,
		sqlf.Join(commitQueries, ", "),
		repositoryID,
		sqlf.Join(commitQueries, ", "),
	)
}

// makeVisibleUploadsQuery returns a SQL query returning the set of identifiers of uploads
// visible from the given commit. This is done by removing the "shadowed" values created
// by looking at a commit and it's ancestors visible commits.
func makeVisibleUploadsQuery(repositoryID int, commit string) *sqlf.Query {
	return sqlf.Sprintf(`
		SELECT
			t.upload_id
		FROM (
			SELECT
				t.*,
				row_number() OVER (PARTITION BY root, indexer ORDER BY distance) AS r
			FROM (%s) t
			JOIN lsif_uploads u ON u.id = upload_id
		) t
		WHERE t.r <= 1
	`, makeVisibleUploadCandidatesQuery(repositoryID, commit))
}

func makeFindClosestDumpConditions(path string, rootMustEnclosePath bool, indexer string) (conds []*sqlf.Query) {
	if rootMustEnclosePath {
		// Ensure that the root is a prefix of the path
		conds = append(conds, sqlf.Sprintf(`%s LIKE (d.root || '%%%%')`, path))
	} else {
		// Ensure that the root is a prefix of the path or vice versa
		conds = append(conds, sqlf.Sprintf(`(%s LIKE (d.root || '%%%%') OR d.root LIKE (%s || '%%%%'))`, path, path))
	}
	if indexer != "" {
		conds = append(conds, sqlf.Sprintf("indexer = %s", indexer))
	}

	return conds
}

// SoftDeleteOldDumps marks dumps older than the given age that are not visible at the tip of the default branch
// as deleted. The associated repositories will be marked as dirty so that their commit graphs are updated in the
// background.
func (s *Store) SoftDeleteOldDumps(ctx context.Context, maxAge time.Duration, now time.Time) (count int, err error) {
	ctx, endObservation := s.operations.softDeleteOldDumps.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("maxAge", maxAge.String()),
	}})
	defer endObservation(1, observation.Args{})

	tx, err := s.transact(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { err = tx.Done(err) }()

	repositoryIDs, err := scanCounts(tx.Store.Query(ctx, sqlf.Sprintf(`
		WITH u AS (
			UPDATE lsif_uploads u
				SET state = 'deleted'
				WHERE
					%s - u.finished_at > (%s || ' second')::interval AND
					u.id NOT IN (SELECT uv.upload_id FROM lsif_uploads_visible_at_tip uv WHERE uv.repository_id = u.repository_id)
				RETURNING id, repository_id
		)
		SELECT u.repository_id, count(*) FROM u GROUP BY u.repository_id
	`, now, maxAge/time.Second)))
	if err != nil {
		return 0, err
	}

	for repositoryID, numUpdated := range repositoryIDs {
		if err := tx.MarkRepositoryAsDirty(ctx, repositoryID); err != nil {
			return 0, err
		}

		count += numUpdated
	}

	return count, nil
}

// DeleteOverlapapingDumps deletes all completed uploads for the given repository with the same
// commit, root, and indexer. This is necessary to perform during conversions before changing
// the state of a processing upload to completed as there is a unique index on these four columns.
func (s *Store) DeleteOverlappingDumps(ctx context.Context, repositoryID int, commit, root, indexer string) (err error) {
	ctx, endObservation := s.operations.deleteOverlappingDumps.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("commit", commit),
		log.String("root", root),
		log.String("indexer", indexer),
	}})
	defer endObservation(1, observation.Args{})

	return s.Store.Exec(ctx, sqlf.Sprintf(`
		UPDATE lsif_uploads
		SET state = 'deleted'
		WHERE repository_id = %s AND commit = %s AND root = %s AND indexer = %s AND state = 'completed'
	`, repositoryID, commit, root, indexer))
}
