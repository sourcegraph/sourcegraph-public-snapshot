package uploads

import (
	"context"
	"fmt"
	"time"

	"github.com/derision-test/glock"
	"github.com/opentracing/opentracing-go/log"
	logger "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	policiesEnterprise "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/enterprise"
	policiesshared "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/background"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	gitserverOptions "github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Service struct {
	store           store.Store
	repoStore       RepoStore
	workerutilStore dbworkerstore.Store
	lsifstore       lsifstore.LsifStore
	gitserverClient GitserverClient
	policySvc       PolicyService
	policyMatcher   PolicyMatcher
	locker          Locker
	backgroundJob   background.BackgroundJob
	logger          logger.Logger
	operations      *operations
	clock           glock.Clock
}

func newService(
	store store.Store,
	repoStore RepoStore,
	lsifstore lsifstore.LsifStore,
	gsc GitserverClient,
	policySvc PolicyService,
	policyMatcher PolicyMatcher,
	locker Locker,
	backgroundJob background.BackgroundJob,
	observationContext *observation.Context,
) *Service {
	workerutilStore := store.WorkerutilStore(observationContext)

	// TODO - move this to metric reporter?
	dbworker.InitPrometheusMetric(observationContext, workerutilStore, "codeintel", "upload", nil)

	return &Service{
		store:           store,
		repoStore:       repoStore,
		workerutilStore: workerutilStore,
		lsifstore:       lsifstore,
		gitserverClient: gsc,
		policySvc:       policySvc,
		backgroundJob:   backgroundJob,
		policyMatcher:   policyMatcher,
		locker:          locker,
		logger:          observationContext.Logger,
		operations:      newOperations(observationContext),
		clock:           glock.NewRealClock(),
	}
}

func GetBackgroundJob(s *Service) background.BackgroundJob {
	return s.backgroundJob
}

func (s *Service) GetWorkerutilStore() dbworkerstore.Store {
	return s.workerutilStore
}

func (s *Service) GetCommitsVisibleToUpload(ctx context.Context, uploadID, limit int, token *string) (_ []string, nextToken *string, err error) {
	ctx, _, endObservation := s.operations.getCommitsVisibleToUpload.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.GetCommitsVisibleToUpload(ctx, uploadID, limit, token)
}

func (s *Service) GetCommitGraphMetadata(ctx context.Context, repositoryID int) (stale bool, updatedAt *time.Time, err error) {
	ctx, _, endObservation := s.operations.getCommitGraphMetadata.With(ctx, &err, observation.Args{LogFields: []log.Field{log.Int("repositoryID", repositoryID)}})
	defer endObservation(1, observation.Args{})

	return s.store.GetCommitGraphMetadata(ctx, repositoryID)
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

// NOTE: Used by autoindexing (for some reason?)
func (s *Service) GetDirtyRepositories(ctx context.Context) (_ map[int]int, err error) {
	ctx, _, endObservation := s.operations.getDirtyRepositories.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.GetDirtyRepositories(ctx)
}

func (s *Service) GetUploads(ctx context.Context, opts shared.GetUploadsOptions) (uploads []types.Upload, totalCount int, err error) {
	ctx, _, endObservation := s.operations.getUploads.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("repositoryID", opts.RepositoryID), log.String("state", opts.State), log.String("term", opts.Term)},
	})
	defer endObservation(1, observation.Args{})

	return s.store.GetUploads(ctx, opts)
}

// TODO: Not being used in the resolver layer
func (s *Service) GetUploadByID(ctx context.Context, id int) (_ types.Upload, _ bool, err error) {
	ctx, _, endObservation := s.operations.getUploadByID.With(ctx, &err, observation.Args{LogFields: []log.Field{log.Int("id", id)}})
	defer endObservation(1, observation.Args{})

	return s.store.GetUploadByID(ctx, id)
}

func (s *Service) GetUploadsByIDs(ctx context.Context, ids ...int) (_ []types.Upload, err error) {
	ctx, _, endObservation := s.operations.getUploadsByIDs.With(ctx, &err, observation.Args{LogFields: []log.Field{log.String("ids", fmt.Sprintf("%v", ids))}})
	defer endObservation(1, observation.Args{})

	return s.store.GetUploadsByIDs(ctx, ids...)
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

func (s *Service) DeleteUploadByID(ctx context.Context, id int) (_ bool, err error) {
	ctx, _, endObservation := s.operations.deleteUploadByID.With(ctx, &err, observation.Args{LogFields: []log.Field{log.Int("id", id)}})
	defer endObservation(1, observation.Args{})

	return s.store.DeleteUploadByID(ctx, id)
}

func (s *Service) DeleteUploads(ctx context.Context, opts shared.DeleteUploadsOptions) (err error) {
	ctx, _, endObservation := s.operations.deleteUploadByID.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.DeleteUploads(ctx, opts)
}

func (s *Service) SetRepositoriesForRetentionScan(ctx context.Context, processDelay time.Duration, limit int) (_ []int, err error) {
	ctx, _, endObservation := s.operations.setRepositoriesForRetentionScan.With(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("processDelayInMs", int(processDelay.Milliseconds())),
			log.Int("limit", limit),
		},
	})
	defer endObservation(1, observation.Args{})

	return s.store.SetRepositoriesForRetentionScan(ctx, processDelay, limit)
}

func (s *Service) GetRepositoriesMaxStaleAge(ctx context.Context) (_ time.Duration, err error) {
	ctx, _, endObservation := s.operations.getRepositoriesMaxStaleAge.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.GetRepositoriesMaxStaleAge(ctx)
}

// buildCommitMap will iterate the complete set of configuration policies that apply to a particular
// repository and build a map from commits to the policies that apply to them.
func (s *Service) BuildCommitMap(ctx context.Context, repositoryID int, cfg background.ExpirerConfig, now time.Time) (map[string][]policiesEnterprise.PolicyMatch, error) {
	var (
		offset   int
		policies []types.ConfigurationPolicy
	)

	for {
		// Retrieve the complete set of configuration policies that affect data retention for this repository
		policyBatch, totalCount, err := s.policySvc.GetConfigurationPolicies(ctx, policiesshared.GetConfigurationPoliciesOptions{
			RepositoryID:     repositoryID,
			ForDataRetention: true,
			Limit:            cfg.PolicyBatchSize,
			Offset:           offset,
		})
		if err != nil {
			return nil, errors.Wrap(err, "policySvc.GetConfigurationPolicies")
		}

		offset += len(policyBatch)
		policies = append(policies, policyBatch...)

		if len(policyBatch) == 0 || offset >= totalCount {
			break
		}
	}

	// Get the set of commits within this repository that match a data retention policy
	return s.policyMatcher.CommitsDescribedByPolicy(ctx, repositoryID, policies, now)
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
func (s *Service) InferClosestUploads(ctx context.Context, repositoryID int, commit, path string, exactPath bool, indexer string) (_ []types.Dump, err error) {
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

func (s *Service) GetDumpsWithDefinitionsForMonikers(ctx context.Context, monikers []precise.QualifiedMonikerData) (_ []types.Dump, err error) {
	ctx, _, endObservation := s.operations.getDumpsWithDefinitionsForMonikers.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.String("monikers", fmt.Sprintf("%v", monikers))},
	})
	defer endObservation(1, observation.Args{})

	return s.store.GetDumpsWithDefinitionsForMonikers(ctx, monikers)
}

func (s *Service) GetDumpsByIDs(ctx context.Context, ids []int) (_ []types.Dump, err error) {
	ctx, _, endObservation := s.operations.getDumpsByIDs.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("total_ids", len(ids)), log.String("ids", fmt.Sprintf("%v", ids))},
	})
	defer endObservation(1, observation.Args{})

	return s.store.GetDumpsByIDs(ctx, ids)
}

func (s *Service) ReferencesForUpload(ctx context.Context, uploadID int) (_ shared.PackageReferenceScanner, err error) {
	ctx, _, endObservation := s.operations.referencesForUpload.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("uploadID", uploadID)},
	})
	defer endObservation(1, observation.Args{})

	return s.store.ReferencesForUpload(ctx, uploadID)
}

func (s *Service) GetAuditLogsForUpload(ctx context.Context, uploadID int) (_ []types.UploadLog, err error) {
	ctx, _, endObservation := s.operations.getAuditLogsForUpload.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("uploadID", uploadID)},
	})
	defer endObservation(1, observation.Args{})

	return s.store.GetAuditLogsForUpload(ctx, uploadID)
}

func (s *Service) GetUploadDocumentsForPath(ctx context.Context, bundleID int, pathPattern string) (_ []string, _ int, err error) {
	ctx, _, endObservation := s.operations.getUploadDocumentsForPath.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("bundleID", bundleID), log.String("pathPattern", pathPattern)},
	})
	defer endObservation(1, observation.Args{})

	return s.lsifstore.GetUploadDocumentsForPath(ctx, bundleID, pathPattern)
}

func (s *Service) GetRecentUploadsSummary(ctx context.Context, repositoryID int) (upload []shared.UploadsWithRepositoryNamespace, err error) {
	ctx, _, endObservation := s.operations.getRecentUploadsSummary.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("repositoryID", repositoryID)},
	})
	defer endObservation(1, observation.Args{})

	return s.store.GetRecentUploadsSummary(ctx, repositoryID)
}

func (s *Service) GetLastUploadRetentionScanForRepository(ctx context.Context, repositoryID int) (_ *time.Time, err error) {
	ctx, _, endObservation := s.operations.getLastUploadRetentionScanForRepository.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("repositoryID", repositoryID)},
	})
	defer endObservation(1, observation.Args{})

	return s.store.GetLastUploadRetentionScanForRepository(ctx, repositoryID)
}

func (s *Service) GetListTags(ctx context.Context, repo api.RepoName, commitObjs ...string) (_ []*gitdomain.Tag, err error) {
	ctx, _, endObservation := s.operations.getListTags.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.String("repo", string(repo)), log.String("commitObjs", fmt.Sprintf("%v", commitObjs))},
	})
	defer endObservation(1, observation.Args{})

	return s.gitserverClient.ListTags(ctx, repo, commitObjs...)
}

func (s *Service) DeleteOldAuditLogs(ctx context.Context, maxAge time.Duration, now time.Time) (count int, err error) {
	ctx, _, endObservation := s.operations.deleteOldAuditLogs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.DeleteOldAuditLogs(ctx, maxAge, now)
}

func (s *Service) HardDeleteUploadsByIDs(ctx context.Context, ids ...int) error {
	ctx, _, endObservation := s.operations.hardDeleteUploadsByIDs.With(ctx, nil, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.HardDeleteUploadsByIDs(ctx, ids...)
}

func (s *Service) DeleteLsifDataByUploadIds(ctx context.Context, bundleIDs ...int) (err error) {
	ctx, _, endObservation := s.operations.deleteLsifDataByUploadIds.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("bundleIDs", len(bundleIDs))},
	})
	defer endObservation(1, observation.Args{})

	return s.lsifstore.DeleteLsifDataByUploadIds(ctx, bundleIDs...)
}

func (s *Service) GetStaleSourcedCommits(ctx context.Context, minimumTimeSinceLastCheck time.Duration, limit int, now time.Time) (_ []shared.SourcedCommits, err error) {
	ctx, _, endObservation := s.operations.getStaleSourcedCommits.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.String("minimumTimeSinceLastCheck", minimumTimeSinceLastCheck.String()), log.Int("limit", limit), log.String("now", now.String())},
	})
	defer endObservation(1, observation.Args{})

	return s.store.GetStaleSourcedCommits(ctx, minimumTimeSinceLastCheck, limit, now)
}

func (s *Service) DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, maximumCommitLag time.Duration, now time.Time) (uploadsUpdated int, uploadsDeleted int, err error) {
	ctx, _, endObservation := s.operations.deleteSourcedCommits.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("repositoryID", repositoryID), log.String("commit", commit), log.String("maximumCommitLag", maximumCommitLag.String()), log.String("now", now.String())},
	})
	defer endObservation(1, observation.Args{})

	return s.store.DeleteSourcedCommits(ctx, repositoryID, commit, maximumCommitLag, now)
}

func (s *Service) UpdateSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (uploadsUpdated int, err error) {
	ctx, _, endObservation := s.operations.updateSourcedCommits.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("repositoryID", repositoryID), log.String("commit", commit), log.String("now", now.String())},
	})
	defer endObservation(1, observation.Args{})

	return s.store.UpdateSourcedCommits(ctx, repositoryID, commit, now)
}

func (s *Service) DeleteUploadsStuckUploading(ctx context.Context, uploadedBefore time.Time) (_ int, err error) {
	ctx, _, endObservation := s.operations.deleteUploadsStuckUploading.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.String("uploadedBefore", uploadedBefore.String())},
	})
	defer endObservation(1, observation.Args{})

	return s.store.DeleteUploadsStuckUploading(ctx, uploadedBefore)
}

func (s *Service) SoftDeleteExpiredUploads(ctx context.Context, batchSize int) (int, error) {
	ctx, _, endObservation := s.operations.softDeleteExpiredUploads.With(ctx, nil, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.SoftDeleteExpiredUploads(ctx, batchSize)
}

func (s *Service) SoftDeleteExpiredUploadsViaTraversal(ctx context.Context, maxTraversal int) (int, error) {
	ctx, _, endObservation := s.operations.softDeleteExpiredUploadsViaTraversal.With(ctx, nil, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.SoftDeleteExpiredUploadsViaTraversal(ctx, maxTraversal)
}

// BackfillCommittedAtBatch calculates the committed_at value for a batch of upload records that do not have
// this value set. This method is used to backfill old upload records prior to this value being reliably set
// during processing.
func (s *Service) BackfillCommittedAtBatch(ctx context.Context, batchSize int) (err error) {
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

func (s *Service) DeleteUploadsWithoutRepository(ctx context.Context, now time.Time) (_ map[int]int, err error) {
	ctx, _, endObservation := s.operations.deleteUploadsWithoutRepository.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.String("now", now.Format(time.RFC3339))},
	})
	defer endObservation(1, observation.Args{})

	return s.store.DeleteUploadsWithoutRepository(ctx, now)
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

// Handle periodically re-calculates the commit and upload visibility graph for repositories
// that are marked as dirty by the worker process. This is done out-of-band from the rest of
// the upload processing as it is likely that we are processing multiple uploads concurrently
// for the same repository and should not repeat the work since the last calculation performed
// will always be the one we want.

func (s *Service) UpdateAllDirtyCommitGraphs(ctx context.Context, maxAgeForNonStaleBranches time.Duration, maxAgeForNonStaleTags time.Duration) (err error) {
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
	if err := s.store.UpdateUploadsVisibleToCommits(ctx, repositoryID, commitGraph, refDescriptions, maxAgeForNonStaleBranches, maxAgeForNonStaleTags, dirtyToken, time.Time{}); err != nil {
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
	commitDate, ok, err := s.store.GetOldestCommitDate(ctx, repositoryID)
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
