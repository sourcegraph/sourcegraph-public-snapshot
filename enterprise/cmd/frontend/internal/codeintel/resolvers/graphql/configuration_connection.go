package graphql

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type codeIntelligenceConfigurationPolicyConnectionResolver struct {
	db         database.DB
	policies   []dbstore.ConfigurationPolicy
	totalCount int
	errTracer  *observation.ErrCollector
}

func NewCodeIntelligenceConfigurationPolicyConnectionResolver(db database.DB, policies []dbstore.ConfigurationPolicy, totalCount int, errTracer *observation.ErrCollector) gql.CodeIntelligenceConfigurationPolicyConnectionResolver {
	return &codeIntelligenceConfigurationPolicyConnectionResolver{
		db:         db,
		policies:   policies,
		totalCount: totalCount,
		errTracer:  errTracer,
	}
}

func (r *codeIntelligenceConfigurationPolicyConnectionResolver) Nodes(ctx context.Context) ([]gql.CodeIntelligenceConfigurationPolicyResolver, error) {
	resolvers := make([]gql.CodeIntelligenceConfigurationPolicyResolver, 0, len(r.policies))
	for _, policy := range r.policies {
		resolvers = append(resolvers, NewConfigurationPolicyResolver(r.db, policy, r.errTracer))
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
