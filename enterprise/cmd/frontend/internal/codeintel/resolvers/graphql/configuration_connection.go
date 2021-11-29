package graphql

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
)

type codeIntelligenceConfigurationPolicyConnectionResolver struct {
	policies   []dbstore.ConfigurationPolicy
	totalCount int
}

func NewCodeIntelligenceConfigurationPolicyConnectionResolver(policies []dbstore.ConfigurationPolicy, totalCount int) gql.CodeIntelligenceConfigurationPolicyConnectionResolver {
	return &codeIntelligenceConfigurationPolicyConnectionResolver{
		policies:   policies,
		totalCount: totalCount,
	}
}

func (r *codeIntelligenceConfigurationPolicyConnectionResolver) Nodes(ctx context.Context) ([]gql.CodeIntelligenceConfigurationPolicyResolver, error) {
	resolvers := make([]gql.CodeIntelligenceConfigurationPolicyResolver, 0, len(r.policies))
	for _, policy := range r.policies {
		resolvers = append(resolvers, NewConfigurationPolicyResolver(policy))
	}

	return resolvers, nil
}

func (r *codeIntelligenceConfigurationPolicyConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	v := int32(r.totalCount)
	return &v, nil
}

func (r *codeIntelligenceConfigurationPolicyConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(len(r.policies) < r.totalCount), nil
}
