package graphql

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type codeIntelligenceRetentionPolicyMatcherConnectionResolver struct {
	db         database.DB
	resolver   resolvers.Resolver
	policies   []RetentionPolicyMatchCandidate
	totalCount int
	errTracer  *observation.ErrCollector
}

func NewCodeIntelligenceRetentionPolicyMatcherConnectionResolver(db database.DB, resolver resolvers.Resolver, policies []RetentionPolicyMatchCandidate, totalCount int, errTracer *observation.ErrCollector) *codeIntelligenceRetentionPolicyMatcherConnectionResolver {
	return &codeIntelligenceRetentionPolicyMatcherConnectionResolver{
		db:         db,
		resolver:   resolver,
		policies:   policies,
		totalCount: totalCount,
		errTracer:  errTracer,
	}
}

func (r *codeIntelligenceRetentionPolicyMatcherConnectionResolver) Nodes(ctx context.Context) ([]gql.CodeIntelligenceRetentionPolicyMatchResolver, error) {
	resolvers := make([]gql.CodeIntelligenceRetentionPolicyMatchResolver, 0, len(r.policies))
	for _, policy := range r.policies {
		resolvers = append(resolvers, NewRetentionPolicyMatcherResolver(r.db, policy))
	}

	return resolvers, nil
}

func (r *codeIntelligenceRetentionPolicyMatcherConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	v := int32(r.totalCount)
	return &v, nil
}

func (r *codeIntelligenceRetentionPolicyMatcherConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(len(r.policies) < r.totalCount), nil
}
