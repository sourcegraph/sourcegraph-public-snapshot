package repomatcher

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
)

type PolicyService interface {
	SelectPoliciesForRepositoryMembershipUpdate(ctx context.Context, batchSize int) (configurationPolicies []types.ConfigurationPolicy, err error)
	UpdateReposMatchingPatterns(ctx context.Context, patterns []string, policyID int, repositoryMatchLimit *int) (err error)
}
