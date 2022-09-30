package scheduler

import (
	"context"
	"time"

	policies "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/enterprise"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
)

type PolicyMatcher interface {
	CommitsDescribedByPolicyInternal(ctx context.Context, repositoryID int, policies []types.ConfigurationPolicy, now time.Time, filterCommits ...string) (map[string][]policies.PolicyMatch, error)
}

type PolicyService interface {
	GetConfigurationPolicies(ctx context.Context, opts types.GetConfigurationPoliciesOptions) ([]types.ConfigurationPolicy, int, error)
}

type UploadService interface {
	GetRepositoriesForIndexScan(ctx context.Context, table, column string, processDelay time.Duration, allowGlobalPolicies bool, repositoryMatchLimit *int, limit int, now time.Time) (_ []int, err error)
}
