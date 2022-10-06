package policies

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *Service) NewRepositoryMatcher(interval time.Duration, configurationPolicyMembershipBatchSize int) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		return s.handleRepositoryMatcherBatch(ctx, configurationPolicyMembershipBatchSize)
	}))
}

func (s *Service) handleRepositoryMatcherBatch(ctx context.Context, batchSize int) error {
	policies, err := s.SelectPoliciesForRepositoryMembershipUpdate(ctx, batchSize)
	if err != nil {
		return err
	}

	for _, policy := range policies {
		var patterns []string
		if policy.RepositoryPatterns != nil {
			patterns = *policy.RepositoryPatterns
		}

		var repositoryMatchLimit *int
		if val := conf.CodeIntelAutoIndexingPolicyRepositoryMatchLimit(); val != -1 {
			repositoryMatchLimit = &val
		}

		// Always call this even if patterns are not supplied. Otherwise we run into the
		// situation where we have deleted all of the patterns associated with a policy
		// but it still has entries in the lookup table.
		if err := s.UpdateReposMatchingPatterns(ctx, patterns, policy.ID, repositoryMatchLimit); err != nil {
			return err
		}

		s.matcherMetrics.numPoliciesUpdated.Inc()
	}

	return nil
}

func (s *Service) SelectPoliciesForRepositoryMembershipUpdate(ctx context.Context, batchSize int) (configurationPolicies []types.ConfigurationPolicy, err error) {
	ctx, _, endObservation := s.operations.selectPoliciesForRepositoryMembershipUpdate.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	configurationPolicies, err = s.store.SelectPoliciesForRepositoryMembershipUpdate(ctx, batchSize)
	if err != nil {
		return nil, err
	}

	return configurationPolicies, nil
}

func (s *Service) UpdateReposMatchingPatterns(ctx context.Context, patterns []string, policyID int, repositoryMatchLimit *int) (err error) {
	ctx, _, endObservation := s.operations.updateReposMatchingPatterns.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.UpdateReposMatchingPatterns(ctx, patterns, policyID, repositoryMatchLimit)
}
