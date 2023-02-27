package graphql

import (
	"context"
	"time"

	policiesshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
)

type Service interface {
	// Configurations
	GetConfigurationPolicies(ctx context.Context, opts policiesshared.GetConfigurationPoliciesOptions) ([]types.ConfigurationPolicy, int, error)
	GetConfigurationPolicyByID(ctx context.Context, id int) (_ types.ConfigurationPolicy, _ bool, err error)
	CreateConfigurationPolicy(ctx context.Context, configurationPolicy types.ConfigurationPolicy) (types.ConfigurationPolicy, error)
	UpdateConfigurationPolicy(ctx context.Context, policy types.ConfigurationPolicy) (err error)
	DeleteConfigurationPolicyByID(ctx context.Context, id int) (err error)

	// Retention Policy
	GetRetentionPolicyOverview(ctx context.Context, upload types.Upload, matchesOnly bool, first int, after int64, query string, now time.Time) (matches []types.RetentionPolicyMatchCandidate, totalCount int, err error)

	// Repository
	GetPreviewRepositoryFilter(ctx context.Context, patterns []string, limit, offset int) (_ []int, totalCount int, matchesAll bool, repositoryMatchLimit *int, _ error)
	GetPreviewGitObjectFilter(ctx context.Context, repositoryID int, gitObjectType types.GitObjectType, pattern string) (map[string][]string, error)
}
