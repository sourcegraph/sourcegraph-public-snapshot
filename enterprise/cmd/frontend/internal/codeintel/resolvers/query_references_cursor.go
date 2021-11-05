package resolvers

import (
	"encoding/base64"
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// referencesCursor stores (enough of) the state of a previous References request used to
// calculate the offset into the result set to be returned by the current request.
type referencesCursor struct {
	AdjustedUploads []cursorAdjustedUpload         `json:"adjustedUploads"`
	OrderedMonikers []precise.QualifiedMonikerData `json:"orderedMonikers"`
	Phase           string                         `json:"phase"`
	LocalCursor     localCursor                    `json:"localCursor"`
	RemoteCursor    remoteCursor                   `json:"remoteCursor"`
}

// decodeReferencesCursor is the inverse of encodeCursor. If the given encoded string is empty, then
// a fresh cursor is returned.
func decodeReferencesCursor(rawEncoded string) (referencesCursor, error) {
	if rawEncoded == "" {
		return referencesCursor{Phase: "local"}, nil
	}

	raw, err := base64.RawURLEncoding.DecodeString(rawEncoded)
	if err != nil {
		return referencesCursor{}, err
	}

	var cursor referencesCursor
	err = json.Unmarshal(raw, &cursor)
	return cursor, err
}

// encodeReferencesCursor returns an encoding of the given cursor suitable for a URL or a GraphQL token.
func encodeReferencesCursor(cursor referencesCursor) string {
	rawEncoded, _ := json.Marshal(cursor)
	return base64.RawURLEncoding.EncodeToString(rawEncoded)
}
