package sharedresolvers

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type retentionPolicyMatcherResolver struct {
	svc          AutoIndexingService
	policy       types.RetentionPolicyMatchCandidate
	errCollector *observation.ErrCollector
}

func NewRetentionPolicyMatcherResolver(svc AutoIndexingService, policy types.RetentionPolicyMatchCandidate) resolverstubs.RetentionPolicyMatcherResolver {
	return &retentionPolicyMatcherResolver{svc: svc, policy: policy}
}

func (r *retentionPolicyMatcherResolver) ConfigurationPolicy() resolverstubs.CodeIntelligenceConfigurationPolicyResolver {
	if r.policy.ConfigurationPolicy == nil {
		return nil
	}
	return NewConfigurationPolicyResolver(r.svc, *r.policy.ConfigurationPolicy, r.errCollector)
}

func (r *retentionPolicyMatcherResolver) Matches() bool {
	return r.policy.Matched
}

func (r *retentionPolicyMatcherResolver) ProtectingCommits() *[]string {
	return &r.policy.ProtectingCommits
}
