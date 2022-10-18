package sharedresolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type CodeIntelligenceRetentionPolicyMatchesConnectionResolver interface {
	Nodes(ctx context.Context) ([]CodeIntelligenceRetentionPolicyMatchResolver, error)
	TotalCount(ctx context.Context) (*int32, error)
	PageInfo(ctx context.Context) (*PageInfo, error)
}

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

func (r *codeIntelligenceRetentionPolicyMatcherConnectionResolver) Nodes(ctx context.Context) ([]CodeIntelligenceRetentionPolicyMatchResolver, error) {
	resolvers := make([]CodeIntelligenceRetentionPolicyMatchResolver, 0, len(r.policies))
	for _, policy := range r.policies {
		resolvers = append(resolvers, NewRetentionPolicyMatcherResolver(r.svc, policy))
	}

	return resolvers, nil
}

func (r *codeIntelligenceRetentionPolicyMatcherConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	v := int32(r.totalCount)
	return &v, nil
}

func (r *codeIntelligenceRetentionPolicyMatcherConnectionResolver) PageInfo(ctx context.Context) (*PageInfo, error) {
	return HasNextPage(len(r.policies) < r.totalCount), nil
}
