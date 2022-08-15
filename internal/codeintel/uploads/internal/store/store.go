package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"
	logger "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Store provides the interface for uploads storage.
type Store interface {
	// Not in use yet.
	List(ctx context.Context, opts ListOpts) (uploads []shared.Upload, err error)

	// Commits
	GetCommitsVisibleToUpload(ctx context.Context, uploadID, limit int, token *string) (_ []string, nextToken *string, err error)
	GetOldestCommitDate(ctx context.Context, repositoryID int) (time.Time, bool, error)
	GetStaleSourcedCommits(ctx context.Context, minimumTimeSinceLastCheck time.Duration, limit int, now time.Time) (_ []shared.SourcedCommits, err error)
	UpdateSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (uploadsUpdated int, err error)
	DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, maximumCommitLag time.Duration, now time.Time) (uploadsUpdated int, uploadsDeleted int, err error)
	HasCommit(ctx context.Context, repositoryID int, commit string) (_ bool, err error)

	// Repositories
	GetRepositoriesMaxStaleAge(ctx context.Context) (_ time.Duration, err error)
	SetRepositoryAsDirty(ctx context.Context, repositoryID int) (err error)
	GetDirtyRepositories(ctx context.Context) (_ map[int]int, err error)
	RepoName(ctx context.Context, repositoryID int) (_ string, err error)              // TODO(numbers88s): renaming this after I remove dbStore from gitserver init.
	RepoNames(ctx context.Context, repositoryIDs ...int) (_ map[int]string, err error) // TODO(numbers88s): renaming this after I remove dbStore from gitserver init.
	SetRepositoriesForRetentionScan(ctx context.Context, processDelay time.Duration, limit int) (_ []int, err error)
	HasRepository(ctx context.Context, repositoryID int) (_ bool, err error)

	// Uploads
	GetUploads(ctx context.Context, opts shared.GetUploadsOptions) (_ []shared.Upload, _ int, err error)
	GetUploadIDsWithReferences(ctx context.Context, orderedMonikers []precise.QualifiedMonikerData, ignoreIDs []int, repositoryID int, commit string, limit int, offset int, trace observation.TraceLogger) (ids []int, recordsScanned int, totalCount int, err error)
	GetVisibleUploadsMatchingMonikers(ctx context.Context, repositoryID int, commit string, orderedMonikers []precise.QualifiedMonikerData, limit, offset int) (_ shared.PackageReferenceScanner, _ int, err error)
	UpdateUploadsVisibleToCommits(ctx context.Context, repositoryID int, graph *gitdomain.CommitGraph, refDescriptions map[string][]gitdomain.RefDescription, maxAgeForNonStaleBranches, maxAgeForNonStaleTags time.Duration, dirtyToken int, now time.Time) error
	UpdateUploadRetention(ctx context.Context, protectedIDs, expiredIDs []int) (err error)
	UpdateUploadsReferenceCounts(ctx context.Context, ids []int, dependencyUpdateType shared.DependencyReferenceCountUpdateType) (updated int, err error)
	SoftDeleteExpiredUploads(ctx context.Context) (int, error)
	HardDeleteUploadsByIDs(ctx context.Context, ids ...int) error
	DeleteUploadsStuckUploading(ctx context.Context, uploadedBefore time.Time) (_ int, err error)
	DeleteUploadsWithoutRepository(ctx context.Context, now time.Time) (_ map[int]int, err error)

	// Dumps
	FindClosestDumps(ctx context.Context, repositoryID int, commit, path string, rootMustEnclosePath bool, indexer string) (_ []shared.Dump, err error)
	FindClosestDumpsFromGraphFragment(ctx context.Context, repositoryID int, commit, path string, rootMustEnclosePath bool, indexer string, commitGraph *gitdomain.CommitGraph) (_ []shared.Dump, err error)
	GetDumpsWithDefinitionsForMonikers(ctx context.Context, monikers []precise.QualifiedMonikerData) (_ []shared.Dump, err error)
	GetDumpsByIDs(ctx context.Context, ids []int) (_ []shared.Dump, err error)

	// Packages
	UpdatePackages(ctx context.Context, dumpID int, packages []precise.Package) (err error)

	// References
	UpdatePackageReferences(ctx context.Context, dumpID int, references []precise.PackageReference) (err error)

	// Audit Logs
	DeleteOldAuditLogs(ctx context.Context, maxAge time.Duration, now time.Time) (count int, err error)
}

// store manages the database operations for uploads.
type store struct {
	logger     logger.Logger
	db         *basestore.Store
	operations *operations
}

// New returns a new uploads store.
func New(db database.DB, observationContext *observation.Context) Store {
	return &store{
		logger:     logger.Scoped("uploads.store", ""),
		db:         basestore.NewWithHandle(db.Handle()),
		operations: newOperations(observationContext),
	}
}

// ListOpts specifies options for listing uploads.
type ListOpts struct {
	Limit int
}

// List returns a list of uploads.
func (s *store) List(ctx context.Context, opts ListOpts) (uploads []shared.Upload, err error) {
	ctx, _, endObservation := s.operations.list.With(ctx, &err, observation.Args{})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("numUploads", len(uploads)),
		}})
	}()

	// This is only a stub and will be replaced or significantly modified
	// in https://github.com/sourcegraph/sourcegraph/issues/33375
	_, _ = scanUploads(s.db.Query(ctx, sqlf.Sprintf(listQuery, opts.Limit)))
	return nil, errors.Newf("unimplemented: uploads.store.List")
}

const listQuery = `
-- source: internal/codeintel/uploads/internal/store/store.go:List
SELECT id FROM TODO
LIMIT %s
`
