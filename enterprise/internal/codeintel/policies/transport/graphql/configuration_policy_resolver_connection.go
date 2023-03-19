package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type codeIntelligenceConfigurationPolicyConnectionResolver struct {
	repoStore  database.RepoStore
	policies   []types.ConfigurationPolicy
	totalCount int
	errTracer  *observation.ErrCollector
}

func NewCodeIntelligenceConfigurationPolicyConnectionResolver(
	repoStore database.RepoStore,
	policies []types.ConfigurationPolicy,
	totalCount int,
	errTracer *observation.ErrCollector,
) resolverstubs.CodeIntelligenceConfigurationPolicyConnectionResolver {
	return &codeIntelligenceConfigurationPolicyConnectionResolver{
		repoStore:  repoStore,
		policies:   policies,
		totalCount: totalCount,
		errTracer:  errTracer,
	}
}

func (r *codeIntelligenceConfigurationPolicyConnectionResolver) Nodes(ctx context.Context) ([]resolverstubs.CodeIntelligenceConfigurationPolicyResolver, error) {
	resolvers := make([]resolverstubs.CodeIntelligenceConfigurationPolicyResolver, 0, len(r.policies))
	for _, policy := range r.policies {
		resolvers = append(resolvers, NewConfigurationPolicyResolver(r.repoStore, policy, r.errTracer))
	}

	return resolvers, nil
}

func (r *codeIntelligenceConfigurationPolicyConnectionResolver) TotalCount() *int32 {
	v := int32(r.totalCount)
	return &v
}

func (r *codeIntelligenceConfigurationPolicyConnectionResolver) PageInfo() resolverstubs.PageInfo {
	return HasNextPage(len(r.policies) < r.totalCount)
}
