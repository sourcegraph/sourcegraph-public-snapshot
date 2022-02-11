package graphqlbackend

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// This constant defines the cursor prefix, which disambiguates a repository
// cursor from other types of cursors in the system.
const repositoryCursorKind = "RepositoryCursor"

// MarshalRepositoryCursor marshals a repository pagination cursor.
func MarshalRepositoryCursor(cursor *types.Cursor) string {
	return string(relay.MarshalID(repositoryCursorKind, cursor))
}

// UnmarshalRepositoryCursor unmarshals a repository pagination cursor.
func UnmarshalRepositoryCursor(cursor *string) (*types.Cursor, error) {
	if cursor == nil {
		return nil, nil
	}
	if kind := relay.UnmarshalKind(graphql.ID(*cursor)); kind != repositoryCursorKind {
		return nil, errors.Errorf("cannot unmarshal repository cursor type: %q", kind)
	}
	var spec *types.Cursor
	if err := relay.UnmarshalSpec(graphql.ID(*cursor), &spec); err != nil {
		return nil, err
	}
	return spec, nil
}
