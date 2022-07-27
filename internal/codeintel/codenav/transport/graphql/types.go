package graphql

import (
	"encoding/base64"
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// Constants for Ordered Monikers Types
const (
	implementation = "implementation"
	export         = "export"
)

type requestArgs struct {
	repo   *types.Repo
	commit string
	path   string
}

func (r *requestArgs) GetRepoID() int {
	return int(r.repo.ID)
}

// decodeCursor is the inverse of encodeCursor. If the given encoded string is empty, then
// a fresh cursor is returned.
func decodeImplementationsCursor(rawEncoded string) (shared.ImplementationsCursor, error) {
	if rawEncoded == "" {
		return shared.ImplementationsCursor{Phase: "local"}, nil
	}

	raw, err := base64.RawURLEncoding.DecodeString(rawEncoded)
	if err != nil {
		return shared.ImplementationsCursor{}, err
	}

	var cursor shared.ImplementationsCursor
	err = json.Unmarshal(raw, &cursor)
	return cursor, err
}

// encodeCursor returns an encoding of the given cursor suitable for a URL or a GraphQL token.
func encodeImplementationsCursor(cursor shared.ImplementationsCursor) string {
	rawEncoded, _ := json.Marshal(cursor)
	return base64.RawURLEncoding.EncodeToString(rawEncoded)
}

// decodeReferencesCursor is the inverse of encodeCursor. If the given encoded string is empty, then
// a fresh cursor is returned.
func decodeReferencesCursor(rawEncoded string) (shared.ReferencesCursor, error) {
	if rawEncoded == "" {
		return shared.ReferencesCursor{Phase: "local"}, nil
	}

	raw, err := base64.RawURLEncoding.DecodeString(rawEncoded)
	if err != nil {
		return shared.ReferencesCursor{}, err
	}

	var cursor shared.ReferencesCursor
	err = json.Unmarshal(raw, &cursor)
	return cursor, err
}

// encodeReferencesCursor returns an encoding of the given cursor suitable for a URL or a GraphQL token.
func encodeReferencesCursor(cursor shared.ReferencesCursor) string {
	rawEncoded, _ := json.Marshal(cursor)
	return base64.RawURLEncoding.EncodeToString(rawEncoded)
}

// visibleUpload pairs an upload visible from the current target commit with the
// current target path and position matched to the data within the underlying index.
type visibleUpload struct {
	Upload                shared.Dump
	TargetPath            string
	TargetPosition        shared.Position
	TargetPathWithoutRoot string
}
