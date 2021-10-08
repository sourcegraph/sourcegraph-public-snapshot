package resolvers

import (
	"encoding/base64"
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// referencesCursor stores (enough of) the state of a previous References request used to
// calculate the offset into the result set to be returned by the current request.
type referencesCursor struct {
	AdjustedUploads     []cursorAdjustedUpload         `json:"adjustedUploads"`
	DefinitionUploadIDs []int                          `json:"definitionUploadIDs"`
	OrderedMonikers     []precise.QualifiedMonikerData `json:"orderedMonikers"`
	Phase               string                         `json:"phase"`
	LocalCursor         localCursor                    `json:"localCursor"`
	RemoteCursor        remoteCursor                   `json:"remoteCursor"`
}

// localCursor is an upload offset and a location offset within that upload.
type localCursor struct {
	UploadOffset int `json:"uploadOffset"`
	// The location offset within the associated upload.
	LocationOffset int `json:"locationOffset"`
}

// remoteCursor is an upload offset, the current batch of uploads, and a location offset within the batch of uploads.
type remoteCursor struct {
	UploadOffset   int   `json:"batchOffset"`
	UploadBatchIDs []int `json:"uploadBatchIDs"`
	// The location offset within the associated batch of uploads.
	LocationOffset int `json:"locationOffset"`
}

type cursorAdjustedUpload struct {
	DumpID               int                `json:"dumpID"`
	AdjustedPath         string             `json:"adjustedPath"`
	AdjustedPosition     lsifstore.Position `json:"adjustedPosition"`
	AdjustedPathInBundle string             `json:"adjustedPathInBundle"`
}

// decodeCursor is the inverse of encodeCursor. If the given encoded string is empty, then
// a fresh cursor is returned.
func decodeCursor(rawEncoded string) (referencesCursor, error) {
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

// encodeCursor returns an encoding of the given cursor suitable for a URL or a GraphQL token.
func encodeCursor(cursor referencesCursor) string {
	rawEncoded, _ := json.Marshal(cursor)
	return base64.RawURLEncoding.EncodeToString(rawEncoded)
}
