package resolvers

import (
	"encoding/base64"
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// implementationsCursor stores (enough of) the state of a previous Implementations request used to
// calculate the offset into the result set to be returned by the current request.
type implementationsCursor struct {
	AdjustedUploads     []cursorAdjustedUpload         `json:"adjustedUploads"`
	DefinitionUploadIDs []int                          `json:"definitionUploadIDs"`
	OrderedMonikers     []precise.QualifiedMonikerData `json:"orderedMonikers"`
	Phase               string                         `json:"phase"`
	LocalCursor         localCursor                    `json:"localCursor"`
	RemoteCursor        remoteCursor                   `json:"remoteCursor"`
}

// decodeCursor is the inverse of encodeCursor. If the given encoded string is empty, then
// a fresh cursor is returned.
func decodeImplementationsCursor(rawEncoded string) (implementationsCursor, error) {
	if rawEncoded == "" {
		return implementationsCursor{Phase: "local"}, nil
	}

	raw, err := base64.RawURLEncoding.DecodeString(rawEncoded)
	if err != nil {
		return implementationsCursor{}, err
	}

	var cursor implementationsCursor
	err = json.Unmarshal(raw, &cursor)
	return cursor, err
}

// encodeCursor returns an encoding of the given cursor suitable for a URL or a GraphQL token.
func encodeImplementationsCursor(cursor implementationsCursor) string {
	rawEncoded, _ := json.Marshal(cursor)
	return base64.RawURLEncoding.EncodeToString(rawEncoded)
}
