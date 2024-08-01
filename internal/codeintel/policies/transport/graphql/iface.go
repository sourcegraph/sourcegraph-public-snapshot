package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
	policiesshared "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
)

type PoliciesService interface {
	// Fetch policies
	GetConfigurationPolicies(ctx context.Context, opts policiesshared.GetConfigurationPoliciesOptions) ([]shared.ConfigurationPolicy, int, error)
	GetConfigurationPolicyByID(ctx context.Context, id int) (shared.ConfigurationPolicy, bool, error)

	// Modify policies
	CreateConfigurationPolicy(ctx context.Context, configurationPolicy shared.ConfigurationPolicy) (shared.ConfigurationPolicy, error)
	UpdateConfigurationPolicy(ctx context.Context, policy shared.ConfigurationPolicyPatch) (err error)
	DeleteConfigurationPolicyByID(ctx context.Context, id int) error

	// Filter previews
	GetPreviewRepositoryFilter(ctx context.Context, patterns []string, limit int) (_ []int, totalCount int, matchesAll bool, repositoryMatchLimit *int, _ error)
	GetPreviewGitObjectFilter(ctx context.Context, repositoryID int, gitObjectType shared.GitObjectType, pattern string, limit int, countObjectsYoungerThanHours *int32) (_ []policies.GitObject, totalCount int, totalCountYoungerThanThreshold *int, _ error)
}
