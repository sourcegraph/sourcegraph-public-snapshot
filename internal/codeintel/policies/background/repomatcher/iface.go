package repomatcher

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
)

type PolicyService interface {
	SelectPoliciesForRepositoryMembershipUpdate(ctx context.Context, batchSize int) (configurationPolicies []shared.ConfigurationPolicy, err error)
	UpdateReposMatchingPatterns(ctx context.Context, patterns []string, policyID int, repositoryMatchLimit *int) (err error)
}
