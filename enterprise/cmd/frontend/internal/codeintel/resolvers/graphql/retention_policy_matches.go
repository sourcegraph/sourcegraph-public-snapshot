package graphql

import (
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type RetentionPolicyMatcherResolver struct {
	db           database.DB
	policy       retentionPolicyMatchCandidate
	errCollector *observation.ErrCollector
}

type retentionPolicyMatchCandidate struct {
	dbstore.ConfigurationPolicy
	matched           bool
	protectingCommits []string
}

func NewRetentionPolicyMatcherResolver(db database.DB, policy retentionPolicyMatchCandidate) *RetentionPolicyMatcherResolver {
	return &RetentionPolicyMatcherResolver{db: db, policy: policy}
}

func (r *RetentionPolicyMatcherResolver) ConfigurationPolicy() gql.CodeIntelligenceConfigurationPolicyResolver {
	return NewConfigurationPolicyResolver(r.db, r.policy.ConfigurationPolicy, r.errCollector)
}

func (r *RetentionPolicyMatcherResolver) Matches() bool {
	return r.policy.matched
}

func (r *RetentionPolicyMatcherResolver) ProtectingCommits() *[]string {
	return &r.policy.protectingCommits
}
