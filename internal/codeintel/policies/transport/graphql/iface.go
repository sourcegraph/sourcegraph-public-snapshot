package graphql

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
)

type Service interface {
	// Configurations
	GetConfigurationPolicies(ctx context.Context, opts shared.GetConfigurationPoliciesOptions) ([]shared.ConfigurationPolicy, int, error)
	GetConfigurationPolicyByID(ctx context.Context, id int) (_ shared.ConfigurationPolicy, _ bool, err error)
	CreateConfigurationPolicy(ctx context.Context, configurationPolicy shared.ConfigurationPolicy) (shared.ConfigurationPolicy, error)
	UpdateConfigurationPolicy(ctx context.Context, policy shared.ConfigurationPolicy) (err error)
	DeleteConfigurationPolicyByID(ctx context.Context, id int) (err error)

	// Retention Policy
	GetRetentionPolicyOverview(ctx context.Context, upload shared.Upload, matchesOnly bool, first int, after int64, query string, now time.Time) (matches []shared.RetentionPolicyMatchCandidate, totalCount int, err error)

	// Repository
	GetPreviewRepositoryFilter(ctx context.Context, patterns []string, limit, offset int) (_ []int, totalCount int, repositoryMatchLimit *int, _ error)
	GetPreviewGitObjectFilter(ctx context.Context, repositoryID int, gitObjectType shared.GitObjectType, pattern string) (map[string][]string, error)
}
