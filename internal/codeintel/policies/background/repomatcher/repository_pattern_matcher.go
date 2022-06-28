package repomatcher

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/conf"
)

func (m *matcher) HandleRepositoryPatternMatcher(ctx context.Context) error {
	policies, err := m.dbStore.SelectPoliciesForRepositoryMembershipUpdate(ctx, ConfigInst.ConfigurationPolicyMembershipBatchSize)
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
		if err := m.dbStore.UpdateReposMatchingPatterns(ctx, patterns, policy.ID, repositoryMatchLimit); err != nil {
			return err
		}

		m.metrics.numPoliciesUpdated.Inc()
	}

	return nil
}

// func (m *matcher) HandleError(err error) {
// 	m.metrics.numErrors.Inc()
// 	log15.Error("Failed to match pattern for repository", "error", err)
// }
