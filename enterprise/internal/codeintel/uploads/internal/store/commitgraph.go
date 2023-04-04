package store

import (
	"context"
	"strconv"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/commitgraph"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// SetRepositoryAsDirty marks the given repository's commit graph as out of date.
func (s *store) SetRepositoryAsDirty(ctx context.Context, repositoryID int) (err error) {
	ctx, _, endObservation := s.operations.setRepositoryAsDirty.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	return s.db.Exec(ctx, sqlf.Sprintf(setRepositoryAsDirtyQuery, repositoryID))
}

// GetDirtyRepositories returns list of repositories whose commit graph is out of date. The dirty token should be
// passed to CalculateVisibleUploads in order to unmark the repository.
func (s *store) GetDirtyRepositories(ctx context.Context) (_ []shared.DirtyRepository, err error) {
	ctx, trace, endObservation := s.operations.getDirtyRepositories.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	repositories, err := scanDirtyRepositories(s.db.Query(ctx, sqlf.Sprintf(dirtyRepositoriesQuery)))
	if err != nil {
		return nil, err
	}
	trace.AddEvent("TODO Domain Owner", attribute.Int("numRepositories", len(repositories)))

	return repositories, nil
}

const dirtyRepositoriesQuery = `
SELECT ldr.repository_id, repo.name, ldr.dirty_token
  FROM lsif_dirty_repositories ldr
    INNER JOIN repo ON repo.id = ldr.repository_id
  WHERE ldr.dirty_token > ldr.update_token
    AND repo.deleted_at IS NULL
	AND repo.blocked IS NULL
`

var scanDirtyRepositories = basestore.NewSliceScanner(func(s dbutil.Scanner) (dr shared.DirtyRepository, _ error) {
	err := s.Scan(&dr.RepositoryID, &dr.RepositoryName, &dr.DirtyToken)
	return dr, err
})

// UpdateUploadsVisibleToCommits uses the given commit graph and the tip of non-stale branches and tags to determine the
// set of LSIF uploads that are visible for each commit, and the set of uploads which are visible at the tip of a
// non-stale branch or tag. The decorated commit graph is serialized to Postgres for use by find closest dumps
// queries.
//
// If dirtyToken is supplied, the repository will be unmarked when the supplied token does matches the most recent
// token stored in the database, the flag will not be cleared as another request for update has come in since this
// token has been read.
func (s *store) UpdateUploadsVisibleToCommits(
	ctx context.Context,
	repositoryID int,
	commitGraph *gitdomain.CommitGraph,
	refDescriptions map[string][]gitdomain.RefDescription,
	maxAgeForNonStaleBranches time.Duration,
	maxAgeForNonStaleTags time.Duration,
	dirtyToken int,
	now time.Time,
) (err error) {
	ctx, trace, endObservation := s.operations.updateUploadsVisibleToCommits.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", repositoryID),
			log.Int("numCommitGraphKeys", len(commitGraph.Order())),
			log.Int("numRefDescriptions", len(refDescriptions)),
			log.Int("dirtyToken", dirtyToken),
		},
	})
	defer endObservation(1, observation.Args{})

	tx, err := s.db.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Determine the retention policy for this repository
	maxAgeForNonStaleBranches, maxAgeForNonStaleTags, err = refineRetentionConfiguration(ctx, tx, repositoryID, maxAgeForNonStaleBranches, maxAgeForNonStaleTags)
	if err != nil {
		return err
	}
	trace.AddEvent("TODO Domain Owner",
		attribute.String("maxAgeForNonStaleBranches", maxAgeForNonStaleBranches.String()),
		attribute.String("maxAgeForNonStaleTags", maxAgeForNonStaleTags.String()))

	// Pull all queryable upload metadata known to this repository so we can correlate
	// it with the current  commit graph.
	commitGraphView, err := scanCommitGraphView(tx.Query(ctx, sqlf.Sprintf(calculateVisibleUploadsCommitGraphQuery, repositoryID)))
	if err != nil {
		return err
	}
	trace.AddEvent("TODO Domain Owner",
		attribute.Int("numCommitGraphViewMetaKeys", len(commitGraphView.Meta)),
		attribute.Int("numCommitGraphViewTokenKeys", len(commitGraphView.Tokens)))

	// Determine which uploads are visible to which commits for this repository
	graph := commitgraph.NewGraph(commitGraph, commitGraphView)

	pctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Return a structure holding several channels that are populated by a background goroutine.
	// When we write this data to temporary tables, we have three consumers pulling values from
	// these channels in parallel. We need to make sure that once we return from this function that
	// the producer routine shuts down. This prevents the producer from leaking if there is an
	// error in one of the consumers before all values have been emitted.
	sanitizedInput := sanitizeCommitInput(pctx, graph, refDescriptions, maxAgeForNonStaleBranches, maxAgeForNonStaleTags)

	// Write the graph into temporary tables in Postgres
	if err := s.writeVisibleUploads(ctx, sanitizedInput, tx); err != nil {
		return err
	}

	// Persist data to permenant table: t_lsif_nearest_uploads -> lsif_nearest_uploads
	if err := s.persistNearestUploads(ctx, repositoryID, tx); err != nil {
		return err
	}

	// Persist data to permenant table: t_lsif_nearest_uploads_links -> lsif_nearest_uploads_links
	if err := s.persistNearestUploadsLinks(ctx, repositoryID, tx); err != nil {
		return err
	}

	// Persist data to permenant table: t_lsif_uploads_visible_at_tip -> lsif_uploads_visible_at_tip
	if err := s.persistUploadsVisibleAtTip(ctx, repositoryID, tx); err != nil {
		return err
	}

	if dirtyToken != 0 {
		// If the user requests us to clear a dirty token, set the updated_token value to
		// the dirty token if it wouldn't decrease the value. Dirty repositories are determined
		// by having a non-equal dirty and update token, and we want the most recent upload
		// token to win this write.
		nowTimestamp := sqlf.Sprintf("transaction_timestamp()")
		if !now.IsZero() {
			nowTimestamp = sqlf.Sprintf("%s", now)
		}
		if err := tx.Exec(ctx, sqlf.Sprintf(calculateVisibleUploadsDirtyRepositoryQuery, dirtyToken, nowTimestamp, repositoryID)); err != nil {
			return err
		}
	}

	// All completed uploads are now visible. Mark any uploads queued for deletion as deleted as
	// they are no longer reachable from the commit graph and cannot be used to fulfill any API
	// requests.
	unset, _ := tx.SetLocal(ctx, "codeintel.lsif_uploads_audit.reason", "upload not reachable within the commit graph")
	defer unset(ctx)
	if err := tx.Exec(ctx, sqlf.Sprintf(calculateVisibleUploadsDeleteUploadsQueuedForDeletionQuery, repositoryID)); err != nil {
		return err
	}

	return nil
}

const calculateVisibleUploadsCommitGraphQuery = `
SELECT id, commit, md5(root || ':' || indexer) as token, 0 as distance FROM lsif_uploads WHERE state = 'completed' AND repository_id = %s
`

const calculateVisibleUploadsDirtyRepositoryQuery = `
UPDATE lsif_dirty_repositories SET update_token = GREATEST(update_token, %s), updated_at = %s WHERE repository_id = %s
`

const calculateVisibleUploadsDeleteUploadsQueuedForDeletionQuery = `
WITH
candidates AS (
	SELECT u.id
	FROM lsif_uploads u
	WHERE u.state = 'deleting' AND u.repository_id = %s

	-- Lock these rows in a deterministic order so that we don't
	-- deadlock with other processes updating the lsif_uploads table.
	ORDER BY u.id FOR UPDATE
)
UPDATE lsif_uploads
SET state = 'deleted'
WHERE id IN (SELECT id FROM candidates)
`

// refineRetentionConfiguration returns the maximum age for no-stale branches and tags, effectively, as configured
// for the given repository. If there is no retention configuration for the given repository, the given default
// values are returned unchanged.
func refineRetentionConfiguration(ctx context.Context, store *basestore.Store, repositoryID int, maxAgeForNonStaleBranches, maxAgeForNonStaleTags time.Duration) (_, _ time.Duration, err error) {
	rows, err := store.Query(ctx, sqlf.Sprintf(retentionConfigurationQuery, repositoryID))
	if err != nil {
		return 0, 0, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		var v1, v2 int
		if err := rows.Scan(&v1, &v2); err != nil {
			return 0, 0, err
		}

		maxAgeForNonStaleBranches = time.Second * time.Duration(v1)
		maxAgeForNonStaleTags = time.Second * time.Duration(v2)
	}

	return maxAgeForNonStaleBranches, maxAgeForNonStaleTags, nil
}

const retentionConfigurationQuery = `
SELECT max_age_for_non_stale_branches_seconds, max_age_for_non_stale_tags_seconds
FROM lsif_retention_configuration
WHERE repository_id = %s
`

// GetCommitsVisibleToUpload returns the set of commits for which the given upload can answer code intelligence queries.
// To paginate, supply the token returned from this method to the invocation for the next page.
func (s *store) GetCommitsVisibleToUpload(ctx context.Context, uploadID, limit int, token *string) (_ []string, nextToken *string, err error) {
	ctx, _, endObservation := s.operations.getCommitsVisibleToUpload.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("uploadID", uploadID),
		log.Int("limit", limit),
	}})
	defer endObservation(1, observation.Args{})

	after := ""
	if token != nil {
		after = *token
	}

	commits, err := basestore.ScanStrings(s.db.Query(ctx, sqlf.Sprintf(commitsVisibleToUploadQuery, strconv.Itoa(uploadID), after, limit)))
	if err != nil {
		return nil, nil, err
	}

	if len(commits) > 0 {
		last := commits[len(commits)-1]
		nextToken = &last
	}

	return commits, nextToken, nil
}

const commitsVisibleToUploadQuery = `
WITH
direct_commits AS (
	SELECT nu.repository_id, nu.commit_bytea
	FROM lsif_nearest_uploads nu
	WHERE nu.uploads ? %s
),
linked_commits AS (
	SELECT ul.commit_bytea
	FROM direct_commits dc
	JOIN lsif_nearest_uploads_links ul
	ON
		ul.repository_id = dc.repository_id AND
		ul.ancestor_commit_bytea = dc.commit_bytea
),
combined_commits AS (
	SELECT dc.commit_bytea FROM direct_commits dc
	UNION ALL
	SELECT lc.commit_bytea FROM linked_commits lc
)
SELECT encode(c.commit_bytea, 'hex') as commit
FROM combined_commits c
WHERE decode(%s, 'hex') < c.commit_bytea
ORDER BY c.commit_bytea
LIMIT %s
`

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
//
// It is possible for some dumps to overlap theoretically, e.g. if someone uploads one dump covering the repository root and then later
// splits the repository into multiple dumps. For this reason, the returned dumps are always sorted in most-recently-finished order to
// prevent returning data from stale dumps.
func (s *store) FindClosestDumps(ctx context.Context, repositoryID int, commit, path string, rootMustEnclosePath bool, indexer string) (_ []shared.Dump, err error) {
	ctx, trace, endObservation := s.operations.findClosestDumps.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", repositoryID),
			log.String("commit", commit),
			log.String("path", path),
			log.Bool("rootMustEnclosePath", rootMustEnclosePath),
			log.String("indexer", indexer),
		},
	})
	defer endObservation(1, observation.Args{})

	conds := makeFindClosestDumpConditions(path, rootMustEnclosePath, indexer)
	query := sqlf.Sprintf(findClosestDumpsQuery, makeVisibleUploadsQuery(repositoryID, commit), sqlf.Join(conds, " AND "))

	dumps, err := scanDumps(s.db.Query(ctx, query))
	if err != nil {
		return nil, err
	}
	trace.AddEvent("TODO Domain Owner", attribute.Int("numDumps", len(dumps)))

	return dumps, nil
}

const findClosestDumpsQuery = `
WITH
visible_uploads AS (%s)
SELECT
	u.id,
	u.commit,
	u.root,
	EXISTS (` + visibleAtTipSubselectQuery + `) AS visible_at_tip,
	u.uploaded_at,
	u.state,
	u.failure_message,
	u.started_at,
	u.finished_at,
	u.process_after,
	u.num_resets,
	u.num_failures,
	u.repository_id,
	u.repository_name,
	u.indexer,
	u.indexer_version,
	u.associated_index_id
FROM visible_uploads vu
JOIN lsif_dumps_with_repository_name u ON u.id = vu.upload_id
WHERE %s
ORDER BY u.finished_at DESC
`

// FindClosestDumpsFromGraphFragment returns the set of dumps that can most accurately answer queries for the given repository, commit,
// path, and optional indexer by only considering the given fragment of the full git graph. See FindClosestDumps for additional details.
func (s *store) FindClosestDumpsFromGraphFragment(ctx context.Context, repositoryID int, commit, path string, rootMustEnclosePath bool, indexer string, commitGraph *gitdomain.CommitGraph) (_ []shared.Dump, err error) {
	ctx, trace, endObservation := s.operations.findClosestDumpsFromGraphFragment.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", repositoryID),
			log.String("commit", commit),
			log.String("path", path),
			log.Bool("rootMustEnclosePath", rootMustEnclosePath),
			log.String("indexer", indexer),
			log.Int("numCommitGraphKeys", len(commitGraph.Order())),
		},
	})
	defer endObservation(1, observation.Args{})

	if len(commitGraph.Order()) == 0 {
		return nil, nil
	}

	commits := make([]string, 0, len(commitGraph.Graph()))
	for commit := range commitGraph.Graph() {
		commits = append(commits, commit)
	}

	commitGraphView, err := scanCommitGraphView(s.db.Query(ctx, sqlf.Sprintf(
		findClosestDumpsFromGraphFragmentCommitGraphQuery,
		makeVisibleUploadCandidatesQuery(repositoryID, commits...)),
	))
	if err != nil {
		return nil, err
	}
	trace.AddEvent("TODO Domain Owner",
		attribute.Int("numCommitGraphViewMetaKeys", len(commitGraphView.Meta)),
		attribute.Int("numCommitGraphViewTokenKeys", len(commitGraphView.Tokens)))

	var ids []*sqlf.Query
	for _, uploadMeta := range commitgraph.NewGraph(commitGraph, commitGraphView).UploadsVisibleAtCommit(commit) {
		ids = append(ids, sqlf.Sprintf("%d", uploadMeta.UploadID))
	}
	if len(ids) == 0 {
		return nil, nil
	}

	conds := makeFindClosestDumpConditions(path, rootMustEnclosePath, indexer)
	query := sqlf.Sprintf(findClosestDumpsFromGraphFragmentQuery, sqlf.Join(ids, ","), sqlf.Join(conds, " AND "))

	dumps, err := scanDumps(s.db.Query(ctx, query))
	if err != nil {
		return nil, err
	}
	trace.AddEvent("TODO Domain Owner", attribute.Int("numDumps", len(dumps)))

	return dumps, nil
}

const findClosestDumpsFromGraphFragmentCommitGraphQuery = `
WITH
visible_uploads AS (%s)
SELECT
	vu.upload_id,
	encode(vu.commit_bytea, 'hex'),
	md5(u.root || ':' || u.indexer) as token,
	vu.distance
FROM visible_uploads vu
JOIN lsif_uploads u ON u.id = vu.upload_id
`

const findClosestDumpsFromGraphFragmentQuery = `
SELECT
	u.id,
	u.commit,
	u.root,
	EXISTS (` + visibleAtTipSubselectQuery + `) AS visible_at_tip,
	u.uploaded_at,
	u.state,
	u.failure_message,
	u.started_at,
	u.finished_at,
	u.process_after,
	u.num_resets,
	u.num_failures,
	u.repository_id,
	u.repository_name,
	u.indexer,
	u.indexer_version,
	u.associated_index_id
FROM lsif_dumps_with_repository_name u
WHERE u.id IN (%s) AND %s
`

// GetRepositoriesMaxStaleAge returns the longest duration that a repository has been (currently) stale for. This method considers
// only repositories that would be returned by DirtyRepositories. This method returns a duration of zero if there
// are no stale repositories.
func (s *store) GetRepositoriesMaxStaleAge(ctx context.Context) (_ time.Duration, err error) {
	ctx, _, endObservation := s.operations.getRepositoriesMaxStaleAge.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	ageSeconds, ok, err := basestore.ScanFirstInt(s.db.Query(ctx, sqlf.Sprintf(maxStaleAgeQuery)))
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, nil
	}

	return time.Duration(ageSeconds) * time.Second, nil
}

const maxStaleAgeQuery = `
SELECT EXTRACT(EPOCH FROM NOW() - ldr.set_dirty_at)::integer AS age
  FROM lsif_dirty_repositories ldr
    INNER JOIN repo ON repo.id = ldr.repository_id
  WHERE ldr.dirty_token > ldr.update_token
    AND repo.deleted_at IS NULL
    AND repo.blocked IS NULL
  ORDER BY age DESC
  LIMIT 1
`

// CommitGraphMetadata returns whether or not the commit graph for the given repository is stale, along with the date of
// the most recent commit graph refresh for the given repository.
func (s *store) GetCommitGraphMetadata(ctx context.Context, repositoryID int) (stale bool, updatedAt *time.Time, err error) {
	ctx, _, endObservation := s.operations.getCommitGraphMetadata.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	updateToken, dirtyToken, updatedAt, exists, err := scanCommitGraphMetadata(s.db.Query(ctx, sqlf.Sprintf(commitGraphQuery, repositoryID)))
	if err != nil {
		return false, nil, err
	}
	if !exists {
		return false, nil, nil
	}

	return updateToken != dirtyToken, updatedAt, err
}

const commitGraphQuery = `
SELECT update_token, dirty_token, updated_at FROM lsif_dirty_repositories WHERE repository_id = %s LIMIT 1
`
