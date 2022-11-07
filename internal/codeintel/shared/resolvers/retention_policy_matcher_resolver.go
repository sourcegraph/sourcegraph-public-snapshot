package sharedresolvers

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// TODO(@Numbers88s): consolidate reference to these two interfaces
type CodeIntelligenceRetentionPolicyMatchResolver interface {
	ConfigurationPolicy() CodeIntelligenceConfigurationPolicyResolver
	Matches() bool
	ProtectingCommits() *[]string
}

type RetentionPolicyMatcherResolver interface {
	ConfigurationPolicy() CodeIntelligenceConfigurationPolicyResolver
	Matches() bool
	ProtectingCommits() *[]string
}

type retentionPolicyMatcherResolver struct {
	svc          AutoIndexingService
	policy       types.RetentionPolicyMatchCandidate
	errCollector *observation.ErrCollector
}

func NewRetentionPolicyMatcherResolver(svc AutoIndexingService, policy types.RetentionPolicyMatchCandidate) RetentionPolicyMatcherResolver {
	return &retentionPolicyMatcherResolver{svc: svc, policy: policy}
}

func (r *retentionPolicyMatcherResolver) ConfigurationPolicy() CodeIntelligenceConfigurationPolicyResolver {
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
