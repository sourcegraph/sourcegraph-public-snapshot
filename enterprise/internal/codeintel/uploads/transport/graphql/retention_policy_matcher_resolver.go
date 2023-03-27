package graphql

import (
	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type retentionPolicyMatcherResolver struct {
	repoStore    database.RepoStore
	policy       types.RetentionPolicyMatchCandidate
	errCollector *observation.ErrCollector
}

func NewRetentionPolicyMatcherResolver(repoStore database.RepoStore, policy types.RetentionPolicyMatchCandidate) resolverstubs.CodeIntelligenceRetentionPolicyMatchResolver {
	return &retentionPolicyMatcherResolver{repoStore: repoStore, policy: policy}
}

func (r *retentionPolicyMatcherResolver) ConfigurationPolicy() resolverstubs.CodeIntelligenceConfigurationPolicyResolver {
	if r.policy.ConfigurationPolicy == nil {
		return nil
	}
	return sharedresolvers.NewConfigurationPolicyResolver(r.repoStore, *r.policy.ConfigurationPolicy, r.errCollector)
}

func (r *retentionPolicyMatcherResolver) Matches() bool {
	return r.policy.Matched
}

func (r *retentionPolicyMatcherResolver) ProtectingCommits() *[]string {
	return &r.policy.ProtectingCommits
}
