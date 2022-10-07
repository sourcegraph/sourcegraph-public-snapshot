package uploads

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	policiesEnterprise "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/enterprise"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (s *Service) NewUploadExpirer(
	interval time.Duration,
	repositoryProcessDelay time.Duration,
	repositoryBatchSize int,
	uploadProcessDelay time.Duration,
	uploadBatchSize int,
	commitBatchSize int,
	policyBatchSize int,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		return s.handleExpiredUploadsBatch(ctx, expirerConfig{
			repositoryProcessDelay: repositoryProcessDelay,
			repositoryBatchSize:    repositoryBatchSize,
			uploadProcessDelay:     uploadProcessDelay,
			uploadBatchSize:        uploadBatchSize,
			commitBatchSize:        commitBatchSize,
			policyBatchSize:        policyBatchSize,
		})
	}))
}

type expirerConfig struct {
	repositoryProcessDelay time.Duration
	repositoryBatchSize    int
	uploadProcessDelay     time.Duration
	uploadBatchSize        int
	commitBatchSize        int
	policyBatchSize        int
}

// handleExpiredUploadsBatch compares the age of upload records against the age of uploads
// protected by global and repository specific data retention policies.
//
// Uploads that are older than the protected retention age are marked as expired. Expired records with
// no dependents will be removed by the expiredUploadDeleter.
func (e *Service) handleExpiredUploadsBatch(ctx context.Context, cfg expirerConfig) (err error) {
	// Get the batch of repositories that we'll handle in this invocation of the periodic goroutine. This
	// set should contain repositories that have yet to be updated, or that have been updated least recently.
	// This allows us to update every repository reliably, even if it takes a long time to process through
	// the backlog. Note that this set of repositories require a fresh commit graph, so we're not trying to
	// process records that have been uploaded but the commits from which they are visible have yet to be
	// determined (and appearing as if they are visible to no commit).
	repositories, err := e.store.SetRepositoriesForRetentionScan(ctx, cfg.repositoryProcessDelay, cfg.repositoryBatchSize)
	if err != nil {
		return errors.Wrap(err, "uploadSvc.SelectRepositoriesForRetentionScan")
	}
	if len(repositories) == 0 {
		// All repositories updated recently enough
		return nil
	}

	now := timeutil.Now()

	for _, repositoryID := range repositories {
		if repositoryErr := e.handleRepository(ctx, repositoryID, cfg, now); repositoryErr != nil {
			if err == nil {
				err = repositoryErr
			} else {
				err = errors.Append(err, repositoryErr)
			}
		}
	}

	return err
}

func (e *Service) handleRepository(ctx context.Context, repositoryID int, cfg expirerConfig, now time.Time) error {
	e.expirationMetrics.numRepositoriesScanned.Inc()

	// Build a map from commits to the set of policies that affect them. Note that this map should
	// never be empty as we have multiple protected data retention policies on the global scope so
	// that all data visible from a tag or branch tip is protected for at least a short amount of
	// time after upload.
	commitMap, err := e.buildCommitMap(ctx, repositoryID, cfg, now)
	if err != nil {
		return err
	}

	// Mark the time after which all unprocessed uploads for this repository will not be touched.
	// This timestamp field is used as a rate limiting device so we do not busy-loop over the same
	// protected records in the background.
	//
	// This value should be assigned OUTSIDE of the following loop to prevent the case where the
	// upload process delay is shorter than the time it takes to process one batch of uploads. This
	// is obviously a mis-configuration, but one we can make a bit less catastrophic by not updating
	// this value in the loop.
	lastRetentionScanBefore := now.Add(-cfg.uploadProcessDelay)

	for {
		// Each record pulled back by this query will either have its expired flag or its last
		// retention scan timestamp updated by the following handleUploads call. This guarantees
		// that the loop will terminate naturally after the entire set of candidate uploads have
		// been seen and updated with a time necessarily greater than lastRetentionScanBefore.
		//
		// Additionally, we skip the set of uploads that have finished processing strictly after
		// the last update to the commit graph for that repository. This ensures we do not throw
		// out new uploads that would happen to be visible to no commits since they were never
		// installed into the commit graph.

		uploads, _, err := e.GetUploads(ctx, types.GetUploadsOptions{
			State:                   "completed",
			RepositoryID:            repositoryID,
			AllowExpired:            false,
			OldestFirst:             true,
			Limit:                   cfg.uploadBatchSize,
			LastRetentionScanBefore: &lastRetentionScanBefore,
			InCommitGraph:           true,
		})
		if err != nil || len(uploads) == 0 {
			return err
		}

		if err := e.handleUploads(ctx, commitMap, uploads, cfg, now); err != nil {
			// Note that we collect errors in the lop of the handleUploads call, but we will still terminate
			// this loop on any non-nil error from that function. This is required to prevent us from pullling
			// back the same set of failing records from the database in a tight loop.
			return err
		}
	}
}

// buildCommitMap will iterate the complete set of configuration policies that apply to a particular
// repository and build a map from commits to the policies that apply to them.
func (e *Service) buildCommitMap(ctx context.Context, repositoryID int, cfg expirerConfig, now time.Time) (map[string][]policiesEnterprise.PolicyMatch, error) {
	var (
		offset   int
		policies []types.ConfigurationPolicy
	)

	for {
		// Retrieve the complete set of configuration policies that affect data retention for this repository
		policyBatch, totalCount, err := e.policySvc.GetConfigurationPolicies(ctx, types.GetConfigurationPoliciesOptions{
			RepositoryID:     repositoryID,
			ForDataRetention: true,
			Limit:            cfg.policyBatchSize,
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
	return e.policyMatcher.CommitsDescribedByPolicy(ctx, repositoryID, policies, now)
}

func (e *Service) handleUploads(
	ctx context.Context,
	commitMap map[string][]policiesEnterprise.PolicyMatch,
	uploads []types.Upload,
	cfg expirerConfig,
	now time.Time,
) (err error) {
	// Categorize each upload as protected or expired
	var (
		protectedUploadIDs = make([]int, 0, len(uploads))
		expiredUploadIDs   = make([]int, 0, len(uploads))
	)

	for _, upload := range uploads {
		protected, checkErr := e.isUploadProtectedByPolicy(ctx, commitMap, upload, cfg, now)
		if checkErr != nil {
			if err == nil {
				err = checkErr
			} else {
				err = errors.Append(err, checkErr)
			}

			// Collect errors but not prevent other commits from being successfully processed. We'll leave the
			// ones that fail here alone to be re-checked the next time records for this repository are scanned.
			continue
		}

		if protected {
			protectedUploadIDs = append(protectedUploadIDs, upload.ID)
		} else {
			expiredUploadIDs = append(expiredUploadIDs, upload.ID)
		}
	}

	// Update the last data retention scan timestamp on the upload records with the given protected identifiers
	// (so that we do not re-select the same uploads on the next batch) and sets the expired field on the upload
	// records with the given expired identifiers so that the expiredUploadDeleter process can remove then once
	// they are no longer referenced.

	if updateErr := e.store.UpdateUploadRetention(ctx, protectedUploadIDs, expiredUploadIDs); updateErr != nil {
		if updateErr := errors.Wrap(err, "uploadSvc.UpdateUploadRetention"); err == nil {
			err = updateErr
		} else {
			err = errors.Append(err, updateErr)
		}
	}

	if count := len(expiredUploadIDs); count > 0 {
		e.logger.Info("Expiring codeintel uploads", log.Int("count", count))
		e.expirationMetrics.numUploadsExpired.Add(float64(count))
	}

	return err
}

func (e *Service) isUploadProtectedByPolicy(
	ctx context.Context,
	commitMap map[string][]policiesEnterprise.PolicyMatch,
	upload types.Upload,
	cfg expirerConfig,
	now time.Time,
) (bool, error) {
	e.expirationMetrics.numUploadsScanned.Inc()

	var token *string

	for first := true; first || token != nil; first = false {
		// Fetch the set of commits for which this upload can resolve code intelligence queries. This will necessarily
		// include the exact commit indicated by the upload, but may also provide best-effort code intelligence to
		// nearby commits.
		//
		// We need to consider all visible commits, as we may otherwise delete the uploads providing code intelligence
		// for  the tip of a branch between the time gitserver is updated and new the associated code intelligence index
		// is processed.
		//
		// We check the set of commits visible to an upload in batches as in some cases it can be very large; for
		// example, a single historic commit providing code intelligence for all descendants.
		commits, nextToken, err := e.GetCommitsVisibleToUpload(ctx, upload.ID, cfg.commitBatchSize, token)
		if err != nil {
			return false, errors.Wrap(err, "uploadSvc.CommitsVisibleToUpload")
		}
		token = nextToken

		e.expirationMetrics.numCommitsScanned.Add(float64(len(commits)))

		for _, commit := range commits {
			if policyMatches, ok := commitMap[commit]; ok {
				for _, policyMatch := range policyMatches {
					if policyMatch.PolicyDuration == nil || now.Sub(upload.UploadedAt) < *policyMatch.PolicyDuration {
						return true, nil
					}
				}
			}
		}
	}

	return false, nil
}
