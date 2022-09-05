package uploads

import (
	"context"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/opentracing/opentracing-go/log"

	logger "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	gitserverOptions "github.com/sourcegraph/sourcegraph/internal/gitserver"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ service = (*Service)(nil)

type service interface {
	// Not in use yet.
	Get(ctx context.Context, id int) (upload Upload, ok bool, err error)
	GetBatch(ctx context.Context, ids ...int) (uploads []Upload, err error)
	Enqueue(ctx context.Context, state UploadState, reader io.Reader) (err error)
	Delete(ctx context.Context, id int) (err error)
	UploadsVisibleToCommit(ctx context.Context, commit string) (uploads []Upload, err error)

	// Commits
	GetOldestCommitDate(ctx context.Context, repositoryID int) (time.Time, bool, error)
	GetCommitsVisibleToUpload(ctx context.Context, uploadID, limit int, token *string) (_ []string, nextToken *string, err error)
	GetStaleSourcedCommits(ctx context.Context, minimumTimeSinceLastCheck time.Duration, limit int, now time.Time) (_ []shared.SourcedCommits, err error)
	UpdateSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (uploadsUpdated int, err error)
	DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, maximumCommitLag time.Duration, now time.Time) (uploadsUpdated int, uploadsDeleted int, err error)

	// Repositories
	GetRepoName(ctx context.Context, repositoryID int) (_ string, err error)
	GetRepositoriesForIndexScan(ctx context.Context, table, column string, processDelay time.Duration, allowGlobalPolicies bool, repositoryMatchLimit *int, limit int, now time.Time) (_ []int, err error)
	GetRepositoriesMaxStaleAge(ctx context.Context) (_ time.Duration, err error)
	GetDirtyRepositories(ctx context.Context) (_ map[int]int, err error)
	SetRepositoryAsDirty(ctx context.Context, repositoryID int) (err error)
	UpdateDirtyRepositories(ctx context.Context, maxAgeForNonStaleBranches time.Duration, maxAgeForNonStaleTags time.Duration) (err error)
	SetRepositoriesForRetentionScan(ctx context.Context, processDelay time.Duration, limit int) (_ []int, err error)

	// Uploads
	GetUploads(ctx context.Context, opts shared.GetUploadsOptions) (uploads []shared.Upload, totalCount int, err error)
	GetUploadIDsWithReferences(ctx context.Context, orderedMonikers []precise.QualifiedMonikerData, ignoreIDs []int, repositoryID int, commit string, limit int, offset int) (ids []int, recordsScanned int, totalCount int, err error)
	UpdateUploadsVisibleToCommits(ctx context.Context, repositoryID int, graph *gitdomain.CommitGraph, refDescriptions map[string][]gitdomain.RefDescription, maxAgeForNonStaleBranches, maxAgeForNonStaleTags time.Duration, dirtyToken int, now time.Time) error
	UpdateUploadRetention(ctx context.Context, protectedIDs, expiredIDs []int) (err error)
	UpdateUploadsReferenceCounts(ctx context.Context, ids []int, dependencyUpdateType shared.DependencyReferenceCountUpdateType) (updated int, err error)
	SoftDeleteExpiredUploads(ctx context.Context) (count int, err error)
	HardDeleteExpiredUploads(ctx context.Context) (count int, err error)
	DeleteUploadsStuckUploading(ctx context.Context, uploadedBefore time.Time) (_ int, err error)
	DeleteUploadsWithoutRepository(ctx context.Context, now time.Time) (_ map[int]int, err error)
	InferClosestUploads(ctx context.Context, repositoryID int, commit, path string, exactPath bool, indexer string) ([]shared.Dump, error)

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

type Service struct {
	store           store.Store
	lsifstore       lsifstore.LsifStore
	gitserverClient shared.GitserverClient
	locker          Locker
	logger          logger.Logger
	operations      *operations
}

func newService(store store.Store, lsifstore lsifstore.LsifStore, gsc shared.GitserverClient, locker Locker, observationCtx *observation.Context) *Service {
	return &Service{
		store:           store,
		lsifstore:       lsifstore,
		gitserverClient: gsc,
		locker:          locker,
		logger:          logger.Scoped("uploads.service", ""),
		operations:      newOperations(observationCtx),
	}
}

type Upload = shared.Upload

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

func (s *Service) GetCommitsVisibleToUpload(ctx context.Context, uploadID, limit int, token *string) (_ []string, nextToken *string, err error) {
	ctx, _, endObservation := s.operations.getCommitsVisibleToUpload.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.GetCommitsVisibleToUpload(ctx, uploadID, limit, token)
}

func (s *Service) UploadsVisibleToCommit(ctx context.Context, commit string) (uploads []Upload, err error) {
	ctx, _, endObservation := s.operations.uploadsVisibleTo.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33375
	_ = ctx
	return nil, errors.Newf("unimplemented: uploads.UploadsVisibleToCommit")
}

func (s *Service) GetStaleSourcedCommits(ctx context.Context, minimumTimeSinceLastCheck time.Duration, limit int, now time.Time) (_ []shared.SourcedCommits, err error) {
	ctx, _, endObservation := s.operations.getStaleSourcedCommits.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("minimumTimeSinceLastCheck in ms", int(minimumTimeSinceLastCheck.Milliseconds())),
			log.Int("limit", limit),
			log.String("now", now.String()),
		},
	})
	defer endObservation(1, observation.Args{})

	return s.store.GetStaleSourcedCommits(ctx, minimumTimeSinceLastCheck, limit, now)
}

func (s *Service) UpdateSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (uploadsUpdated int, err error) {
	ctx, _, endObservation := s.operations.updateSourcedCommits.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("repositoryID", repositoryID), log.String("commit", commit), log.String("now", now.String())},
	})
	defer endObservation(1, observation.Args{})

	return s.store.UpdateSourcedCommits(ctx, repositoryID, commit, now)
}

func (s *Service) DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, maximumCommitLag time.Duration, now time.Time) (uploadsUpdated int, uploadsDeleted int, err error) {
	ctx, _, endObservation := s.operations.deleteSourcedCommits.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("repositoryID", repositoryID), log.String("commit", commit), log.Int("maximumCommitLag in ms", int(maximumCommitLag.Milliseconds())), log.String("now", now.String())},
	})
	defer endObservation(1, observation.Args{})

	return s.store.DeleteSourcedCommits(ctx, repositoryID, commit, maximumCommitLag, now)
}

func (s *Service) GetOldestCommitDate(ctx context.Context, repositoryID int) (time.Time, bool, error) {
	ctx, _, endObservation := s.operations.getOldestCommitDate.With(ctx, nil, observation.Args{
		LogFields: []log.Field{log.Int("repositoryID", repositoryID)},
	})
	defer endObservation(1, observation.Args{})

	return s.store.GetOldestCommitDate(ctx, repositoryID)
}

func (s *Service) SetRepositoryAsDirty(ctx context.Context, repositoryID int) (err error) {
	ctx, _, endObservation := s.operations.setRepositoryAsDirty.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("repositoryID", repositoryID)},
	})
	defer endObservation(1, observation.Args{})

	return s.store.SetRepositoryAsDirty(ctx, repositoryID)
}

func (s *Service) UpdateDirtyRepositories(ctx context.Context, maxAgeForNonStaleBranches time.Duration, maxAgeForNonStaleTags time.Duration) (err error) {
	ctx, _, endObservation := s.operations.updateDirtyRepositories.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("maxAgeForNonStaleBranches in ms", int(maxAgeForNonStaleBranches.Milliseconds())),
			log.Int("maxAgeForNonStaleTags in ms", int(maxAgeForNonStaleTags.Milliseconds())),
		},
	})
	defer endObservation(1, observation.Args{})

	repositoryIDs, err := s.GetDirtyRepositories(ctx)
	if err != nil {
		return errors.Wrap(err, "uploadSvc.DirtyRepositories")
	}

	var updateErr error
	for repositoryID, dirtyFlag := range repositoryIDs {
		if err := s.lockAndUpdateUploadsVisibleToCommits(ctx, repositoryID, dirtyFlag, maxAgeForNonStaleBranches, maxAgeForNonStaleTags); err != nil {
			if updateErr == nil {
				updateErr = err
			} else {
				updateErr = errors.Append(updateErr, err)
			}
		}
	}

	return updateErr
}

// lockAndUpdateUploadsVisibleToCommits will call UpdateUploadsVisibleToCommits while holding an advisory lock to give exclusive access to the
// update procedure for this repository. If the lock is already held, this method will simply do nothing.
func (s *Service) lockAndUpdateUploadsVisibleToCommits(ctx context.Context, repositoryID, dirtyToken int, maxAgeForNonStaleBranches time.Duration, maxAgeForNonStaleTags time.Duration) (err error) {
	ctx, trace, endObservation := s.operations.updateUploadsVisibleToCommits.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", repositoryID),
			log.Int("dirtyToken", dirtyToken),
		},
	})
	defer endObservation(1, observation.Args{})

	ok, unlock, err := s.locker.Lock(ctx, int32(repositoryID), false)
	if err != nil || !ok {
		return errors.Wrap(err, "locker.Lock")
	}
	defer func() {
		err = unlock(err)
	}()

	// The following process pulls the commit graph for the given repository from gitserver, pulls the set of LSIF
	// upload objects for the given repository from Postgres, and correlates them into a visibility
	// graph. This graph is then upserted back into Postgres for use by find closest dumps queries.
	//
	// The user should supply a dirty token that is associated with the given repository so that
	// the repository can be unmarked as long as the repository is not marked as dirty again before
	// the update completes.

	// Construct a view of the git graph that we will later decorate with upload information.
	commitGraph, err := s.getCommitGraph(ctx, repositoryID)
	if err != nil {
		return err
	}
	trace.Log(log.Int("numCommitGraphKeys", len(commitGraph.Order())))

	refDescriptions, err := s.gitserverClient.RefDescriptions(ctx, repositoryID)
	if err != nil {
		return errors.Wrap(err, "gitserver.RefDescriptions")
	}
	trace.Log(log.Int("numRefDescriptions", len(refDescriptions)))

	// Decorate the commit graph with the set of processed uploads are visible from each commit,
	// then bulk update the denormalized view in Postgres. We call this with an empty graph as well
	// so that we end up clearing the stale data and bulk inserting nothing.
	if err := s.UpdateUploadsVisibleToCommits(ctx, repositoryID, commitGraph, refDescriptions, maxAgeForNonStaleBranches, maxAgeForNonStaleTags, dirtyToken, time.Time{}); err != nil {
		return errors.Wrap(err, "uploadSvc.UpdateUploadsVisibleToCommits")
	}

	return nil
}

// getCommitGraph builds a partial commit graph that includes the most recent commits on each branch
// extending back as as the date of the oldest commit for which we have a processed upload for this
// repository.
//
// This optimization is necessary as decorating the commit graph is an operation that scales with
// the size of both the git graph and the number of uploads (multiplicatively). For repositories with
// a very large number of commits or distinct roots (most monorepos) this is a necessary optimization.
//
// The number of commits pulled back here should not grow over time unless the repo is growing at an
// accelerating rate, as we routinely expire old information for active repositories in a janitor
// process.
func (s *Service) getCommitGraph(ctx context.Context, repositoryID int) (*gitdomain.CommitGraph, error) {
	commitDate, ok, err := s.GetOldestCommitDate(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	if !ok {
		// No uploads exist for this repository
		return gitdomain.ParseCommitGraph(nil), nil
	}

	// The --since flag for git log is exclusive, but we want to include the commit where the
	// oldest dump is defined. This flag only has second resolution, so we shouldn't be pulling
	// back any more data than we wanted.
	commitDate = commitDate.Add(-time.Second)

	commitGraph, err := s.gitserverClient.CommitGraph(ctx, repositoryID, gitserverOptions.CommitGraphOptions{
		AllRefs: true,
		Since:   &commitDate,
	})
	if err != nil {
		return nil, errors.Wrap(err, "gitserver.CommitGraph")
	}

	return commitGraph, nil
}

func (s *Service) SetRepositoriesForRetentionScan(ctx context.Context, processDelay time.Duration, limit int) (_ []int, err error) {
	ctx, _, endObservation := s.operations.setRepositoriesForRetentionScan.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("processDelay in ms", int(processDelay.Milliseconds())), log.Int("limit", limit)},
	})
	defer endObservation(1, observation.Args{})

	return s.store.SetRepositoriesForRetentionScan(ctx, processDelay, limit)
}

func (s *Service) GetRepoName(ctx context.Context, repositoryID int) (_ string, err error) {
	ctx, _, endObservation := s.operations.getRepoName.With(ctx, &err, observation.Args{LogFields: []log.Field{log.Int("repositoryID", repositoryID)}})
	defer endObservation(1, observation.Args{})

	return s.store.RepoName(ctx, repositoryID)
}

func (s *Service) GetRepositoriesForIndexScan(ctx context.Context, table, column string, processDelay time.Duration, allowGlobalPolicies bool, repositoryMatchLimit *int, limit int, now time.Time) (_ []int, err error) {
	ctx, _, endObservation := s.operations.getRepositoriesForIndexScan.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.String("table", table),
			log.String("column", column),
			log.Int("processDelay in ms", int(processDelay.Milliseconds())),
			log.Bool("allowGlobalPolicies", allowGlobalPolicies),
			log.Int("limit", limit),
			log.String("now", now.String()),
		},
	})
	defer endObservation(1, observation.Args{})

	return s.store.GetRepositoriesForIndexScan(ctx, table, column, processDelay, allowGlobalPolicies, repositoryMatchLimit, limit, now)
}

func (s *Service) GetRepositoriesMaxStaleAge(ctx context.Context) (_ time.Duration, err error) {
	ctx, _, endObservation := s.operations.getRepositoriesMaxStaleAge.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.GetRepositoriesMaxStaleAge(ctx)
}

func (s *Service) GetDirtyRepositories(ctx context.Context) (_ map[int]int, err error) {
	ctx, _, endObservation := s.operations.getDirtyRepositories.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.GetDirtyRepositories(ctx)
}

func (s *Service) GetUploads(ctx context.Context, opts shared.GetUploadsOptions) (uploads []Upload, totalCount int, err error) {
	ctx, _, endObservation := s.operations.getUploads.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("repositoryID", opts.RepositoryID), log.String("state", opts.State), log.String("term", opts.Term)},
	})
	defer endObservation(1, observation.Args{})

	return s.store.GetUploads(ctx, opts)
}

func (s *Service) GetUploadIDsWithReferences(ctx context.Context, orderedMonikers []precise.QualifiedMonikerData, ignoreIDs []int, repositoryID int, commit string, limit int, offset int) (ids []int, recordsScanned int, totalCount int, err error) {
	ctx, trace, endObservation := s.operations.getVisibleUploadsMatchingMonikers.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", repositoryID),
			log.String("commit", commit),
			log.Int("limit", limit),
			log.Int("offset", offset),
			log.String("orderedMonikers", fmt.Sprintf("%v", orderedMonikers)),
			log.String("ignoreIDs", fmt.Sprintf("%v", ignoreIDs)),
		},
	})
	defer endObservation(1, observation.Args{})

	return s.store.GetUploadIDsWithReferences(ctx, orderedMonikers, ignoreIDs, repositoryID, commit, limit, offset, trace)
}

func (s *Service) UpdateUploadsVisibleToCommits(ctx context.Context, repositoryID int, graph *gitdomain.CommitGraph, refDescriptions map[string][]gitdomain.RefDescription, maxAgeForNonStaleBranches, maxAgeForNonStaleTags time.Duration, dirtyToken int, now time.Time) (err error) {
	ctx, _, endObservation := s.operations.updateUploadsVisibleToCommits.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", repositoryID),
			log.String("graph", fmt.Sprintf("%v", graph)),
			log.String("refDescriptions", fmt.Sprintf("%v", refDescriptions)),
			log.String("maxAgeForNonStaleBranches", maxAgeForNonStaleBranches.String()),
			log.String("maxAgeForNonStaleTags", maxAgeForNonStaleTags.String()),
			log.Int("dirtyToken", dirtyToken),
			log.String("now", now.String()),
		},
	})
	defer endObservation(1, observation.Args{})

	return s.store.UpdateUploadsVisibleToCommits(ctx, repositoryID, graph, refDescriptions, maxAgeForNonStaleBranches, maxAgeForNonStaleTags, dirtyToken, now)
}

func (s *Service) UpdateUploadRetention(ctx context.Context, protectedIDs, expiredIDs []int) (err error) {
	ctx, _, endObservation := s.operations.updateUploadRetention.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.String("protectedIDs", fmt.Sprintf("%v", protectedIDs)), log.String("expiredIDs", fmt.Sprintf("%v", expiredIDs))},
	})
	defer endObservation(1, observation.Args{})

	return s.store.UpdateUploadRetention(ctx, protectedIDs, expiredIDs)
}

func (s *Service) BackfillReferenceCountBatch(ctx context.Context, batchSize int) (err error) {
	ctx, _, endObservation := s.operations.backfillReferenceCountBatch.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("batchSize", batchSize)},
	})
	defer endObservation(1, observation.Args{})

	return s.store.BackfillReferenceCountBatch(ctx, batchSize)
}

func (s *Service) UpdateUploadsReferenceCounts(ctx context.Context, ids []int, dependencyUpdateType shared.DependencyReferenceCountUpdateType) (updated int, err error) {
	ctx, _, endObservation := s.operations.updateUploadsReferenceCounts.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.String("ids", fmt.Sprintf("%v", ids))},
	})
	defer endObservation(1, observation.Args{})

	return s.store.UpdateUploadsReferenceCounts(ctx, ids, dependencyUpdateType)
}

func (s *Service) SoftDeleteExpiredUploads(ctx context.Context) (count int, err error) {
	ctx, _, endObservation := s.operations.softDeleteExpiredUploads.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.SoftDeleteExpiredUploads(ctx)
}

func (s *Service) DeleteUploadsStuckUploading(ctx context.Context, uploadedBefore time.Time) (_ int, err error) {
	ctx, _, endObservation := s.operations.deleteUploadsStuckUploading.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.String("uploadedBefore", uploadedBefore.String())},
	})
	defer endObservation(1, observation.Args{})

	return s.store.DeleteUploadsStuckUploading(ctx, uploadedBefore)
}

func (s *Service) DeleteUploadsWithoutRepository(ctx context.Context, now time.Time) (_ map[int]int, err error) {
	ctx, _, endObservation := s.operations.deleteUploadsWithoutRepository.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.String("now", now.String())},
	})
	defer endObservation(1, observation.Args{})

	return s.store.DeleteUploadsWithoutRepository(ctx, now)
}

// numAncestors is the number of ancestors to query from gitserver when trying to find the closest
// ancestor we have data for. Setting this value too low (relative to a repository's commit rate)
// will cause requests for an unknown commit return too few results; setting this value too high
// will raise the latency of requests for an unknown commit.
//
// TODO(efritz) - make adjustable via site configuration
const numAncestors = 100

// inferClosestUploads will return the set of visible uploads for the given commit. If this commit is
// newer than our last refresh of the lsif_nearest_uploads table for this repository, then we will mark
// the repository as dirty and quickly approximate the correct set of visible uploads.
//
// Because updating the entire commit graph is a blocking, expensive, and lock-guarded process, we  want
// to only do that in the background and do something chearp in latency-sensitive paths. To construct an
// approximate result, we query gitserver for a (relatively small) set of ancestors for the given commit,
// correlate that with the upload data we have for those commits, and re-run the visibility algorithm over
// the graph. This will not always produce the full set of visible commits - some responses may not contain
// all results while a subsequent request made after the lsif_nearest_uploads has been updated to include
// this commit will.
//
func (s *Service) InferClosestUploads(ctx context.Context, repositoryID int, commit, path string, exactPath bool, indexer string) (_ []shared.Dump, err error) {
	ctx, _, endObservation := s.operations.inferClosestUploads.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("repositoryID", repositoryID), log.String("commit", commit), log.String("path", path), log.Bool("exactPath", exactPath), log.String("indexer", indexer)},
	})
	defer endObservation(1, observation.Args{})

	// The parameters exactPath and rootMustEnclosePath align here: if we're looking for dumps
	// that can answer queries for a directory (e.g. diagnostics), we want any dump that happens
	// to intersect the target directory. If we're looking for dumps that can answer queries for
	// a single file, then we need a dump with a root that properly encloses that file.
	if dumps, err := s.store.FindClosestDumps(ctx, repositoryID, commit, path, exactPath, indexer); err != nil {
		return nil, errors.Wrap(err, "store.FindClosestDumps")
	} else if len(dumps) != 0 {
		return dumps, nil
	}

	// Repository has no LSIF data at all
	if repositoryExists, err := s.store.HasRepository(ctx, repositoryID); err != nil {
		return nil, errors.Wrap(err, "dbstore.HasRepository")
	} else if !repositoryExists {
		return nil, nil
	}

	// Commit is known and the empty dumps list explicitly means nothing is visible
	if commitExists, err := s.store.HasCommit(ctx, repositoryID, commit); err != nil {
		return nil, errors.Wrap(err, "dbstore.HasCommit")
	} else if commitExists {
		return nil, nil
	}

	// Otherwise, the repository has LSIF data but we don't know about the commit. This commit
	// is probably newer than our last upload. Pull back a portion of the updated commit graph
	// and try to link it with what we have in the database. Then mark the repository's commit
	// graph as dirty so it's updated for subsequent requests.

	graph, err := s.gitserverClient.CommitGraph(ctx, repositoryID, gitserver.CommitGraphOptions{
		Commit: commit,
		Limit:  numAncestors,
	})
	if err != nil {
		return nil, errors.Wrap(err, "gitserverClient.CommitGraph")
	}

	dumps, err := s.store.FindClosestDumpsFromGraphFragment(ctx, repositoryID, commit, path, exactPath, indexer, graph)
	if err != nil {
		return nil, errors.Wrap(err, "dbstore.FindClosestDumpsFromGraphFragment")
	}

	if err := s.store.SetRepositoryAsDirty(ctx, repositoryID); err != nil {
		return nil, errors.Wrap(err, "dbstore.MarkRepositoryAsDirty")
	}

	return dumps, nil
}

func (s *Service) FindClosestDumps(ctx context.Context, repositoryID int, commit, path string, rootMustEnclosePath bool, indexer string) (_ []shared.Dump, err error) {
	ctx, _, endObservation := s.operations.findClosestDumps.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", repositoryID), log.String("commit", commit), log.String("path", path),
			log.Bool("rootMustEnclosePath", rootMustEnclosePath), log.String("indexer", indexer),
		},
	})
	defer endObservation(1, observation.Args{})

	return s.store.FindClosestDumps(ctx, repositoryID, commit, path, rootMustEnclosePath, indexer)
}

func (s *Service) FindClosestDumpsFromGraphFragment(ctx context.Context, repositoryID int, commit, path string, rootMustEnclosePath bool, indexer string, commitGraph *gitdomain.CommitGraph) (_ []shared.Dump, err error) {
	ctx, _, endObservation := s.operations.findClosestDumpsFromGraphFragment.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", repositoryID), log.String("commit", commit), log.String("path", path),
			log.Bool("rootMustEnclosePath", rootMustEnclosePath), log.String("indexer", indexer),
		},
	})
	defer endObservation(1, observation.Args{})

	return s.store.FindClosestDumpsFromGraphFragment(ctx, repositoryID, commit, path, rootMustEnclosePath, indexer, commitGraph)
}

func (s *Service) GetDumpsWithDefinitionsForMonikers(ctx context.Context, monikers []precise.QualifiedMonikerData) (_ []shared.Dump, err error) {
	ctx, _, endObservation := s.operations.getDumpsWithDefinitionsForMonikers.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.String("monikers", fmt.Sprintf("%v", monikers))},
	})
	defer endObservation(1, observation.Args{})

	return s.store.GetDumpsWithDefinitionsForMonikers(ctx, monikers)
}

func (s *Service) GetDumpsByIDs(ctx context.Context, ids []int) (_ []shared.Dump, err error) {
	ctx, _, endObservation := s.operations.getDumpsByIDs.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("total_ids", len(ids)), log.String("ids", fmt.Sprintf("%v", ids))},
	})
	defer endObservation(1, observation.Args{})

	return s.store.GetDumpsByIDs(ctx, ids)
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
		if err := s.lsifstore.DeleteLsifDataByUploadIds(ctx, ids...); err != nil {
			return 0, errors.Wrap(err, "lsifstore.Clear")
		}

		if err := s.store.HardDeleteUploadsByIDs(ctx, ids...); err != nil {
			return 0, errors.Wrap(err, "store.HardDeleteUploadsByIDs")
		}

		count += len(uploads)
		if count >= totalCount {
			break
		}
	}

	return count, nil
}

func (s *Service) UpdatePackages(ctx context.Context, dumpID int, packages []precise.Package) (err error) {
	ctx, _, endObservation := s.operations.updatePackages.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("dumpID", dumpID), log.String("packages", fmt.Sprintf("%v", packages))},
	})
	defer endObservation(1, observation.Args{})

	return s.store.UpdatePackages(ctx, dumpID, packages)
}

func (s *Service) UpdatePackageReferences(ctx context.Context, dumpID int, references []precise.PackageReference) (err error) {
	ctx, _, endObservation := s.operations.updatePackageReferences.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("dumpID", dumpID), log.String("references", fmt.Sprintf("%v", references))},
	})
	defer endObservation(1, observation.Args{})

	return s.store.UpdatePackageReferences(ctx, dumpID, references)
}

func (s *Service) DeleteOldAuditLogs(ctx context.Context, maxAge time.Duration, now time.Time) (count int, err error) {
	ctx, _, endObservation := s.operations.deleteOldAuditLogs.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.String("maxAge", maxAge.String()), log.String("now", now.String())},
	})
	defer endObservation(1, observation.Args{})

	return s.store.DeleteOldAuditLogs(ctx, maxAge, now)
}

// BackfillCommittedAtBatch calculates the committed_at value for a batch of upload records that do not have
// this value set. This method is used to backfill old upload records prior to this value being reliably set
// during processing.
func (s *Service) BackfillCommittedAtBatch(ctx context.Context, batchSize int) (err error) {
	ctx, _, endObservation := s.operations.backfillCommittedAtBatch.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("batchSize", batchSize),
	}})
	defer endObservation(1, observation.Args{})

	tx, err := s.store.Transact(ctx)
	defer func() {
		err = tx.Done(err)
	}()

	batch, err := tx.SourcedCommitsWithoutCommittedAt(ctx, batchSize)
	if err != nil {
		return errors.Wrap(err, "store.SourcedCommitsWithoutCommittedAt")
	}

	for _, sourcedCommits := range batch {
		for _, commit := range sourcedCommits.Commits {
			commitDateString, err := s.getCommitDate(ctx, sourcedCommits.RepositoryID, commit)
			if err != nil {
				return err
			}

			// Update commit date of all uploads attached to this this repository and commit
			if err := tx.UpdateCommittedAt(ctx, sourcedCommits.RepositoryID, commit, commitDateString); err != nil {
				return errors.Wrap(err, "store.UpdateCommittedAt")
			}
		}

		// Mark repository as dirty so the commit graph is recalculated with fresh data
		if err := tx.SetRepositoryAsDirty(ctx, sourcedCommits.RepositoryID); err != nil {
			return errors.Wrap(err, "store.SetRepositoryAsDirty")
		}
	}

	return nil
}

func (s *Service) getCommitDate(ctx context.Context, repositoryID int, commit string) (string, error) {
	_, commitDate, revisionExists, err := s.gitserverClient.CommitDate(ctx, repositoryID, commit)
	if err != nil {
		return "", errors.Wrap(err, "gitserver.CommitDate")
	}

	var commitDateString string
	if revisionExists {
		commitDateString = commitDate.Format(time.RFC3339)
	} else {
		// Set a value here that we'll filter out on the query side so that we don't
		// reprocess the same failing batch infinitely. We could alternatively soft
		// delete the record, but it would be better to keep record deletion behavior
		// together in the same place (so we have unified metrics on that event).
		commitDateString = "-infinity"
	}

	return commitDateString, nil
}

func uploadIDs(uploads []shared.Upload) []int {
	ids := make([]int, 0, len(uploads))
	for i := range uploads {
		ids = append(ids, uploads[i].ID)
	}
	sort.Ints(ids)

	return ids
}
