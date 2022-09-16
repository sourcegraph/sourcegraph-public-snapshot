package graphql

import (
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type RetentionPolicyMatcherResolver struct {
	db           database.DB
	policy       RetentionPolicyMatchCandidate
	errCollector *observation.ErrCollector
}

func NewRetentionPolicyMatcherResolver(db database.DB, policy RetentionPolicyMatchCandidate) *RetentionPolicyMatcherResolver {
	return &RetentionPolicyMatcherResolver{db: db, policy: policy}
}

func (r *RetentionPolicyMatcherResolver) ConfigurationPolicy() gql.CodeIntelligenceConfigurationPolicyResolver {
	if r.policy.ConfigurationPolicy == nil {
		return nil
	}
	return NewConfigurationPolicyResolver(r.db, *r.policy.ConfigurationPolicy, r.errCollector)
}

func (r *RetentionPolicyMatcherResolver) Matches() bool {
	return r.policy.Matched
}

func (r *RetentionPolicyMatcherResolver) ProtectingCommits() *[]string {
	return &r.policy.ProtectingCommits
}
