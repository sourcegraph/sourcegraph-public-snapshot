package graphqlbackend

import (
	"github.com/cockroachdb/errors"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

// This constant defines the cursor prefix, which disambiguates a repository
// cursor from other types of cursors in the system.
const repositoryCursorKind = "RepositoryCursor"

// marshalRepositoryCursor marshals a repository pagination cursor.
func marshalRepositoryCursor(cursor *database.Cursor) string {
	return string(relay.MarshalID(repositoryCursorKind, cursor))
}

// unmarshalRepositoryCursor unmarshals a repository pagination cursor.
func unmarshalRepositoryCursor(cursor *string) (*database.Cursor, error) {
	if cursor == nil {
		return nil, nil
	}
	if kind := relay.UnmarshalKind(graphql.ID(*cursor)); kind != repositoryCursorKind {
		return nil, errors.Errorf("cannot unmarshal repository cursor type: %q", kind)
	}
	var spec *database.Cursor
	if err := relay.UnmarshalSpec(graphql.ID(*cursor), &spec); err != nil {
		return nil, err
	}
	return spec, nil
}
