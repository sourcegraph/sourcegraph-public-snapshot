package janitor

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"

	policies "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/enterprise"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type uploadExpirer struct {
	dbStore                DBStore
	policyMatcher          PolicyMatcher
	metrics                *metrics
	repositoryProcessDelay time.Duration
	repositoryBatchSize    int
	uploadProcessDelay     time.Duration
	uploadBatchSize        int
	policyBatchSize        int
	commitBatchSize        int
	branchesCacheMaxKeys   int
}

var (
	_ goroutine.Handler      = &uploadExpirer{}
	_ goroutine.ErrorHandler = &uploadExpirer{}
)

// NewUploadExpirer returns a background routine that periodically compares the age of upload records
// against the age of uploads protected by global and repository specific data retention policies.
//
// Uploads that are older than the protected retention age are marked as expired. Expired records with
// no dependents will be removed by the expiredUploadDeleter.
func NewUploadExpirer(
	dbStore DBStore,
	policyMatcher PolicyMatcher,
	repositoryProcessDelay time.Duration,
	repositoryBatchSize int,
	uploadProcessDelay time.Duration,
	uploadBatchSize int,
	policyBatchSize int,
	commitBatchSize int,
	branchesCacheMaxKeys int,
	interval time.Duration,
	metrics *metrics,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, &uploadExpirer{
		dbStore:                dbStore,
		policyMatcher:          policyMatcher,
		metrics:                metrics,
		repositoryProcessDelay: repositoryProcessDelay,
		repositoryBatchSize:    repositoryBatchSize,
		uploadProcessDelay:     uploadProcessDelay,
		uploadBatchSize:        uploadBatchSize,
		policyBatchSize:        policyBatchSize,
		commitBatchSize:        commitBatchSize,
		branchesCacheMaxKeys:   branchesCacheMaxKeys,
	})
}

func (e *uploadExpirer) Handle(ctx context.Context) (err error) {
	// Get the batch of repositories that we'll handle in this invocation of the periodic goroutine. This
	// set should contain repositories that have yet to be updated, or that have been updated least recently.
	// This allows us to update every repository reliably, even if it takes a long time to process through
	// the backlog. Note that this set of repositories require a fresh commit graph, so we're not trying to
	// process records that have been uploaded but the commits from which they are visible have yet to be
	// determined (and appearing as if they are visible to no commit).
	repositories, err := e.dbStore.SelectRepositoriesForRetentionScan(ctx, e.repositoryProcessDelay, e.repositoryBatchSize)
	if err != nil {
		return errors.Wrap(err, "dbstore.SelectRepositoriesForRetentionScan")
	}
	if len(repositories) == 0 {
		// All repositories updated recently enough
		return nil
	}

	now := timeutil.Now()

	for _, repositoryID := range repositories {
		if repositoryErr := e.handleRepository(ctx, repositoryID, now); repositoryErr != nil {
			if err == nil {
				err = repositoryErr
			} else {
				err = errors.Append(err, repositoryErr)
			}
		}
	}

	return err
}

func (e *uploadExpirer) HandleError(err error) {
	e.metrics.numErrors.Inc()
	log15.Error("Failed to expire old codeintel records", "error", err)
}

func (e *uploadExpirer) handleRepository(
	ctx context.Context,
	repositoryID int,
	now time.Time,
) error {
	e.metrics.numRepositoriesScanned.Inc()

	// Build a map from commits to the set of policies that affect them. Note that this map should
	// never be empty as we have multiple protected data retention policies on the global scope so
	// that all data visible from a tag or branch tip is protected for at least a short amount of
	// time after upload.
	commitMap, err := e.buildCommitMap(ctx, repositoryID, now)
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
	lastRetentionScanBefore := now.Add(-e.uploadProcessDelay)

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

		uploads, _, err := e.dbStore.GetUploads(ctx, dbstore.GetUploadsOptions{
			State:                   "completed",
			RepositoryID:            repositoryID,
			AllowExpired:            false,
			OldestFirst:             true,
			Limit:                   e.uploadBatchSize,
			LastRetentionScanBefore: &lastRetentionScanBefore,
			InCommitGraph:           true,
		})
		if err != nil || len(uploads) == 0 {
			return err
		}

		if err := e.handleUploads(ctx, commitMap, uploads, now); err != nil {
			// Note that we collect errors in the lop of the handleUploads call, but we will still terminate
			// this loop on any non-nil error from that function. This is required to prevent us from pullling
			// back the same set of failing records from the database in a tight loop.
			return err
		}
	}
}

// buildCommitMap will iterate the complete set of configuration policies that apply to a particular
// repository and build a map from commits to the policies that apply to them.
func (e *uploadExpirer) buildCommitMap(ctx context.Context, repositoryID int, now time.Time) (map[string][]policies.PolicyMatch, error) {
	var (
		offset   int
		policies []dbstore.ConfigurationPolicy
	)

	for {
		// Retrieve the complete set of configuration policies that affect data retention for this repository
		policyBatch, totalCount, err := e.dbStore.GetConfigurationPolicies(ctx, dbstore.GetConfigurationPoliciesOptions{
			RepositoryID:     repositoryID,
			ForDataRetention: true,
			Limit:            e.policyBatchSize,
			Offset:           offset,
		})
		if err != nil {
			return nil, errors.Wrap(err, "dbstore.GetConfigurationPolicies")
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

func (e *uploadExpirer) handleUploads(
	ctx context.Context,
	commitMap map[string][]policies.PolicyMatch,
	uploads []dbstore.Upload,
	now time.Time,
) (err error) {
	// Categorize each upload as protected or expired
	var (
		protectedUploadIDs = make([]int, 0, len(uploads))
		expiredUploadIDs   = make([]int, 0, len(uploads))
	)

	for _, upload := range uploads {
		protected, checkErr := e.isUploadProtectedByPolicy(ctx, commitMap, upload, now)
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

	if updateErr := e.dbStore.UpdateUploadRetention(ctx, protectedUploadIDs, expiredUploadIDs); updateErr != nil {
		if updateErr := errors.Wrap(err, "dbstore.UpdateUploadRetention"); err == nil {
			err = updateErr
		} else {
			err = errors.Append(err, updateErr)
		}
	}

	if count := len(expiredUploadIDs); count > 0 {
		log15.Info("Expiring codeintel uploads", "count", count)
		e.metrics.numUploadsExpired.Add(float64(count))
	}

	return err
}

func (e *uploadExpirer) isUploadProtectedByPolicy(
	ctx context.Context,
	commitMap map[string][]policies.PolicyMatch,
	upload dbstore.Upload,
	now time.Time,
) (bool, error) {
	e.metrics.numUploadsScanned.Inc()

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
		commits, nextToken, err := e.dbStore.CommitsVisibleToUpload(ctx, upload.ID, e.commitBatchSize, token)
		if err != nil {
			return false, errors.Wrap(err, "dbstore.CommitsVisibleToUpload")
		}
		token = nextToken

		e.metrics.numCommitsScanned.Add(float64(len(commits)))

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
