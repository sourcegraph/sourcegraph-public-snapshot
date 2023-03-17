package graphql

import (
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func marshalConfigurationPolicyGQLID(configurationPolicyID int64) graphql.ID {
	return relay.MarshalID("CodeIntelligenceConfigurationPolicy", configurationPolicyID)
}

func unmarshalRepositoryID(id graphql.ID) (repositoryID int64, err error) {
	err = relay.UnmarshalSpec(id, &repositoryID)
	return repositoryID, err
}

// PageInfo implements the GraphQL type PageInfo.
type PageInfo struct {
	endCursor   *string
	hasNextPage bool
}

// HasNextPage returns a new PageInfo with the given hasNextPage value.
func HasNextPage(hasNextPage bool) *PageInfo {
	return &PageInfo{hasNextPage: hasNextPage}
}

func (r *PageInfo) EndCursor() *string { return r.endCursor }
func (r *PageInfo) HasNextPage() bool  { return r.hasNextPage }

func validateConfigurationPolicy(policy resolverstubs.CodeIntelConfigurationPolicy) error {
	switch types.GitObjectType(policy.Type) {
	case types.GitObjectTypeCommit:
	case types.GitObjectTypeTag:
	case types.GitObjectTypeTree:
	default:
		return errors.Errorf("illegal git object type '%s', expected 'GIT_COMMIT', 'GIT_TAG', or 'GIT_TREE'", policy.Type)
	}

	if policy.Name == "" {
		return errors.Errorf("no name supplied")
	}
	if policy.Pattern == "" {
		return errors.Errorf("no pattern supplied")
	}
	if types.GitObjectType(policy.Type) == types.GitObjectTypeCommit && policy.Pattern != "HEAD" {
		return errors.Errorf("pattern must be HEAD for policy type 'GIT_COMMIT'")
	}
	if policy.RetentionDurationHours != nil && (*policy.RetentionDurationHours < 0 || *policy.RetentionDurationHours > maxDurationHours) {
		return errors.Errorf("illegal retention duration '%d'", *policy.RetentionDurationHours)
	}
	if policy.IndexingEnabled && policy.IndexCommitMaxAgeHours != nil && (*policy.IndexCommitMaxAgeHours < 0 || *policy.IndexCommitMaxAgeHours > maxDurationHours) {
		return errors.Errorf("illegal index commit max age '%d'", *policy.IndexCommitMaxAgeHours)
	}

	return nil
}

const maxDurationHours = 87600 // 10 years

func toDuration(hours *int32) *time.Duration {
	if hours == nil {
		return nil
	}

	v := time.Duration(*hours) * time.Hour
	return &v
}

func unmarshalConfigurationPolicyGQLID(id graphql.ID) (configurationPolicyID int64, err error) {
	err = relay.UnmarshalSpec(id, &configurationPolicyID)
	return configurationPolicyID, err
}

// toInt32 translates the given int pointer into an int32 pointer.
func toInt32(val *int) *int32 {
	if val == nil {
		return nil
	}

	v := int32(*val)
	return &v
}

func unmarshalLSIFIndexGQLID(id graphql.ID) (indexID int64, err error) {
	err = relay.UnmarshalSpec(id, &indexID)
	return indexID, err
}
