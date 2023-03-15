package sharedresolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type codeIntelligenceRetentionPolicyMatcherConnectionResolver struct {
	repoStore  database.RepoStore
	policies   []types.RetentionPolicyMatchCandidate
	totalCount int
	errTracer  *observation.ErrCollector
}

func NewCodeIntelligenceRetentionPolicyMatcherConnectionResolver(repoStore database.RepoStore, policies []types.RetentionPolicyMatchCandidate, totalCount int, errTracer *observation.ErrCollector) *codeIntelligenceRetentionPolicyMatcherConnectionResolver {
	return &codeIntelligenceRetentionPolicyMatcherConnectionResolver{
		repoStore:  repoStore,
		policies:   policies,
		totalCount: totalCount,
		errTracer:  errTracer,
	}
}

func (r *codeIntelligenceRetentionPolicyMatcherConnectionResolver) Nodes(ctx context.Context) ([]resolverstubs.CodeIntelligenceRetentionPolicyMatchResolver, error) {
	resolvers := make([]resolverstubs.CodeIntelligenceRetentionPolicyMatchResolver, 0, len(r.policies))
	for _, policy := range r.policies {
		resolvers = append(resolvers, NewRetentionPolicyMatcherResolver(r.repoStore, policy))
	}

	return resolvers, nil
}

func (r *codeIntelligenceRetentionPolicyMatcherConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	v := int32(r.totalCount)
	return &v, nil
}

func (r *codeIntelligenceRetentionPolicyMatcherConnectionResolver) PageInfo(ctx context.Context) (resolverstubs.PageInfo, error) {
	return HasNextPage(len(r.policies) < r.totalCount), nil
}
