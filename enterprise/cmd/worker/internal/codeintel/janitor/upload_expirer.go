package janitor

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gobwas/glob"
	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

type uploadExpirer struct {
	dbStore                DBStore
	gitserverClient        GitserverClient
	metrics                *metrics
	repositoryProcessDelay time.Duration
	repositoryBatchSize    int
	uploadProcessDelay     time.Duration
	uploadBatchSize        int
	commitBatchSize        int
	branchesCacheMaxKeys   int
}

var _ goroutine.Handler = &uploadExpirer{}
var _ goroutine.ErrorHandler = &uploadExpirer{}

// NewUploadExpirer returns a background routine that periodically compares the age of upload records
// against the age of uploads protected by global and repository specific data retention policies.
//
// Uploads that are older than the protected retention age are marked as expired. Expired records with
// no dependents will be removed by the expiredUploadDeleter.
func NewUploadExpirer(
	dbStore DBStore,
	gitserverClient GitserverClient,
	repositoryProcessDelay time.Duration,
	repositoryBatchSize int,
	uploadProcessDelay time.Duration,
	uploadBatchSize int,
	commitBatchSize int,
	branchesCacheMaxKeys int,
	interval time.Duration,
	metrics *metrics,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, &uploadExpirer{
		dbStore:                dbStore,
		gitserverClient:        gitserverClient,
		metrics:                metrics,
		repositoryProcessDelay: repositoryProcessDelay,
		repositoryBatchSize:    repositoryBatchSize,
		uploadProcessDelay:     uploadProcessDelay,
		uploadBatchSize:        uploadBatchSize,
		commitBatchSize:        commitBatchSize,
		branchesCacheMaxKeys:   branchesCacheMaxKeys,
	})
}

func (e *uploadExpirer) Handle(ctx context.Context) (err error) {
	// Get the batch of repositories that we'll handle in this invocation of the periodic goroutine. This
	// set should contain repositories that have yet to be updated, or that have been updated least recently.
	// This allows us to update every repository reliably, even if it takes a long time to process through
	// the backlog.
	lastUpdatedAtByRepository, err := e.dbStore.SelectRepositoriesForRetentionScan(ctx, e.repositoryProcessDelay, e.repositoryBatchSize)
	if err != nil {
		return errors.Wrap(err, "dbstore.SelectRepositoriesForRetentionScan")
	}
	if len(lastUpdatedAtByRepository) == 0 {
		// All repositories updated recently enough
		return nil
	}

	// Retrieve the set of global configuration policies that affect data retention. These policies are
	// applied to all repositories.
	globalPolicies, err := e.dbStore.GetConfigurationPolicies(ctx, dbstore.GetConfigurationPoliciesOptions{
		ForDataRetention: true,
	})
	if err != nil {
		return errors.Wrap(err, "dbstore.GetConfigurationPolicies")
	}

	now := timeutil.Now()

	for repositoryID, repositoryLastUpdatedAt := range lastUpdatedAtByRepository {
		if repositoryErr := e.handleRepository(ctx, repositoryID, repositoryLastUpdatedAt, globalPolicies, now); repositoryErr != nil {
			if err == nil {
				err = repositoryErr
			} else {
				err = multierror.Append(err, repositoryErr)
			}
		}
	}

	return nil
}

func (e *uploadExpirer) HandleError(err error) {
	e.metrics.numErrors.Inc()
	log15.Error("Failed to expire old codeintel records", "error", err)
}

func (e *uploadExpirer) handleRepository(
	ctx context.Context,
	repositoryID int,
	repositoryLastUpdatedAt *time.Time,
	globalPolicies []dbstore.ConfigurationPolicy,
	now time.Time,
) error {
	e.metrics.numRepositoriesScanned.Inc()

	// Retrieve the set of configuration policies that affect data retention. These policies are applied
	// only to this repository.
	repositoryPolicies, err := e.dbStore.GetConfigurationPolicies(ctx, dbstore.GetConfigurationPoliciesOptions{
		RepositoryID:     repositoryID,
		ForDataRetention: true,
	})
	if err != nil {
		return errors.Wrap(err, "dbstore.GetConfigurationPolicies")
	}

	// Combine global and repository-specific policies. Note that this resulting slice should never be
	// empty as we have a pair of protected data retention policies on the global scope so that all data
	// visible from a tag or branch tip is protected for at least a short amount of time after upload.
	policies := append(globalPolicies, repositoryPolicies...)

	// Pre-compile the glob patterns in all the policies to reduce the number of times we need to compile
	// the pattern in the loops below.
	patterns, err := compilePatterns(policies)
	if err != nil {
		return err
	}

	// Get a list of relevant branch and tag heads of this repository
	refDescriptions, err := e.gitserverClient.RefDescriptions(ctx, repositoryID)
	if err != nil {
		return errors.Wrap(err, "gitserver.RefDescriptions")
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

	// Create a cache structure shared by the routine that processes each upload. An upload can be
	// visible from many commits at once, so it is likely that the same commit is re-processed many
	// times. This cache prevents us from making redundant gitserver requests, and from wasting
	// compute time iterating through the same data already in memory.
	repositoryState, err := newRepositoryExpirationState(repositoryID, e.gitserverClient, e.branchesCacheMaxKeys)
	if err != nil {
		return err
	}

	for {
		// Each record pulled back by this query will either have its expired flag or its last
		// retention scan timestamp updated by the following handleUploads call. This guarantees
		// that the loop will terminate naturally after the entire set of candidate uploads have
		// been seen and updated with a time necessarily greater than lastRetentionScanBefore.

		uploads, _, err := e.dbStore.GetUploads(ctx, dbstore.GetUploadsOptions{
			State:                   "completed",
			RepositoryID:            repositoryID,
			AllowExpired:            false,
			OldestFirst:             true,
			Limit:                   e.uploadBatchSize,
			LastRetentionScanBefore: &lastRetentionScanBefore,
			UploadedBefore:          repositoryLastUpdatedAt,
		})
		if err != nil || len(uploads) == 0 {
			return err
		}

		if err := e.handleUploads(ctx, policies, patterns, refDescriptions, repositoryState, uploads, now); err != nil {
			// Note that we collect errors in the lop of the handleUploads call, but we will still terminate
			// this loop on any non-nil error from that function. This is required to prevent us from pullling
			// back the same set of failing records from the database in a tight loop.
			return err
		}
	}
}

func (e *uploadExpirer) handleUploads(
	ctx context.Context,
	policies []dbstore.ConfigurationPolicy,
	patterns map[string]glob.Glob,
	refDescriptions map[string][]gitserver.RefDescription,
	repositoryState *repositoryExpirationState,
	uploads []dbstore.Upload,
	now time.Time,
) (err error) {
	e.metrics.numUploadsScanned.Inc()

	// Categorize each upload as protected or expired
	protectedUploadIDs := make([]int, 0, len(uploads))
	expiredUploadIDs := make([]int, 0, len(uploads))

	for _, upload := range uploads {
		protected, checkErr := e.isUploadProtectedByPolicy(ctx, policies, patterns, refDescriptions, repositoryState, upload, now)
		if checkErr != nil {
			if err == nil {
				err = checkErr
			} else {
				err = multierror.Append(err, checkErr)
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
			err = multierror.Append(err, updateErr)
		}
	}

	e.metrics.numUploadsExpired.Add(float64(len(expiredUploadIDs)))
	return err
}

func (e *uploadExpirer) isUploadProtectedByPolicy(
	ctx context.Context,
	policies []dbstore.ConfigurationPolicy,
	patterns map[string]glob.Glob,
	refDescriptions map[string][]gitserver.RefDescription,
	repositoryState *repositoryExpirationState,
	upload dbstore.Upload,
	now time.Time,
) (bool, error) {
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

		if ok, err := isUploadCommitProtectedByPolicy(
			ctx,
			policies,
			patterns,
			refDescriptions,
			repositoryState,
			upload,
			commits,
			now,
		); err != nil || ok {
			return ok, err
		}
	}

	return false, nil
}

// compilePatterns constructs a map from patterns in each given policy to a compiled glob object used
// to match to commits, branch names, and tag names. If there are multiple policies with the same pattern,
// the pattern is compiled only once.
func compilePatterns(policies []dbstore.ConfigurationPolicy) (map[string]glob.Glob, error) {
	patterns := make(map[string]glob.Glob, len(policies))

	for _, policy := range policies {
		if _, ok := patterns[policy.Pattern]; ok {
			continue
		}

		pattern, err := glob.Compile(policy.Pattern)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to compile glob pattern `%s` in configuration policy %d", policy.Pattern, policy.ID))
		}

		patterns[policy.Pattern] = pattern
	}

	return patterns, nil
}
