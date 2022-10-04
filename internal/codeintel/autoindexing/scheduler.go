package autoindexing

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (s *Service) NewScheduler(
	interval time.Duration,
	repositoryProcessDelay time.Duration,
	repositoryBatchSize int,
	policyBatchSize int,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutineWithMetrics(context.Background(), interval, &scheduler{
		autoindexingSvc:        s,
		policySvc:              s.policiesSvc,
		uploadSvc:              s.uploadSvc,
		policyMatcher:          s.policyMatcher,
		logger:                 log.Scoped("autoindexing-scheduler", ""),
		repositoryProcessDelay: repositoryProcessDelay,
		repositoryBatchSize:    repositoryBatchSize,
		policyBatchSize:        policyBatchSize,
	}, s.operations.handleIndexScheduler)
}

type scheduler struct {
	autoindexingSvc        *Service
	policySvc              PoliciesService
	uploadSvc              shared.UploadService
	policyMatcher          PolicyMatcher
	logger                 log.Logger
	repositoryProcessDelay time.Duration
	repositoryBatchSize    int
	policyBatchSize        int
}

var (
	_ goroutine.Handler      = &scheduler{}
	_ goroutine.ErrorHandler = &scheduler{}
)

func (s *scheduler) Handle(ctx context.Context) error {
	if !autoIndexingEnabled() {
		return nil
	}

	var repositoryMatchLimit *int
	if val := conf.CodeIntelAutoIndexingPolicyRepositoryMatchLimit(); val != -1 {
		repositoryMatchLimit = &val
	}

	// Get the batch of repositories that we'll handle in this invocation of the periodic goroutine. This
	// set should contain repositories that have yet to be updated, or that have been updated least recently.
	// This allows us to update every repository reliably, even if it takes a long time to process through
	// the backlog.
	repositories, err := s.uploadSvc.GetRepositoriesForIndexScan(
		ctx,
		"lsif_last_index_scan",
		"last_index_scan_at",
		s.repositoryProcessDelay,
		conf.CodeIntelAutoIndexingAllowGlobalPolicies(),
		repositoryMatchLimit,
		s.repositoryBatchSize,
		time.Now(),
	)
	if err != nil {
		return errors.Wrap(err, "uploadSvc.GetRepositoriesForIndexScan")
	}
	if len(repositories) == 0 {
		// All repositories updated recently enough
		return nil
	}

	now := timeutil.Now()

	for _, repositoryID := range repositories {
		if repositoryErr := s.handleRepository(ctx, repositoryID, now); repositoryErr != nil {
			if err == nil {
				err = repositoryErr
			} else {
				err = errors.Append(err, repositoryErr)
			}
		}
	}

	return err
}

func (s *scheduler) handleRepository(ctx context.Context, repositoryID int, now time.Time) error {
	offset := 0

	for {
		// Retrieve the set of configuration policies that affect indexing for this repository.
		policies, totalCount, err := s.policySvc.GetConfigurationPolicies(ctx, types.GetConfigurationPoliciesOptions{
			RepositoryID: repositoryID,
			ForIndexing:  true,
			Limit:        s.policyBatchSize,
			Offset:       offset,
		})
		if err != nil {
			return errors.Wrap(err, "policySvc.GetConfigurationPolicies")
		}
		offset += len(policies)

		// Get the set of commits within this repository that match an indexing policy
		commitMap, err := s.policyMatcher.CommitsDescribedByPolicyInternal(ctx, repositoryID, policies, now)
		if err != nil {
			return errors.Wrap(err, "policies.CommitsDescribedByPolicy")
		}

		for commit, policyMatches := range commitMap {
			if len(policyMatches) == 0 {
				continue
			}

			// Attempt to queue an index if one does not exist for each of the matching commits
			if _, err := s.autoindexingSvc.QueueIndexes(ctx, repositoryID, commit, "", false, false); err != nil {
				if errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
					continue
				}

				return errors.Wrap(err, "indexEnqueuer.QueueIndexes")
			}
		}

		if len(policies) == 0 || offset >= totalCount {
			return nil
		}
	}
}

func (s *scheduler) HandleError(err error) {
	s.logger.Error("Failed to schedule index jobs", log.Error(err))
}
