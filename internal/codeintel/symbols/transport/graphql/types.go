package graphql

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/shared"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
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

// implementationsCursor stores (enough of) the state of a previous Implementations request used to
// calculate the offset into the result set to be returned by the current request.
type implementationsCursor struct {
	CursorsToVisibleUploads       []cursorToVisibleUpload        `json:"visibleUploads"`
	OrderedImplementationMonikers []precise.QualifiedMonikerData `json:"orderedImplementationMonikers"`
	OrderedExportMonikers         []precise.QualifiedMonikerData `json:"orderedExportMonikers"`
	Phase                         string                         `json:"phase"`
	LocalCursor                   localCursor                    `json:"localCursor"`
	RemoteCursor                  remoteCursor                   `json:"remoteCursor"`
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

// cursorAdjustedUpload
type cursorToVisibleUpload struct {
	DumpID                int             `json:"dumpID"`
	TargetPath            string          `json:"adjustedPath"`
	TargetPosition        shared.Position `json:"adjustedPosition"`
	TargetPathWithoutRoot string          `json:"adjustedPathInBundle"`
}

// adjustedUpload
// uploadsMetadata pairs an upload visible from the current target commit with the
// current target path and position matched to the data within the underlying index.
type visibleUpload struct {
	Upload                shared.Dump
	TargetPath            string
	TargetPosition        shared.Position
	TargetPathWithoutRoot string
}

type qualifiedMonikerSet struct {
	monikers       []precise.QualifiedMonikerData
	monikerHashMap map[string]struct{}
}

func newQualifiedMonikerSet() *qualifiedMonikerSet {
	return &qualifiedMonikerSet{
		monikerHashMap: map[string]struct{}{},
	}
}

// add the given qualified moniker to the set if it is distinct from all elements
// currently in the set.
func (s *qualifiedMonikerSet) add(qualifiedMoniker precise.QualifiedMonikerData) {
	monikerHash := strings.Join([]string{
		qualifiedMoniker.PackageInformationData.Name,
		qualifiedMoniker.PackageInformationData.Version,
		qualifiedMoniker.MonikerData.Scheme,
		qualifiedMoniker.MonikerData.Identifier,
	}, ":")

	if _, ok := s.monikerHashMap[monikerHash]; ok {
		return
	}

	s.monikerHashMap[monikerHash] = struct{}{}
	s.monikers = append(s.monikers, qualifiedMoniker)
}
