package store

import (
	"context"
	"time"

	logger "github.com/sourcegraph/log"

	autoindexingshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// Store provides the interface for uploads storage.
type Store interface {
	// Transaction
	Transact(ctx context.Context) (Store, error)
	Done(err error) error

	GetIndexes(ctx context.Context, opts shared.GetIndexesOptions) (_ []types.Index, _ int, err error)
	GetIndexByID(ctx context.Context, id int) (_ types.Index, _ bool, err error)
	GetIndexesByIDs(ctx context.Context, ids ...int) (_ []types.Index, err error)
	DeleteIndexByID(ctx context.Context, id int) (_ bool, err error)
	DeleteIndexes(ctx context.Context, opts shared.DeleteIndexesOptions) (err error)
	ReindexIndexByID(ctx context.Context, id int) (err error)
	ReindexIndexes(ctx context.Context, opts shared.ReindexIndexesOptions) (err error)

	// Commits
	GetCommitsVisibleToUpload(ctx context.Context, uploadID, limit int, token *string) (_ []string, nextToken *string, err error)
	GetOldestCommitDate(ctx context.Context, repositoryID int) (time.Time, bool, error)
	GetStaleSourcedCommits(ctx context.Context, minimumTimeSinceLastCheck time.Duration, limit int, now time.Time) (_ []shared.SourcedCommits, err error)
	ProcessSourcedCommits(
		ctx context.Context,
		minimumTimeSinceLastCheck time.Duration,
		commitResolverMaximumCommitLag time.Duration,
		limit int,
		f func(ctx context.Context, repositoryID int, repositoryName, commit string) (bool, error),
		now time.Time,
	) (_, _ int, err error)
	GetCommitGraphMetadata(ctx context.Context, repositoryID int) (stale bool, updatedAt *time.Time, err error)
	UpdateSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (uploadsUpdated int, err error)
	DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, maximumCommitLag time.Duration, now time.Time) (uploadsUpdated int, uploadsDeleted int, err error)
	HasCommit(ctx context.Context, repositoryID int, commit string) (_ bool, err error)

	// Repositories
	GetRepositoriesMaxStaleAge(ctx context.Context) (_ time.Duration, err error)
	SetRepositoryAsDirty(ctx context.Context, repositoryID int) (err error)
	GetDirtyRepositories(ctx context.Context) (_ []shared.DirtyRepository, err error)
	RepoName(ctx context.Context, repositoryID int) (_ string, err error)              // TODO(numbers88s): renaming this after I remove dbStore from gitserver init.
	RepoNames(ctx context.Context, repositoryIDs ...int) (_ map[int]string, err error) // TODO(numbers88s): renaming this after I remove dbStore from gitserver init.
	SetRepositoriesForRetentionScan(ctx context.Context, processDelay time.Duration, limit int) (_ []int, err error)
	SetRepositoriesForRetentionScanWithTime(ctx context.Context, processDelay time.Duration, limit int, now time.Time) (_ []int, err error)
	HasRepository(ctx context.Context, repositoryID int) (_ bool, err error)

	// Uploads
	GetIndexers(ctx context.Context, opts shared.GetIndexersOptions) ([]string, error)
	GetUploads(ctx context.Context, opts shared.GetUploadsOptions) (_ []types.Upload, _ int, err error)
	GetUploadByID(ctx context.Context, id int) (_ types.Upload, _ bool, err error)
	GetUploadsByIDs(ctx context.Context, ids ...int) (_ []types.Upload, err error)
	GetUploadsByIDsAllowDeleted(ctx context.Context, ids ...int) (_ []types.Upload, err error)
	GetUploadIDsWithReferences(ctx context.Context, orderedMonikers []precise.QualifiedMonikerData, ignoreIDs []int, repositoryID int, commit string, limit int, offset int, trace observation.TraceLogger) (ids []int, recordsScanned int, totalCount int, err error)
	GetVisibleUploadsMatchingMonikers(ctx context.Context, repositoryID int, commit string, orderedMonikers []precise.QualifiedMonikerData, limit, offset int) (_ shared.PackageReferenceScanner, _ int, err error)
	GetRecentUploadsSummary(ctx context.Context, repositoryID int) (upload []shared.UploadsWithRepositoryNamespace, err error)
	GetLastUploadRetentionScanForRepository(ctx context.Context, repositoryID int) (_ *time.Time, err error)
	UpdateUploadsVisibleToCommits(ctx context.Context, repositoryID int, graph *gitdomain.CommitGraph, refDescriptions map[string][]gitdomain.RefDescription, maxAgeForNonStaleBranches, maxAgeForNonStaleTags time.Duration, dirtyToken int, now time.Time) error
	UpdateUploadRetention(ctx context.Context, protectedIDs, expiredIDs []int) (err error)
	SourcedCommitsWithoutCommittedAt(ctx context.Context, batchSize int) ([]shared.SourcedCommits, error)
	UpdateCommittedAt(ctx context.Context, repositoryID int, commit, commitDateString string) error
	SoftDeleteExpiredUploads(ctx context.Context, batchSize int) (int, int, error)
	SoftDeleteExpiredUploadsViaTraversal(ctx context.Context, maxTraversal int) (int, int, error)
	HardDeleteUploadsByIDs(ctx context.Context, ids ...int) error
	DeleteUploadsStuckUploading(ctx context.Context, uploadedBefore time.Time) (_, _ int, err error)
	DeleteUploadsWithoutRepository(ctx context.Context, now time.Time) (_, _ int, err error)
	DeleteUploadByID(ctx context.Context, id int) (_ bool, err error)
	DeleteUploads(ctx context.Context, opts shared.DeleteUploadsOptions) (err error)

	// Uploads (uploading)
	InsertUpload(ctx context.Context, upload types.Upload) (int, error)
	AddUploadPart(ctx context.Context, uploadID, partIndex int) error
	MarkQueued(ctx context.Context, id int, uploadSize *int64) error
	MarkFailed(ctx context.Context, id int, reason string) error

	// Dumps
	FindClosestDumps(ctx context.Context, repositoryID int, commit, path string, rootMustEnclosePath bool, indexer string) (_ []types.Dump, err error)
	FindClosestDumpsFromGraphFragment(ctx context.Context, repositoryID int, commit, path string, rootMustEnclosePath bool, indexer string, commitGraph *gitdomain.CommitGraph) (_ []types.Dump, err error)
	GetDumpsWithDefinitionsForMonikers(ctx context.Context, monikers []precise.QualifiedMonikerData) (_ []types.Dump, err error)
	GetDumpsByIDs(ctx context.Context, ids []int) (_ []types.Dump, err error)
	DeleteOverlappingDumps(ctx context.Context, repositoryID int, commit, root, indexer string) error

	// Packages
	UpdatePackages(ctx context.Context, dumpID int, packages []precise.Package) (err error)

	// References
	UpdatePackageReferences(ctx context.Context, dumpID int, references []precise.PackageReference) (err error)
	ReferencesForUpload(ctx context.Context, uploadID int) (_ shared.PackageReferenceScanner, err error)

	// Audit Logs
	GetAuditLogsForUpload(ctx context.Context, uploadID int) (_ []types.UploadLog, err error)
	DeleteOldAuditLogs(ctx context.Context, maxAge time.Duration, now time.Time) (numRecordsScanned, numRecordsAltered int, err error)

	// Dependencies
	InsertDependencySyncingJob(ctx context.Context, uploadID int) (jobID int, err error)

	// Workerutil
	WorkerutilStore(observationCtx *observation.Context) dbworkerstore.Store[types.Upload]

	ReconcileCandidates(ctx context.Context, batchSize int) (_ []int, err error)

	ReindexUploads(ctx context.Context, opts shared.ReindexUploadsOptions) error
	ReindexUploadByID(ctx context.Context, id int) error

	GetRecentIndexesSummary(ctx context.Context, repositoryID int) (summaries []autoindexingshared.IndexesWithRepositoryNamespace, err error)
	NumRepositoriesWithCodeIntelligence(ctx context.Context) (int, error)
	RepositoryIDsWithErrors(ctx context.Context, offset, limit int) (_ []autoindexingshared.RepositoryWithCount, totalCount int, err error)
}

// store manages the database operations for uploads.
type store struct {
	logger     logger.Logger
	db         *basestore.Store
	operations *operations
}

// New returns a new uploads store.
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
