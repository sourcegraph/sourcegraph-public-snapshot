package graphql

import (
	"time"

	"github.com/cockroachdb/errors"
	"github.com/graph-gophers/graphql-go"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
)

type configurationPolicyResolver struct {
	configurationPolicy store.ConfigurationPolicy
}

func NewConfigurationPolicyResolver(configurationPolicy store.ConfigurationPolicy) gql.CodeIntelligenceConfigurationPolicyResolver {
	return &configurationPolicyResolver{
		configurationPolicy: configurationPolicy,
	}
}

func (r *configurationPolicyResolver) ID() graphql.ID {
	return marshalConfigurationPolicyGQLID(int64(r.configurationPolicy.ID))
}

func (r *configurationPolicyResolver) Name() string {
	return r.configurationPolicy.Name
}

func (r *configurationPolicyResolver) Type() (gql.GitObjectType, error) {
	switch r.configurationPolicy.Type {
	case store.GitObjectTypeCommit:
		return gql.GitObjectTypeCommit, nil
	case store.GitObjectTypeTag:
		return gql.GitObjectTypeTag, nil
	case store.GitObjectTypeTree:
		return gql.GitObjectTypeTree, nil
	default:
		return "", errors.Errorf("unknown git object type %s", r.configurationPolicy.Type)
	}
}

func (r *configurationPolicyResolver) Pattern() string {
	return r.configurationPolicy.Pattern
}

func (r *configurationPolicyResolver) Protected() bool {
	return r.configurationPolicy.Protected
}

func (r *configurationPolicyResolver) RetentionEnabled() bool {
	return r.configurationPolicy.RetentionEnabled
}

func (r *configurationPolicyResolver) RetentionDurationHours() *int32 {
	return toHours(r.configurationPolicy.RetentionDuration)
}

func (r *configurationPolicyResolver) RetainIntermediateCommits() bool {
	return r.configurationPolicy.RetainIntermediateCommits
}

func (r *configurationPolicyResolver) IndexingEnabled() bool {
	return r.configurationPolicy.IndexingEnabled
}

func (r *configurationPolicyResolver) IndexCommitMaxAgeHours() *int32 {
	return toHours(r.configurationPolicy.IndexCommitMaxAge)
}

func (r *configurationPolicyResolver) IndexIntermediateCommits() bool {
	return r.configurationPolicy.IndexIntermediateCommits
}

func toHours(duration *time.Duration) *int32 {
	if duration == nil {
		return nil
	}

	v := int32(*duration / time.Hour)
	return &v
}
