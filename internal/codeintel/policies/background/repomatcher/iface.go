package repomatcher

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
)

type DBStore interface {
	SelectPoliciesForRepositoryMembershipUpdate(ctx context.Context, batchSize int) (configurationPolicies []dbstore.ConfigurationPolicy, err error)
	UpdateReposMatchingPatterns(ctx context.Context, patterns []string, policyID int, repositoryMatchLimit *int) (err error)
}
