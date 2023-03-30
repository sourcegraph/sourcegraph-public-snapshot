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

	GetIndexes(ctx context.Context, opts shared.GetIndexesOptions) ([]types.Index, int, error)
	GetIndexByID(ctx context.Context, id int) (types.Index, bool, error)
	GetIndexesByIDs(ctx context.Context, ids ...int) ([]types.Index, error)
	DeleteIndexByID(ctx context.Context, id int) (bool, error)
	DeleteIndexes(ctx context.Context, opts shared.DeleteIndexesOptions) error
	ReindexIndexByID(ctx context.Context, id int) error
	ReindexIndexes(ctx context.Context, opts shared.ReindexIndexesOptions) error

	// Commits
	GetCommitsVisibleToUpload(ctx context.Context, uploadID, limit int, token *string) ([]string, *string, error)
	GetOldestCommitDate(ctx context.Context, repositoryID int) (time.Time, bool, error)
	ProcessSourcedCommits(
		ctx context.Context,
		minimumTimeSinceLastCheck time.Duration,
		commitResolverMaximumCommitLag time.Duration,
		limit int,
		f func(ctx context.Context, repositoryID int, repositoryName, commit string) (bool, error),
		now time.Time,
	) (int, int, error)
	GetCommitGraphMetadata(ctx context.Context, repositoryID int) (stale bool, updatedAt *time.Time, _ error)
	HasCommit(ctx context.Context, repositoryID int, commit string) (bool, error)

	// Repositories
	GetRepositoriesMaxStaleAge(ctx context.Context) (time.Duration, error)
	SetRepositoryAsDirty(ctx context.Context, repositoryID int) error
	GetDirtyRepositories(ctx context.Context) ([]shared.DirtyRepository, error)
	SetRepositoriesForRetentionScan(ctx context.Context, processDelay time.Duration, limit int) ([]int, error)
	HasRepository(ctx context.Context, repositoryID int) (bool, error)

	// Uploads
	GetIndexers(ctx context.Context, opts shared.GetIndexersOptions) ([]string, error)
	GetUploads(ctx context.Context, opts shared.GetUploadsOptions) ([]types.Upload, int, error)
	GetUploadByID(ctx context.Context, id int) (types.Upload, bool, error)
	GetUploadsByIDs(ctx context.Context, ids ...int) ([]types.Upload, error)
	GetUploadsByIDsAllowDeleted(ctx context.Context, ids ...int) ([]types.Upload, error)
	GetUploadIDsWithReferences(ctx context.Context, orderedMonikers []precise.QualifiedMonikerData, ignoreIDs []int, repositoryID int, commit string, limit int, offset int, trace observation.TraceLogger) ([]int, int, int, error)
	GetVisibleUploadsMatchingMonikers(ctx context.Context, repositoryID int, commit string, orderedMonikers []precise.QualifiedMonikerData, limit, offset int) (shared.PackageReferenceScanner, int, error)
	GetRecentUploadsSummary(ctx context.Context, repositoryID int) ([]shared.UploadsWithRepositoryNamespace, error)
	GetLastUploadRetentionScanForRepository(ctx context.Context, repositoryID int) (*time.Time, error)
	UpdateUploadsVisibleToCommits(ctx context.Context, repositoryID int, graph *gitdomain.CommitGraph, refDescriptions map[string][]gitdomain.RefDescription, maxAgeForNonStaleBranches, maxAgeForNonStaleTags time.Duration, dirtyToken int, now time.Time) error
	UpdateUploadRetention(ctx context.Context, protectedIDs, expiredIDs []int) error
	SourcedCommitsWithoutCommittedAt(ctx context.Context, batchSize int) ([]shared.SourcedCommits, error)
	UpdateCommittedAt(ctx context.Context, repositoryID int, commit, commitDateString string) error
	SoftDeleteExpiredUploads(ctx context.Context, batchSize int) (int, int, error)
	SoftDeleteExpiredUploadsViaTraversal(ctx context.Context, maxTraversal int) (int, int, error)
	HardDeleteUploadsByIDs(ctx context.Context, ids ...int) error
	DeleteUploadsStuckUploading(ctx context.Context, uploadedBefore time.Time) (int, int, error)
	DeleteUploadsWithoutRepository(ctx context.Context, now time.Time) (int, int, error)
	DeleteUploadByID(ctx context.Context, id int) (bool, error)
	DeleteUploads(ctx context.Context, opts shared.DeleteUploadsOptions) error

	// Uploads (uploading)
	InsertUpload(ctx context.Context, upload types.Upload) (int, error)
	AddUploadPart(ctx context.Context, uploadID, partIndex int) error
	MarkQueued(ctx context.Context, id int, uploadSize *int64) error
	MarkFailed(ctx context.Context, id int, reason string) error

	// Dumps
	FindClosestDumps(ctx context.Context, repositoryID int, commit, path string, rootMustEnclosePath bool, indexer string) ([]types.Dump, error)
	FindClosestDumpsFromGraphFragment(ctx context.Context, repositoryID int, commit, path string, rootMustEnclosePath bool, indexer string, commitGraph *gitdomain.CommitGraph) ([]types.Dump, error)
	GetDumpsWithDefinitionsForMonikers(ctx context.Context, monikers []precise.QualifiedMonikerData) ([]types.Dump, error)
	GetDumpsByIDs(ctx context.Context, ids []int) ([]types.Dump, error)
	DeleteOverlappingDumps(ctx context.Context, repositoryID int, commit, root, indexer string) error

	// Packages
	UpdatePackages(ctx context.Context, dumpID int, packages []precise.Package) error

	// References
	UpdatePackageReferences(ctx context.Context, dumpID int, references []precise.PackageReference) error
	ReferencesForUpload(ctx context.Context, uploadID int) (shared.PackageReferenceScanner, error)

	// Audit Logs
	GetAuditLogsForUpload(ctx context.Context, uploadID int) ([]types.UploadLog, error)
	DeleteOldAuditLogs(ctx context.Context, maxAge time.Duration, now time.Time) (numRecordsScanned, numRecordsAltered int, _ error)

	// Dependencies
	InsertDependencySyncingJob(ctx context.Context, uploadID int) (int, error)

	// Workerutil
	WorkerutilStore(observationCtx *observation.Context) dbworkerstore.Store[types.Upload]

	ReconcileCandidates(ctx context.Context, batchSize int) ([]int, error)

	ReindexUploads(ctx context.Context, opts shared.ReindexUploadsOptions) error
	ReindexUploadByID(ctx context.Context, id int) error

	// Commits
	ProcessStaleSourcedCommits(
		ctx context.Context,
		minimumTimeSinceLastCheck time.Duration,
		commitResolverBatchSize int,
		commitResolverMaximumCommitLag time.Duration,
		shouldDelete func(ctx context.Context, repositoryID int, repositoryName, commit string) (bool, error),
	) (int, int, error)

	DeleteIndexesWithoutRepository(ctx context.Context, now time.Time) (int, int, error)

	ExpireFailedRecords(ctx context.Context, batchSize int, failedIndexMaxAge time.Duration, now time.Time) (int, int, error)
	GetRecentIndexesSummary(ctx context.Context, repositoryID int) ([]autoindexingshared.IndexesWithRepositoryNamespace, error)
	NumRepositoriesWithCodeIntelligence(ctx context.Context) (int, error)
	RepositoryIDsWithErrors(ctx context.Context, offset, limit int) ([]autoindexingshared.RepositoryWithCount, int, error)
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
