package uploads

import (
	"context"
	"io"
	"sort"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ service = (*Service)(nil)

type service interface {
	// Not in use yet.
	List(ctx context.Context, opts ListOpts) (uploads []Upload, err error)
	Get(ctx context.Context, id int) (upload Upload, ok bool, err error)
	GetBatch(ctx context.Context, ids ...int) (uploads []Upload, err error)
	Enqueue(ctx context.Context, state UploadState, reader io.Reader) (err error)
	Delete(ctx context.Context, id int) (err error)
	CommitsVisibleToUpload(ctx context.Context, id int) (commits []string, err error)
	UploadsVisibleToCommit(ctx context.Context, commit string) (uploads []Upload, err error)

	// Commits
	StaleSourcedCommits(ctx context.Context, minimumTimeSinceLastCheck time.Duration, limit int, now time.Time) (_ []shared.SourcedCommits, err error)
	UpdateSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (uploadsUpdated int, err error)
	DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, maximumCommitLag time.Duration, now time.Time) (uploadsUpdated int, uploadsDeleted int, err error)

	// Uploads
	GetUploads(ctx context.Context, opts shared.GetUploadsOptions) (uploads []Upload, totalCount int, err error)
	UpdateUploadRetention(ctx context.Context, protectedIDs, expiredIDs []int) (err error)
	UpdateUploadsReferenceCounts(ctx context.Context, ids []int, dependencyUpdateType shared.DependencyReferenceCountUpdateType) (updated int, err error)
	SoftDeleteExpiredUploads(ctx context.Context) (count int, err error)
	HardDeleteExpiredUploads(ctx context.Context) (count int, err error)
	DeleteUploadsStuckUploading(ctx context.Context, uploadedBefore time.Time) (_ int, err error)
	DeleteUploadsWithoutRepository(ctx context.Context, now time.Time) (_ map[int]int, err error)

	// Repositories
	SetRepositoryAsDirty(ctx context.Context, repositoryID int, tx *basestore.Store) (err error)
	GetDirtyRepositories(ctx context.Context) (_ map[int]int, err error)

	// Packages
	UpdatePackages(ctx context.Context, dumpID int, packages []precise.Package) (err error)

	// References
	UpdatePackageReferences(ctx context.Context, dumpID int, references []precise.PackageReference) (err error)

	// Audit Logs
	DeleteOldAuditLogs(ctx context.Context, maxAge time.Duration, now time.Time) (count int, err error)
}

type Service struct {
	store      store.Store
	lsifstore  lsifstore.LsifStore
	operations *operations
}

func newService(store store.Store, lsifstore lsifstore.LsifStore, observationCtx *observation.Context) *Service {
	return &Service{
		store:      store,
		lsifstore:  lsifstore,
		operations: newOperations(observationCtx),
	}
}

type Upload = shared.Upload

type ListOpts struct {
	Limit int
}

func (s *Service) List(ctx context.Context, opts ListOpts) (uploads []Upload, err error) {
	ctx, _, endObservation := s.operations.list.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.List(ctx, store.ListOpts(opts))
}

func (s *Service) Get(ctx context.Context, id int) (upload Upload, ok bool, err error) {
	ctx, _, endObservation := s.operations.get.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33375
	_ = ctx
	return Upload{}, false, errors.Newf("unimplemented: uploads.Get")
}

func (s *Service) GetBatch(ctx context.Context, ids ...int) (uploads []Upload, err error) {
	ctx, _, endObservation := s.operations.getBatch.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33375
	_ = ctx
	return nil, errors.Newf("unimplemented: uploads.GetBatch")
}

type UploadState struct{}

func (s *Service) Enqueue(ctx context.Context, state UploadState, reader io.Reader) (err error) {
	ctx, _, endObservation := s.operations.enqueue.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33375
	_ = ctx
	return errors.Newf("unimplemented: uploads.Enqueue")
}

func (s *Service) Delete(ctx context.Context, id int) (err error) {
	ctx, _, endObservation := s.operations.delete.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33375
	_ = ctx
	return errors.Newf("unimplemented: uploads.Delete")
}

func (s *Service) CommitsVisibleToUpload(ctx context.Context, id int) (commits []string, err error) {
	ctx, _, endObservation := s.operations.commitsVisibleTo.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33375
	_ = ctx
	return nil, errors.Newf("unimplemented: uploads.CommitsVisibleToUpload")
}

func (s *Service) UploadsVisibleToCommit(ctx context.Context, commit string) (uploads []Upload, err error) {
	ctx, _, endObservation := s.operations.uploadsVisibleTo.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33375
	_ = ctx
	return nil, errors.Newf("unimplemented: uploads.UploadsVisibleToCommit")
}

func (s *Service) StaleSourcedCommits(ctx context.Context, minimumTimeSinceLastCheck time.Duration, limit int, now time.Time) (_ []shared.SourcedCommits, err error) {
	ctx, _, endObservation := s.operations.staleSourcedCommits.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.StaleSourcedCommits(ctx, minimumTimeSinceLastCheck, limit, now)
}

func (s *Service) UpdateSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (uploadsUpdated int, err error) {
	ctx, _, endObservation := s.operations.updateSourcedCommits.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.UpdateSourcedCommits(ctx, repositoryID, commit, now)
}

func (s *Service) DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, maximumCommitLag time.Duration, now time.Time) (uploadsUpdated int, uploadsDeleted int, err error) {
	ctx, _, endObservation := s.operations.deleteSourcedCommits.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.DeleteSourcedCommits(ctx, repositoryID, commit, maximumCommitLag, now)
}

func (s *Service) GetUploads(ctx context.Context, opts shared.GetUploadsOptions) (uploads []Upload, totalCount int, err error) {
	ctx, _, endObservation := s.operations.getUploads.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.GetUploads(ctx, opts)
}

func (s *Service) UpdateUploadRetention(ctx context.Context, protectedIDs, expiredIDs []int) (err error) {
	ctx, _, endObservation := s.operations.updateUploadRetention.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.UpdateUploadRetention(ctx, protectedIDs, expiredIDs)
}

func (s *Service) UpdateUploadsReferenceCounts(ctx context.Context, ids []int, dependencyUpdateType shared.DependencyReferenceCountUpdateType) (updated int, err error) {
	ctx, _, endObservation := s.operations.updateUploadsReferenceCounts.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.UpdateUploadsReferenceCounts(ctx, ids, dependencyUpdateType)
}

func (s *Service) SoftDeleteExpiredUploads(ctx context.Context) (count int, err error) {
	ctx, _, endObservation := s.operations.softDeleteExpiredUploads.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.SoftDeleteExpiredUploads(ctx)
}

func (s *Service) DeleteUploadsStuckUploading(ctx context.Context, uploadedBefore time.Time) (_ int, err error) {
	ctx, _, endObservation := s.operations.deleteUploadsStuckUploading.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.DeleteUploadsStuckUploading(ctx, uploadedBefore)
}

func (s *Service) DeleteUploadsWithoutRepository(ctx context.Context, now time.Time) (_ map[int]int, err error) {
	ctx, _, endObservation := s.operations.deleteUploadsWithoutRepository.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.DeleteUploadsWithoutRepository(ctx, now)
}

func (s *Service) HardDeleteExpiredUploads(ctx context.Context) (count int, err error) {
	ctx, _, endObservation := s.operations.hardDeleteUploads.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	const uploadsBatchSize = 100
	options := shared.GetUploadsOptions{
		State:            "deleted",
		Limit:            uploadsBatchSize,
		AllowExpired:     true,
		AllowDeletedRepo: true,
	}

	for {
		// Always request the first page of deleted uploads. If this is not
		// the first iteration of the loop, then the previous iteration has
		// deleted the records that composed the previous page, and the
		// previous "second" page is now the first page.
		uploads, totalCount, err := s.store.GetUploads(ctx, options)
		if err != nil {
			return 0, errors.Wrap(err, "store.GetUploads")
		}

		ids := uploadIDs(uploads)
		if err := s.lsifstore.Clear(ctx, ids...); err != nil {
			return 0, errors.Wrap(err, "lsifstore.Clear")
		}

		if err := s.store.HardDeleteUploadByID(ctx, ids...); err != nil {
			return 0, errors.Wrap(err, "store.HardDeleteUploadByID")
		}

		count += len(uploads)
		if count >= totalCount {
			break
		}
	}

	return count, nil
}

func (s *Service) SetRepositoryAsDirty(ctx context.Context, repositoryID int, tx *basestore.Store) (err error) {
	ctx, _, endObservation := s.operations.setRepositoryAsDirty.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.SetRepositoryAsDirty(ctx, repositoryID, tx)
}

func (s *Service) GetDirtyRepositories(ctx context.Context) (_ map[int]int, err error) {
	ctx, _, endObservation := s.operations.getDirtyRepositories.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.GetDirtyRepositories(ctx)
}

func (s *Service) UpdatePackages(ctx context.Context, dumpID int, packages []precise.Package) (err error) {
	ctx, _, endObservation := s.operations.updatePackages.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.UpdatePackages(ctx, dumpID, packages)
}

func (s *Service) UpdatePackageReferences(ctx context.Context, dumpID int, references []precise.PackageReference) (err error) {
	ctx, _, endObservation := s.operations.updatePackageReferences.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.UpdatePackageReferences(ctx, dumpID, references)
}

func (s *Service) DeleteOldAuditLogs(ctx context.Context, maxAge time.Duration, now time.Time) (count int, err error) {
	ctx, _, endObservation := s.operations.deleteOldAuditLogs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.DeleteOldAuditLogs(ctx, maxAge, now)
}

func uploadIDs(uploads []shared.Upload) []int {
	ids := make([]int, 0, len(uploads))
	for i := range uploads {
		ids = append(ids, uploads[i].ID)
	}
	sort.Ints(ids)

	return ids
}
