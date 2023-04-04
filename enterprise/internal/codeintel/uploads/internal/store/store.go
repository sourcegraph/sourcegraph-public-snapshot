package store

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgtype"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	logger "github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/commitgraph"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

type Store interface {
	Transact(ctx context.Context) (Store, error)
	Done(err error) error

	// Upload records
	GetUploads(ctx context.Context, opts shared.GetUploadsOptions) ([]shared.Upload, int, error)
	GetUploadByID(ctx context.Context, id int) (shared.Upload, bool, error)
	GetDumpsByIDs(ctx context.Context, ids []int) ([]shared.Dump, error)
	GetUploadsByIDs(ctx context.Context, ids ...int) ([]shared.Upload, error)
	GetUploadsByIDsAllowDeleted(ctx context.Context, ids ...int) ([]shared.Upload, error)
	GetUploadIDsWithReferences(ctx context.Context, orderedMonikers []precise.QualifiedMonikerData, ignoreIDs []int, repositoryID int, commit string, limit int, offset int, trace observation.TraceLogger) ([]int, int, int, error)
	GetVisibleUploadsMatchingMonikers(ctx context.Context, repositoryID int, commit string, orderedMonikers []precise.QualifiedMonikerData, limit, offset int) (shared.PackageReferenceScanner, int, error)
	GetDumpsWithDefinitionsForMonikers(ctx context.Context, monikers []precise.QualifiedMonikerData) ([]shared.Dump, error)
	GetAuditLogsForUpload(ctx context.Context, uploadID int) ([]shared.UploadLog, error)
	DeleteUploads(ctx context.Context, opts shared.DeleteUploadsOptions) error
	DeleteUploadByID(ctx context.Context, id int) (bool, error)
	ReindexUploads(ctx context.Context, opts shared.ReindexUploadsOptions) error
	ReindexUploadByID(ctx context.Context, id int) error

	// Index records
	GetIndexes(ctx context.Context, opts shared.GetIndexesOptions) ([]uploadsshared.Index, int, error)
	GetIndexByID(ctx context.Context, id int) (uploadsshared.Index, bool, error)
	GetIndexesByIDs(ctx context.Context, ids ...int) ([]uploadsshared.Index, error)
	DeleteIndexByID(ctx context.Context, id int) (bool, error)
	DeleteIndexes(ctx context.Context, opts shared.DeleteIndexesOptions) error
	ReindexIndexByID(ctx context.Context, id int) error
	ReindexIndexes(ctx context.Context, opts shared.ReindexIndexesOptions) error

	// Upload record insertion + processing
	InsertUpload(ctx context.Context, upload shared.Upload) (int, error)
	AddUploadPart(ctx context.Context, uploadID, partIndex int) error
	MarkQueued(ctx context.Context, id int, uploadSize *int64) error
	MarkFailed(ctx context.Context, id int, reason string) error
	DeleteOverlappingDumps(ctx context.Context, repositoryID int, commit, root, indexer string) error
	WorkerutilStore(observationCtx *observation.Context) dbworkerstore.Store[shared.Upload]

	// Dependencies
	ReferencesForUpload(ctx context.Context, uploadID int) (shared.PackageReferenceScanner, error)
	UpdatePackages(ctx context.Context, dumpID int, packages []precise.Package) error
	UpdatePackageReferences(ctx context.Context, dumpID int, references []precise.PackageReference) error

	// Summary
	GetIndexers(ctx context.Context, opts shared.GetIndexersOptions) ([]string, error)
	GetRecentUploadsSummary(ctx context.Context, repositoryID int) ([]shared.UploadsWithRepositoryNamespace, error)
	GetRecentIndexesSummary(ctx context.Context, repositoryID int) ([]uploadsshared.IndexesWithRepositoryNamespace, error)
	RepositoryIDsWithErrors(ctx context.Context, offset, limit int) ([]uploadsshared.RepositoryWithCount, int, error)
	NumRepositoriesWithCodeIntelligence(ctx context.Context) (int, error)

	// Commit graph
	SetRepositoryAsDirty(ctx context.Context, repositoryID int) error
	GetDirtyRepositories(ctx context.Context) ([]shared.DirtyRepository, error)
	UpdateUploadsVisibleToCommits(ctx context.Context, repositoryID int, graph *gitdomain.CommitGraph, refDescriptions map[string][]gitdomain.RefDescription, maxAgeForNonStaleBranches, maxAgeForNonStaleTags time.Duration, dirtyToken int, now time.Time) error
	GetCommitsVisibleToUpload(ctx context.Context, uploadID, limit int, token *string) ([]string, *string, error)
	FindClosestDumps(ctx context.Context, repositoryID int, commit, path string, rootMustEnclosePath bool, indexer string) ([]shared.Dump, error)
	FindClosestDumpsFromGraphFragment(ctx context.Context, repositoryID int, commit, path string, rootMustEnclosePath bool, indexer string, commitGraph *gitdomain.CommitGraph) ([]shared.Dump, error)
	GetRepositoriesMaxStaleAge(ctx context.Context) (time.Duration, error)
	GetCommitGraphMetadata(ctx context.Context, repositoryID int) (stale bool, updatedAt *time.Time, _ error)

	// Expiration
	GetLastUploadRetentionScanForRepository(ctx context.Context, repositoryID int) (*time.Time, error)
	SetRepositoriesForRetentionScan(ctx context.Context, processDelay time.Duration, limit int) ([]int, error)
	UpdateUploadRetention(ctx context.Context, protectedIDs, expiredIDs []int) error
	SoftDeleteExpiredUploads(ctx context.Context, batchSize int) (int, int, error)
	SoftDeleteExpiredUploadsViaTraversal(ctx context.Context, maxTraversal int) (int, int, error)

	// Commit date
	GetOldestCommitDate(ctx context.Context, repositoryID int) (time.Time, bool, error)
	UpdateCommittedAt(ctx context.Context, repositoryID int, commit, commitDateString string) error
	SourcedCommitsWithoutCommittedAt(ctx context.Context, batchSize int) ([]SourcedCommits, error)

	// TOdO
	// Cleanup
	HardDeleteUploadsByIDs(ctx context.Context, ids ...int) error
	DeleteUploadsStuckUploading(ctx context.Context, uploadedBefore time.Time) (int, int, error)
	DeleteUploadsWithoutRepository(ctx context.Context, now time.Time) (int, int, error)
	DeleteOldAuditLogs(ctx context.Context, maxAge time.Duration, now time.Time) (numRecordsScanned, numRecordsAltered int, _ error)
	ReconcileCandidates(ctx context.Context, batchSize int) ([]int, error)
	ProcessStaleSourcedCommits(ctx context.Context, minimumTimeSinceLastCheck time.Duration, commitResolverBatchSize int, commitResolverMaximumCommitLag time.Duration, shouldDelete func(ctx context.Context, repositoryID int, repositoryName, commit string) (bool, error)) (int, int, error)
	DeleteIndexesWithoutRepository(ctx context.Context, now time.Time) (int, int, error)
	ExpireFailedRecords(ctx context.Context, batchSize int, failedIndexMaxAge time.Duration, now time.Time) (int, int, error)
	ProcessSourcedCommits(ctx context.Context, minimumTimeSinceLastCheck time.Duration, commitResolverMaximumCommitLag time.Duration, limit int, f func(ctx context.Context, repositoryID int, repositoryName, commit string) (bool, error), now time.Time) (int, int, error)

	// Misc
	HasRepository(ctx context.Context, repositoryID int) (bool, error)
	HasCommit(ctx context.Context, repositoryID int, commit string) (bool, error)
	InsertDependencySyncingJob(ctx context.Context, uploadID int) (int, error)
}

type SourcedCommits struct {
	RepositoryID   int
	RepositoryName string
	Commits        []string
}

type store struct {
	logger     logger.Logger
	db         *basestore.Store
	operations *operations
}

func New(observationCtx *observation.Context, db database.DB) Store {
	return &store{
		logger:     logger.Scoped("uploads.store", ""),
		db:         basestore.NewWithHandle(db.Handle()),
		operations: newOperations(observationCtx),
	}
}

func (s *store) Transact(ctx context.Context) (Store, error) {
	return s.transact(ctx)
}

func (s *store) transact(ctx context.Context) (*store, error) {
	tx, err := s.db.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &store{
		logger:     s.logger,
		db:         tx,
		operations: s.operations,
	}, nil
}

func (s *store) Done(err error) error {
	return s.db.Done(err)
}

//
//
//
//
//

type cteDefinition struct {
	name       string
	definition *sqlf.Query
}

type sanitizedCommitInput struct {
	nearestUploadsRowValues       <-chan []any
	nearestUploadsLinksRowValues  <-chan []any
	uploadsVisibleAtTipRowValues  <-chan []any
	numNearestUploadsRecords      uint32 // populated once nearestUploadsRowValues is exhausted
	numNearestUploadsLinksRecords uint32 // populated once nearestUploadsLinksRowValues is exhausted
	numUploadsVisibleAtTipRecords uint32 // populated once uploadsVisibleAtTipRowValues is exhausted
}

type uploadMetaListSerializer struct {
	buf     bytes.Buffer
	scratch []byte
}

func newUploadMetaListSerializer() *uploadMetaListSerializer {
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

// StalledUploadMaxAge is the maximum allowable duration between updating the state of an
// upload as "processing" and locking the upload row during processing. An unlocked row that
// is marked as processing likely indicates that the worker that dequeued the upload has died.
// There should be a nearly-zero delay between these states during normal operation.
const StalledUploadMaxAge = time.Second * 25

// UploadMaxNumResets is the maximum number of times an upload can be reset. If an upload's
// failed attempts counter reaches this threshold, it will be moved into "errored" rather than
// "queued" on its next reset.
const UploadMaxNumResets = 3

var uploadColumnsWithNullRank = []*sqlf.Query{
	sqlf.Sprintf("u.id"),
	sqlf.Sprintf("u.commit"),
	sqlf.Sprintf("u.root"),
	sqlf.Sprintf("EXISTS (" + visibleAtTipSubselectQuery + ") AS visible_at_tip"),
	sqlf.Sprintf("u.uploaded_at"),
	sqlf.Sprintf("u.state"),
	sqlf.Sprintf("u.failure_message"),
	sqlf.Sprintf("u.started_at"),
	sqlf.Sprintf("u.finished_at"),
	sqlf.Sprintf("u.process_after"),
	sqlf.Sprintf("u.num_resets"),
	sqlf.Sprintf("u.num_failures"),
	sqlf.Sprintf("u.repository_id"),
	sqlf.Sprintf("u.repository_name"),
	sqlf.Sprintf("u.indexer"),
	sqlf.Sprintf("u.indexer_version"),
	sqlf.Sprintf("u.num_parts"),
	sqlf.Sprintf("u.uploaded_parts"),
	sqlf.Sprintf("u.upload_size"),
	sqlf.Sprintf("u.associated_index_id"),
	sqlf.Sprintf("u.content_type"),
	sqlf.Sprintf("u.should_reindex"),
	sqlf.Sprintf("NULL"),
	sqlf.Sprintf("u.uncompressed_size"),
}

var UploadWorkerStoreOptions = dbworkerstore.Options[shared.Upload]{
	Name:              "codeintel_upload",
	TableName:         "lsif_uploads",
	ViewName:          "lsif_uploads_with_repository_name u",
	ColumnExpressions: uploadColumnsWithNullRank,
	Scan:              dbworkerstore.BuildWorkerScan(scanCompleteUpload),
	OrderByExpression: sqlf.Sprintf(`
		u.associated_index_id IS NULL DESC,
		COALESCE(u.process_after, u.uploaded_at),
		u.id
	`),
	StalledMaxAge: StalledUploadMaxAge,
	MaxNumResets:  UploadMaxNumResets,
}

func scanCompleteUpload(s dbutil.Scanner) (upload shared.Upload, _ error) {
	var rawUploadedParts []sql.NullInt32
	if err := s.Scan(
		&upload.ID,
		&upload.Commit,
		&upload.Root,
		&upload.VisibleAtTip,
		&upload.UploadedAt,
		&upload.State,
		&upload.FailureMessage,
		&upload.StartedAt,
		&upload.FinishedAt,
		&upload.ProcessAfter,
		&upload.NumResets,
		&upload.NumFailures,
		&upload.RepositoryID,
		&upload.RepositoryName,
		&upload.Indexer,
		&dbutil.NullString{S: &upload.IndexerVersion},
		&upload.NumParts,
		pq.Array(&rawUploadedParts),
		&upload.UploadSize,
		&upload.AssociatedIndexID,
		&upload.ContentType,
		&upload.ShouldReindex,
		&upload.Rank,
		&upload.UncompressedSize,
	); err != nil {
		return upload, err
	}

	upload.UploadedParts = make([]int, 0, len(rawUploadedParts))
	for _, uploadedPart := range rawUploadedParts {
		upload.UploadedParts = append(upload.UploadedParts, int(uploadedPart.Int32))
	}

	return upload, nil
}

var scanUploadComplete = basestore.NewSliceScanner(scanCompleteUpload)

// scanFirstUpload scans a slice of uploads from the return value of `*Store.query` and returns the first.
var scanFirstUpload = basestore.NewFirstScanner(scanCompleteUpload)

func scanCountsWithTotalCount(rows *sql.Rows, queryErr error) (totalCount int, _ map[int]int, err error) {
	if queryErr != nil {
		return 0, nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	visibilities := map[int]int{}
	for rows.Next() {
		var id int
		var count int
		if err := rows.Scan(&totalCount, &id, &count); err != nil {
			return 0, nil, err
		}

		visibilities[id] = count
	}

	return totalCount, visibilities, nil
}

// scanDumps scans a slice of dumps from the return value of `*Store.query`.
func scanDump(s dbutil.Scanner) (dump shared.Dump, err error) {
	return dump, s.Scan(
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
		&dbutil.NullString{S: &dump.IndexerVersion},
		&dump.AssociatedIndexID,
	)
}

var scanDumps = basestore.NewSliceScanner(scanDump)

// scanSourcedCommits scans triples of repository ids/repository names/commits from the
// return value of `*Store.query`. The output of this function is ordered by repository
// identifier, then by commit.
func scanSourcedCommits(rows *sql.Rows, queryErr error) (_ []SourcedCommits, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	sourcedCommitsMap := map[int]SourcedCommits{}
	for rows.Next() {
		var repositoryID int
		var repositoryName string
		var commit string
		if err := rows.Scan(&repositoryID, &repositoryName, &commit); err != nil {
			return nil, err
		}

		sourcedCommitsMap[repositoryID] = SourcedCommits{
			RepositoryID:   repositoryID,
			RepositoryName: repositoryName,
			Commits:        append(sourcedCommitsMap[repositoryID].Commits, commit),
		}
	}

	flattened := make([]SourcedCommits, 0, len(sourcedCommitsMap))
	for _, sourcedCommits := range sourcedCommitsMap {
		sort.Strings(sourcedCommits.Commits)
		flattened = append(flattened, sourcedCommits)
	}

	sort.Slice(flattened, func(i, j int) bool {
		return flattened[i].RepositoryID < flattened[j].RepositoryID
	})
	return flattened, nil
}

func scanCount(rows *sql.Rows, queryErr error) (value int, err error) {
	if queryErr != nil {
		return 0, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		if err := rows.Scan(&value); err != nil {
			return 0, err
		}
	}

	return value, nil
}

func scanPairOfCounts(rows *sql.Rows, queryErr error) (value1, value2 int, err error) {
	if queryErr != nil {
		return 0, 0, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		if err := rows.Scan(&value1, &value2); err != nil {
			return 0, 0, err
		}
	}

	return value1, value2, nil
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

func scanRepoNames(rows *sql.Rows, queryErr error) (_ map[int]string, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	names := map[int]string{}

	for rows.Next() {
		var (
			id   int
			name string
		)
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}

		names[id] = name
	}

	return names, nil
}

func scanUploadAuditLog(s dbutil.Scanner) (log shared.UploadLog, _ error) {
	hstores := pgtype.HstoreArray{}
	err := s.Scan(
		&log.LogTimestamp,
		&log.RecordDeletedAt,
		&log.UploadID,
		&log.Commit,
		&log.Root,
		&log.RepositoryID,
		&log.UploadedAt,
		&log.Indexer,
		&log.IndexerVersion,
		&log.UploadSize,
		&log.AssociatedIndexID,
		&hstores,
		&log.Reason,
		&log.Operation,
	)

	for _, hstore := range hstores.Elements {
		m := make(map[string]*string)
		if err := hstore.AssignTo(&m); err != nil {
			return log, err
		}
		log.TransitionColumns = append(log.TransitionColumns, m)
	}

	return log, err
}

var scanUploadAuditLogs = basestore.NewSliceScanner(scanUploadAuditLog)

// scanIndexes scans a slice of indexes from the return value of `*Store.query`.
var scanIndexes = basestore.NewSliceScanner(scanIndex)

// scanFirstIndex scans a slice of indexes from the return value of `*Store.query` and returns the first.
var scanFirstIndex = basestore.NewFirstScanner(scanIndex)

func scanIndex(s dbutil.Scanner) (index uploadsshared.Index, err error) {
	var executionLogs []executor.ExecutionLogEntry
	if err := s.Scan(
		&index.ID,
		&index.Commit,
		&index.QueuedAt,
		&index.State,
		&index.FailureMessage,
		&index.StartedAt,
		&index.FinishedAt,
		&index.ProcessAfter,
		&index.NumResets,
		&index.NumFailures,
		&index.RepositoryID,
		&index.RepositoryName,
		pq.Array(&index.DockerSteps),
		&index.Root,
		&index.Indexer,
		pq.Array(&index.IndexerArgs),
		&index.Outfile,
		pq.Array(&executionLogs),
		&index.Rank,
		pq.Array(&index.LocalSteps),
		&index.AssociatedUploadID,
		&index.ShouldReindex,
		pq.Array(&index.RequestedEnvVars),
	); err != nil {
		return index, err
	}

	index.ExecutionLogs = append(index.ExecutionLogs, executionLogs...)

	return index, nil
}

const sanitizedIndexerExpression = `
(
    split_part(
        split_part(
            CASE
                -- Strip sourcegraph/ prefix if it exists
                WHEN strpos(indexer, 'sourcegraph/') = 1 THEN substr(indexer, length('sourcegraph/') + 1)
                ELSE indexer
            END,
        '@', 1), -- strip off @sha256:...
    ':', 1) -- strip off tag
)
`

const indexAssociatedUploadIDQueryFragment = `
(
	SELECT MAX(id) FROM lsif_uploads WHERE associated_index_id = u.id
) AS associated_upload_id
`

const indexRankQueryFragment = `
SELECT
	r.id,
	ROW_NUMBER() OVER (ORDER BY COALESCE(r.process_after, r.queued_at), r.id) as rank
FROM lsif_indexes_with_repository_name r
WHERE r.state = 'queued'
`

const recentIndexesSummaryQuery = `
WITH ranked_completed AS (
	SELECT
		u.id,
		u.root,
		u.indexer,
		u.finished_at,
		RANK() OVER (PARTITION BY root, ` + sanitizedIndexerExpression + ` ORDER BY finished_at DESC) AS rank
	FROM lsif_indexes u
	WHERE
		u.repository_id = %s AND
		u.state NOT IN ('queued', 'processing', 'deleted')
),
latest_indexes AS (
	SELECT u.id, u.root, u.indexer, u.queued_at
	FROM lsif_indexes u
	WHERE
		u.id IN (
			SELECT rc.id
			FROM ranked_completed rc
			WHERE rc.rank = 1
		)
	ORDER BY u.root, u.indexer
),
new_indexes AS (
	SELECT u.id
	FROM lsif_indexes u
	WHERE
		u.repository_id = %s AND
		u.state IN ('queued', 'processing') AND
		u.queued_at >= (
			SELECT lu.queued_at
			FROM latest_indexes lu
			WHERE
				lu.root = u.root AND
				lu.indexer = u.indexer
			-- condition passes when latest_indexes is empty
			UNION SELECT u.queued_at LIMIT 1
		)
)
SELECT
	u.id,
	u.commit,
	u.queued_at,
	u.state,
	u.failure_message,
	u.started_at,
	u.finished_at,
	u.process_after,
	u.num_resets,
	u.num_failures,
	u.repository_id,
	u.repository_name,
	u.docker_steps,
	u.root,
	u.indexer,
	u.indexer_args,
	u.outfile,
	u.execution_logs,
	s.rank,
	u.local_steps,
	` + indexAssociatedUploadIDQueryFragment + `,
	u.should_reindex,
	u.requested_envvars
FROM lsif_indexes_with_repository_name u
LEFT JOIN (` + indexRankQueryFragment + `) s
ON u.id = s.id
WHERE u.id IN (
	SELECT lu.id FROM latest_indexes lu
	UNION
	SELECT nu.id FROM new_indexes nu
)
ORDER BY u.root, u.indexer
`

// DefinitionDumpsLimit is the maximum number of records that can be returned from DefinitionDumps.
var DefinitionDumpsLimit, _ = strconv.ParseInt(env.Get("PRECISE_CODE_INTEL_DEFINITION_DUMPS_LIMIT", "100", "The maximum number of dumps that can define the same package."), 10, 64)

func monikersToString(vs []precise.QualifiedMonikerData) string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, fmt.Sprintf("%s:%s:%s:%s:%s", v.Kind, v.Scheme, v.Manager, v.Identifier, v.Version))
	}

	return strings.Join(strs, ", ")
}

func makeFindClosestDumpConditions(path string, rootMustEnclosePath bool, indexer string) (conds []*sqlf.Query) {
	if rootMustEnclosePath {
		// Ensure that the root is a prefix of the path
		conds = append(conds, sqlf.Sprintf(`%s LIKE (u.root || '%%%%')`, path))
	} else {
		// Ensure that the root is a prefix of the path or vice versa
		conds = append(conds, sqlf.Sprintf(`(%s LIKE (u.root || '%%%%') OR u.root LIKE (%s || '%%%%'))`, path, path))
	}
	if indexer != "" {
		conds = append(conds, sqlf.Sprintf("indexer = %s", indexer))
	}

	return conds
}

// makeVisibleUploadsQuery returns a SQL query returning the set of identifiers of uploads
// visible from the given commit. This is done by removing the "shadowed" values created
// by looking at a commit and it's ancestors visible commits.
func makeVisibleUploadsQuery(repositoryID int, commit string) *sqlf.Query {
	return sqlf.Sprintf(visibleUploadsQuery, makeVisibleUploadCandidatesQuery(repositoryID, commit))
}

const visibleUploadsQuery = `
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
`

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

	return sqlf.Sprintf(visibleUploadCandidatesQuery, repositoryID, sqlf.Join(commitQueries, ", "), repositoryID, sqlf.Join(commitQueries, ", "))
}

const visibleUploadCandidatesQuery = `
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
`

type backfillIncompleteError struct {
	repositoryID int
}

func (e backfillIncompleteError) Error() string {
	return fmt.Sprintf("repository %d has not yet completed its backfill of commit dates", e.repositoryID)
}

// scanCommitGraphMetadata scans a a commit graph metadata row from the return value of `*Store.query`.
func scanCommitGraphMetadata(rows *sql.Rows, queryErr error) (updateToken, dirtyToken int, updatedAt *time.Time, _ bool, err error) {
	if queryErr != nil {
		return 0, 0, nil, false, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	if rows.Next() {
		if err := rows.Scan(&updateToken, &dirtyToken, &updatedAt); err != nil {
			return 0, 0, nil, false, err
		}

		return updateToken, dirtyToken, updatedAt, true, nil
	}

	return 0, 0, nil, false, nil
}

const uploadRankQueryFragment = `
SELECT
	r.id,
	ROW_NUMBER() OVER (ORDER BY
		-- Note: this should be kept in-sync with the order given to workerutil
		r.associated_index_id IS NULL DESC,
		COALESCE(r.process_after, r.uploaded_at),
		r.id
	) as rank
FROM lsif_uploads_with_repository_name r
WHERE r.state = 'queued'
`

const visibleAtTipSubselectQuery = `
SELECT 1
FROM lsif_uploads_visible_at_tip uvt
WHERE
	uvt.repository_id = u.repository_id AND
	uvt.upload_id = u.id AND
	uvt.is_default_branch
`

const deletedUploadsFromAuditLogsCTEQuery = `
SELECT
	DISTINCT ON(s.upload_id) s.upload_id AS id, au.commit, au.root,
	au.uploaded_at, 'deleted' AS state,
	snapshot->'failure_message' AS failure_message,
	(snapshot->'started_at')::timestamptz AS started_at,
	(snapshot->'finished_at')::timestamptz AS finished_at,
	(snapshot->'process_after')::timestamptz AS process_after,
	COALESCE((snapshot->'num_resets')::integer, -1) AS num_resets,
	COALESCE((snapshot->'num_failures')::integer, -1) AS num_failures,
	au.repository_id,
	au.indexer, au.indexer_version,
	COALESCE((snapshot->'num_parts')::integer, -1) AS num_parts,
	NULL::integer[] as uploaded_parts,
	au.upload_size, au.associated_index_id, au.content_type,
	false AS should_reindex, -- TODO
	COALESCE((snapshot->'expired')::boolean, false) AS expired,
	NULL::bigint AS uncompressed_size
FROM (
	SELECT upload_id, snapshot_transition_columns(transition_columns ORDER BY sequence ASC) AS snapshot
	FROM lsif_uploads_audit_logs
	WHERE record_deleted_at IS NOT NULL
	GROUP BY upload_id
) AS s
JOIN lsif_uploads_audit_logs au ON au.upload_id = s.upload_id
`

const rankedDependencyCandidateCTEQuery = `
SELECT
	p.dump_id as pkg_id,
	r.dump_id as ref_id,
	-- Rank each upload providing the same package from the same directory
	-- within a repository by commit date. We'll choose the oldest commit
	-- date as the canonical choice and ignore the uploads for younger
	-- commits providing the same package.
	` + packageRankingQueryFragment + ` AS rank
FROM lsif_uploads u
JOIN lsif_packages p ON p.dump_id = u.id
JOIN lsif_references r ON
	r.scheme = p.scheme AND
	r.manager = p.manager AND
	r.name = p.name AND
	r.version = p.version AND
	r.dump_id != p.dump_id
WHERE
	-- Don't match deleted uploads
	u.state = 'completed' AND
	%s
`

// packageRankingQueryFragment uses `lsif_uploads u` JOIN `lsif_packages p` to return a rank
// for each row grouped by package and source code location and ordered by the associated Git
// commit date.
const packageRankingQueryFragment = `
rank() OVER (
	PARTITION BY
		-- Group providers of the same package together
		p.scheme, p.manager, p.name, p.version,
		-- Defined by the same directory within a repository
		u.repository_id, u.indexer, u.root
	ORDER BY
		-- Rank each grouped upload by the associated commit date
		(SELECT cd.committed_at FROM codeintel_commit_dates cd WHERE cd.repository_id = u.repository_id AND cd.commit_bytea = decode(u.commit, 'hex')) NULLS LAST,
		-- Break ties via the unique identifier
		u.id
)
`

const rankedDependentCandidateCTEQuery = `
SELECT
	p.dump_id AS pkg_id,
	p.scheme AS scheme,
	p.manager AS manager,
	p.name AS name,
	p.version AS version,
	-- Rank each upload providing the same package from the same directory
	-- within a repository by commit date. We'll choose the oldest commit
	-- date as the canonical choice and ignore the uploads for younger
	-- commits providing the same package.
	` + packageRankingQueryFragment + ` AS rank
FROM lsif_uploads u
JOIN lsif_packages p ON p.dump_id = u.id
WHERE
	-- Don't match deleted uploads
	u.state = 'completed' AND
	%s
`

// DeletedRepositoryGracePeriod is the minimum allowable duration between a repo deletion
// and the upload and index records for that repository being deleted.
const DeletedRepositoryGracePeriod = time.Minute * 30

const referenceIDsCTEDefinitions = `
WITH
visible_uploads AS (
	(%s)
	UNION
	(SELECT uvt.upload_id FROM lsif_uploads_visible_at_tip uvt WHERE uvt.repository_id != %s AND uvt.is_default_branch)
)
`

const referenceIDsBaseQuery = `
FROM lsif_references r
LEFT JOIN lsif_dumps u ON u.id = r.dump_id
JOIN repo ON repo.id = u.repository_id
WHERE
	(r.scheme, r.manager, r.name, r.version) IN (%s) AND
	r.dump_id IN (SELECT * FROM visible_uploads) AND
	%s -- authz conds
`

const referenceIDsQuery = referenceIDsCTEDefinitions + `
SELECT r.dump_id, r.scheme, r.manager, r.name, r.version
` + referenceIDsBaseQuery + `
ORDER BY dump_id
LIMIT %s OFFSET %s
`

const referenceIDsCountQuery = referenceIDsCTEDefinitions + `
SELECT COUNT(distinct r.dump_id)
` + referenceIDsBaseQuery

// sanitizeCommitInput reads the data that needs to be persisted from the given graph and writes the
// sanitized values (ensures values match the column types) into channels for insertion into a particular
// table.
func sanitizeCommitInput(
	ctx context.Context,
	graph *commitgraph.Graph,
	refDescriptions map[string][]gitdomain.RefDescription,
	maxAgeForNonStaleBranches time.Duration,
	maxAgeForNonStaleTags time.Duration,
) *sanitizedCommitInput {
	maxAges := map[gitdomain.RefType]time.Duration{
		gitdomain.RefTypeBranch: maxAgeForNonStaleBranches,
		gitdomain.RefTypeTag:    maxAgeForNonStaleTags,
	}

	nearestUploadsRowValues := make(chan []any)
	nearestUploadsLinksRowValues := make(chan []any)
	uploadsVisibleAtTipRowValues := make(chan []any)

	sanitized := &sanitizedCommitInput{
		nearestUploadsRowValues:      nearestUploadsRowValues,
		nearestUploadsLinksRowValues: nearestUploadsLinksRowValues,
		uploadsVisibleAtTipRowValues: uploadsVisibleAtTipRowValues,
	}

	go func() {
		defer close(nearestUploadsRowValues)
		defer close(nearestUploadsLinksRowValues)
		defer close(uploadsVisibleAtTipRowValues)

		listSerializer := newUploadMetaListSerializer()

		for envelope := range graph.Stream() {
			if envelope.Uploads != nil {
				if !countingWrite(
					ctx,
					nearestUploadsRowValues,
					&sanitized.numNearestUploadsRecords,
					// row values
					dbutil.CommitBytea(envelope.Uploads.Commit),
					listSerializer.Serialize(envelope.Uploads.Uploads),
				) {
					return
				}
			}

			if envelope.Links != nil {
				if !countingWrite(
					ctx,
					nearestUploadsLinksRowValues,
					&sanitized.numNearestUploadsLinksRecords,
					// row values
					dbutil.CommitBytea(envelope.Links.Commit),
					dbutil.CommitBytea(envelope.Links.AncestorCommit),
					envelope.Links.Distance,
				) {
					return
				}
			}
		}

		for commit, refDescriptions := range refDescriptions {
			isDefaultBranch := false
			names := make([]string, 0, len(refDescriptions))

			for _, refDescription := range refDescriptions {
				if refDescription.IsDefaultBranch {
					isDefaultBranch = true
				} else {
					maxAge, ok := maxAges[refDescription.Type]
					if !ok || refDescription.CreatedDate == nil || time.Since(*refDescription.CreatedDate) > maxAge {
						continue
					}
				}

				names = append(names, refDescription.Name)
			}
			sort.Strings(names)

			if len(names) == 0 {
				continue
			}

			for _, uploadMeta := range graph.UploadsVisibleAtCommit(commit) {
				if !countingWrite(
					ctx,
					uploadsVisibleAtTipRowValues,
					&sanitized.numUploadsVisibleAtTipRecords,
					// row values
					uploadMeta.UploadID,
					strings.Join(names, ","),
					isDefaultBranch,
				) {
					return
				}
			}
		}
	}()

	return sanitized
}

// writeVisibleUploads serializes the given input into a the following set of temporary tables in the database.
//
//   - t_lsif_nearest_uploads        (mirroring lsif_nearest_uploads)
//   - t_lsif_nearest_uploads_links  (mirroring lsif_nearest_uploads_links)
//   - t_lsif_uploads_visible_at_tip (mirroring lsif_uploads_visible_at_tip)
//
// The data in these temporary tables can then be moved into a persisted/permanent table. We previously would perform a
// bulk delete of the records associated with a repository, then reinsert all of the data needed to be persisted. This
// caused massive table bloat on some instances. Storing into a temporary table and then inserting/updating/deleting
// records into the persisted table minimizes the number of tuples we need to touch and drastically reduces table bloat.
func (s *store) writeVisibleUploads(ctx context.Context, sanitizedInput *sanitizedCommitInput, tx *basestore.Store) (err error) {
	ctx, trace, endObservation := s.operations.writeVisibleUploads.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	if err := s.createTemporaryNearestUploadsTables(ctx, tx); err != nil {
		return err
	}

	g, gctx := errgroup.WithContext(ctx)

	// Insert the set of uploads that are visible from each commit for a given repository into a temporary table.
	nearestUploadsWriter := func() error {
		return batch.InsertValues(
			gctx,
			tx.Handle(),
			"t_lsif_nearest_uploads",
			batch.MaxNumPostgresParameters,
			[]string{"commit_bytea", "uploads"},
			sanitizedInput.nearestUploadsRowValues,
		)
	}

	// Insert the commits not inserted into the table above by adding links to a unique ancestor and their relative
	// distance in the graph into another temporary table. We use this as a cheap way to reconstruct the full data
	// set, which is multiplicative in the size of the commit graph AND the number of unique roots.
	nearestUploadsLinksWriter := func() error {
		return batch.InsertValues(
			gctx,
			tx.Handle(),
			"t_lsif_nearest_uploads_links",
			batch.MaxNumPostgresParameters,
			[]string{"commit_bytea", "ancestor_commit_bytea", "distance"},
			sanitizedInput.nearestUploadsLinksRowValues,
		)
	}

	// Insert the set of uploads visible from the tip of the default branch into a temporary table. These values are
	// used to determine which bundles for a repository we open during a global find references query.
	uploadsVisibleAtTipWriter := func() error {
		return batch.InsertValues(
			gctx,
			tx.Handle(),
			"t_lsif_uploads_visible_at_tip",
			batch.MaxNumPostgresParameters,
			[]string{"upload_id", "branch_or_tag_name", "is_default_branch"},
			sanitizedInput.uploadsVisibleAtTipRowValues,
		)
	}

	g.Go(nearestUploadsWriter)
	g.Go(nearestUploadsLinksWriter)
	g.Go(uploadsVisibleAtTipWriter)

	if err := g.Wait(); err != nil {
		return err
	}
	trace.AddEvent("TODO Domain Owner",
		attribute.Int("numNearestUploadsRecords", int(sanitizedInput.numNearestUploadsRecords)),
		attribute.Int("numNearestUploadsLinksRecords", int(sanitizedInput.numNearestUploadsLinksRecords)),
		attribute.Int("numUploadsVisibleAtTipRecords", int(sanitizedInput.numUploadsVisibleAtTipRecords)))

	return nil
}

// persistNearestUploads modifies the lsif_nearest_uploads table so that it has same data
// as t_lsif_nearest_uploads for the given repository.
func (s *store) persistNearestUploads(ctx context.Context, repositoryID int, tx *basestore.Store) (err error) {
	ctx, trace, endObservation := s.operations.persistNearestUploads.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	rowsInserted, rowsUpdated, rowsDeleted, err := s.bulkTransfer(
		ctx,
		sqlf.Sprintf(nearestUploadsInsertQuery, repositoryID, repositoryID),
		sqlf.Sprintf(nearestUploadsUpdateQuery, repositoryID),
		sqlf.Sprintf(nearestUploadsDeleteQuery, repositoryID),
		tx,
	)
	if err != nil {
		return err
	}
	trace.AddEvent("TODO Domain Owner",
		attribute.Int("lsif_nearest_uploads.ins", rowsInserted),
		attribute.Int("lsif_nearest_uploads.upd", rowsUpdated),
		attribute.Int("lsif_nearest_uploads.del", rowsDeleted))

	return nil
}

const nearestUploadsInsertQuery = `
INSERT INTO lsif_nearest_uploads
SELECT %s, source.commit_bytea, source.uploads
FROM t_lsif_nearest_uploads source
WHERE source.commit_bytea NOT IN (SELECT nu.commit_bytea FROM lsif_nearest_uploads nu WHERE nu.repository_id = %s)
`

const nearestUploadsUpdateQuery = `
UPDATE lsif_nearest_uploads nu
SET uploads = source.uploads
FROM t_lsif_nearest_uploads source
WHERE
	nu.repository_id = %s AND
	nu.commit_bytea = source.commit_bytea AND
	nu.uploads != source.uploads
`

const nearestUploadsDeleteQuery = `
DELETE FROM lsif_nearest_uploads nu
WHERE
	nu.repository_id = %s AND
	nu.commit_bytea NOT IN (SELECT source.commit_bytea FROM t_lsif_nearest_uploads source)
`

// persistNearestUploadsLinks modifies the lsif_nearest_uploads_links table so that it has same
// data as t_lsif_nearest_uploads_links for the given repository.
func (s *store) persistNearestUploadsLinks(ctx context.Context, repositoryID int, tx *basestore.Store) (err error) {
	ctx, trace, endObservation := s.operations.persistNearestUploadsLinks.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	rowsInserted, rowsUpdated, rowsDeleted, err := s.bulkTransfer(
		ctx,
		sqlf.Sprintf(nearestUploadsLinksInsertQuery, repositoryID, repositoryID),
		sqlf.Sprintf(nearestUploadsLinksUpdateQuery, repositoryID),
		sqlf.Sprintf(nearestUploadsLinksDeleteQuery, repositoryID),
		tx,
	)
	if err != nil {
		return err
	}
	trace.AddEvent("TODO Domain Owner",
		attribute.Int("lsif_nearest_uploads_links.ins", rowsInserted),
		attribute.Int("lsif_nearest_uploads_links.upd", rowsUpdated),
		attribute.Int("lsif_nearest_uploads_links.del", rowsDeleted))

	return nil
}

const nearestUploadsLinksInsertQuery = `
INSERT INTO lsif_nearest_uploads_links
SELECT %s, source.commit_bytea, source.ancestor_commit_bytea, source.distance
FROM t_lsif_nearest_uploads_links source
WHERE source.commit_bytea NOT IN (SELECT nul.commit_bytea FROM lsif_nearest_uploads_links nul WHERE nul.repository_id = %s)
`

const nearestUploadsLinksUpdateQuery = `
UPDATE lsif_nearest_uploads_links nul
SET ancestor_commit_bytea = source.ancestor_commit_bytea, distance = source.distance
FROM t_lsif_nearest_uploads_links source
WHERE
	nul.repository_id = %s AND
	nul.commit_bytea = source.commit_bytea AND
	nul.ancestor_commit_bytea != source.ancestor_commit_bytea AND
	nul.distance != source.distance
`

const nearestUploadsLinksDeleteQuery = `
DELETE FROM lsif_nearest_uploads_links nul
WHERE
	nul.repository_id = %s AND
	nul.commit_bytea NOT IN (SELECT source.commit_bytea FROM t_lsif_nearest_uploads_links source)
`

// persistUploadsVisibleAtTip modifies the lsif_uploads_visible_at_tip table so that it has same
// data as t_lsif_uploads_visible_at_tip for the given repository.
func (s *store) persistUploadsVisibleAtTip(ctx context.Context, repositoryID int, tx *basestore.Store) (err error) {
	ctx, trace, endObservation := s.operations.persistUploadsVisibleAtTip.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	insertQuery := sqlf.Sprintf(uploadsVisibleAtTipInsertQuery, repositoryID, repositoryID)
	deleteQuery := sqlf.Sprintf(uploadsVisibleAtTipDeleteQuery, repositoryID)

	rowsInserted, rowsUpdated, rowsDeleted, err := s.bulkTransfer(ctx, insertQuery, nil, deleteQuery, tx)
	if err != nil {
		return err
	}
	trace.AddEvent("TODO Domain Owner",
		attribute.Int("lsif_uploads_visible_at_tip.ins", rowsInserted),
		attribute.Int("lsif_uploads_visible_at_tip.upd", rowsUpdated),
		attribute.Int("lsif_uploads_visible_at_tip.del", rowsDeleted))

	return nil
}

const uploadsVisibleAtTipInsertQuery = `
INSERT INTO lsif_uploads_visible_at_tip
SELECT %s, source.upload_id, source.branch_or_tag_name, source.is_default_branch
FROM t_lsif_uploads_visible_at_tip source
WHERE NOT EXISTS (
	SELECT 1
	FROM lsif_uploads_visible_at_tip vat
	WHERE
		vat.repository_id = %s AND
		vat.upload_id = source.upload_id AND
		vat.branch_or_tag_name = source.branch_or_tag_name AND
		vat.is_default_branch = source.is_default_branch
)
`

const uploadsVisibleAtTipDeleteQuery = `
DELETE FROM lsif_uploads_visible_at_tip vat
WHERE
	vat.repository_id = %s AND
	NOT EXISTS (
		SELECT 1
		FROM t_lsif_uploads_visible_at_tip source
		WHERE
			source.upload_id = vat.upload_id AND
			source.branch_or_tag_name = vat.branch_or_tag_name AND
			source.is_default_branch = vat.is_default_branch
	)
`

// bulkTransfer performs the given insert, update, and delete queries and returns the number of records
// touched by each. If any query is nil, the returned count will be zero.
func (s *store) bulkTransfer(ctx context.Context, insertQuery, updateQuery, deleteQuery *sqlf.Query, tx *basestore.Store) (rowsInserted int, rowsUpdated int, rowsDeleted int, err error) {
	prepareQuery := func(query *sqlf.Query) *sqlf.Query {
		if query == nil {
			return sqlf.Sprintf("SELECT 0")
		}

		return sqlf.Sprintf("%s RETURNING 1", query)
	}

	rows, err := tx.Query(ctx, sqlf.Sprintf(bulkTransferQuery, prepareQuery(insertQuery), prepareQuery(updateQuery), prepareQuery(deleteQuery)))
	if err != nil {
		return 0, 0, 0, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	if rows.Next() {
		if err := rows.Scan(&rowsInserted, &rowsUpdated, &rowsDeleted); err != nil {
			return 0, 0, 0, err
		}

		return rowsInserted, rowsUpdated, rowsDeleted, nil
	}

	return 0, 0, 0, nil
}

const bulkTransferQuery = `
WITH
	ins AS (%s),
	upd AS (%s),
	del AS (%s)
SELECT
	(SELECT COUNT(*) FROM ins) AS num_ins,
	(SELECT COUNT(*) FROM upd) AS num_upd,
	(SELECT COUNT(*) FROM del) AS num_del
`

func (s *store) createTemporaryNearestUploadsTables(ctx context.Context, tx *basestore.Store) error {
	temporaryTableQueries := []string{
		temporaryNearestUploadsTableQuery,
		temporaryNearestUploadsLinksTableQuery,
		temporaryUploadsVisibleAtTipTableQuery,
	}

	for _, temporaryTableQuery := range temporaryTableQueries {
		if err := tx.Exec(ctx, sqlf.Sprintf(temporaryTableQuery)); err != nil {
			return err
		}
	}

	return nil
}

const temporaryNearestUploadsTableQuery = `
CREATE TEMPORARY TABLE t_lsif_nearest_uploads (
	commit_bytea bytea NOT NULL,
	uploads      jsonb NOT NULL
) ON COMMIT DROP
`

const temporaryNearestUploadsLinksTableQuery = `
CREATE TEMPORARY TABLE t_lsif_nearest_uploads_links (
	commit_bytea          bytea NOT NULL,
	ancestor_commit_bytea bytea NOT NULL,
	distance              integer NOT NULL
) ON COMMIT DROP
`

const temporaryUploadsVisibleAtTipTableQuery = `
CREATE TEMPORARY TABLE t_lsif_uploads_visible_at_tip (
	upload_id integer NOT NULL,
	branch_or_tag_name text NOT NULL,
	is_default_branch boolean NOT NULL
) ON COMMIT DROP
`

// countingWrite writes the given slice of interfaces to the given channel. This function returns true
// if the write succeeded and false if the context was canceled. On success, the counter's underlying
// value will be incremented (non-atomically).
func countingWrite(ctx context.Context, ch chan<- []any, counter *uint32, values ...any) bool {
	select {
	case ch <- values:
		*counter++
		return true

	case <-ctx.Done():
		return false
	}
}

func intsToString(vs []int) string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, strconv.Itoa(v))
	}

	return strings.Join(strs, ", ")
}

func nilTimeToString(t *time.Time) string {
	if t == nil {
		return ""
	}

	return t.String()
}
