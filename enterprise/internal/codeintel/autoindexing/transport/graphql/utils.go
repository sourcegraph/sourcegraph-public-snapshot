package graphql

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

// strPtr creates a pointer to the given value. If the value is an
// empty string, a nil pointer is returned.
func strPtr(val string) *string {
	if val == "" {
		return nil
	}

	return &val
}

func marshalLSIFIndexGQLID(indexID int64) graphql.ID {
	return relay.MarshalID("LSIFIndex", indexID)
}

func unmarshalLSIFIndexGQLID(id graphql.ID) (indexID int64, err error) {
	err = relay.UnmarshalSpec(id, &indexID)
	return indexID, err
}

// resolveRepositoryByID gets a repository's internal identifier from a GraphQL identifier.
func resolveRepositoryID(id graphql.ID) (int, error) {
	if id == "" {
		return 0, nil
	}

	repoID, err := UnmarshalRepositoryID(id)
	if err != nil {
		return 0, err
	}

	return int(repoID), nil
}

func UnmarshalRepositoryID(id graphql.ID) (repo api.RepoID, err error) {
	err = relay.UnmarshalSpec(id, &repo)
	return
}

// derefString returns the underlying value in the given pointer.
// If the pointer is nil, the default value is returned.
func derefString(val *string, defaultValue string) string {
	if val != nil {
		return *val
	}
	return defaultValue
}
