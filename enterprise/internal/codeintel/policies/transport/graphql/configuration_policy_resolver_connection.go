package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type codeIntelligenceConfigurationPolicyConnectionResolver struct {
	policySvc  *policies.Service
	policies   []types.ConfigurationPolicy
	totalCount int
	errTracer  *observation.ErrCollector
}

func NewCodeIntelligenceConfigurationPolicyConnectionResolver(
	policySvc *policies.Service,
	policies []types.ConfigurationPolicy,
	totalCount int,
	errTracer *observation.ErrCollector,
) resolverstubs.CodeIntelligenceConfigurationPolicyConnectionResolver {
	return &codeIntelligenceConfigurationPolicyConnectionResolver{
		policySvc:  policySvc,
		policies:   policies,
		totalCount: totalCount,
		errTracer:  errTracer,
	}
}

func (r *codeIntelligenceConfigurationPolicyConnectionResolver) Nodes(ctx context.Context) ([]resolverstubs.CodeIntelligenceConfigurationPolicyResolver, error) {
	resolvers := make([]resolverstubs.CodeIntelligenceConfigurationPolicyResolver, 0, len(r.policies))
	for _, policy := range r.policies {
		resolvers = append(resolvers, NewConfigurationPolicyResolver(r.policySvc, policy, r.errTracer))
	}

	return resolvers, nil
}

func (r *codeIntelligenceConfigurationPolicyConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	v := int32(r.totalCount)
	return &v, nil
}

func (r *codeIntelligenceConfigurationPolicyConnectionResolver) PageInfo(ctx context.Context) (resolverstubs.PageInfo, error) {
	return HasNextPage(len(r.policies) < r.totalCount), nil
}
