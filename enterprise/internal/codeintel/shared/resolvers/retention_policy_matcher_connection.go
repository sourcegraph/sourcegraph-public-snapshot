package sharedresolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type codeIntelligenceRetentionPolicyMatcherConnectionResolver struct {
	svc        AutoIndexingService
	policies   []types.RetentionPolicyMatchCandidate
	totalCount int
	errTracer  *observation.ErrCollector
}

func NewCodeIntelligenceRetentionPolicyMatcherConnectionResolver(svc AutoIndexingService, policies []types.RetentionPolicyMatchCandidate, totalCount int, errTracer *observation.ErrCollector) *codeIntelligenceRetentionPolicyMatcherConnectionResolver {
	return &codeIntelligenceRetentionPolicyMatcherConnectionResolver{
		svc:        svc,
		policies:   policies,
		totalCount: totalCount,
		errTracer:  errTracer,
	}
}

func (r *codeIntelligenceRetentionPolicyMatcherConnectionResolver) Nodes(ctx context.Context) ([]resolverstubs.CodeIntelligenceRetentionPolicyMatchResolver, error) {
	resolvers := make([]resolverstubs.CodeIntelligenceRetentionPolicyMatchResolver, 0, len(r.policies))
	for _, policy := range r.policies {
		resolvers = append(resolvers, NewRetentionPolicyMatcherResolver(r.svc, policy))
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
