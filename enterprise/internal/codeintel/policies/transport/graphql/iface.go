package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies"
	policiesshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
)

type PoliciesService interface {
	// Fetch policies
	GetConfigurationPolicies(ctx context.Context, opts policiesshared.GetConfigurationPoliciesOptions) ([]types.ConfigurationPolicy, int, error)
	GetConfigurationPolicyByID(ctx context.Context, id int) (types.ConfigurationPolicy, bool, error)

	// Modify policies
	CreateConfigurationPolicy(ctx context.Context, configurationPolicy types.ConfigurationPolicy) (types.ConfigurationPolicy, error)
	UpdateConfigurationPolicy(ctx context.Context, policy types.ConfigurationPolicy) (err error)
	DeleteConfigurationPolicyByID(ctx context.Context, id int) error

	// Filter previews
	GetPreviewRepositoryFilter(ctx context.Context, patterns []string, limit int) (_ []int, totalCount int, matchesAll bool, repositoryMatchLimit *int, _ error)
	GetPreviewGitObjectFilter(ctx context.Context, repositoryID int, gitObjectType types.GitObjectType, pattern string, limit int, countObjectsYoungerThanHours *int32) (_ []policies.GitObject, totalCount int, totalCountYoungerThanThreshold *int, _ error)
}
