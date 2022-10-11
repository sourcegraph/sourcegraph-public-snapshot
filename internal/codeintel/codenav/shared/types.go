package shared

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

type Symbol struct {
	Name string
}

// Location is an LSP-like location scoped to a dump.
type Location struct {
	DumpID int
	Path   string
	Range  types.Range
}

type RequestArgs struct {
	RepositoryID int
	Commit       string
	Path         string
	Line         int
	Character    int
	Limit        int
	RawCursor    string
}

// DiagnosticAtUpload is a diagnostic from within a particular upload. The adjusted commit denotes
// the target commit for which the location was adjusted (the originally requested commit).
type DiagnosticAtUpload struct {
	Diagnostic
	Dump           types.Dump
	AdjustedCommit string
	AdjustedRange  types.Range
}

// Diagnostic describes diagnostic information attached to a location within a
// particular dump.
type Diagnostic struct {
	DumpID int
	Path   string
	precise.DiagnosticData
}

// AdjustedCodeIntelligenceRange stores definition, reference, and hover information for all ranges
// within a block of lines. The definition and reference locations have been adjusted to fit the
// target (originally requested) commit.
type AdjustedCodeIntelligenceRange struct {
	Range           types.Range
	Definitions     []types.UploadLocation
	References      []types.UploadLocation
	Implementations []types.UploadLocation
	HoverText       string
}

// CodeIntelligenceRange pairs a range with its definitions, references, implementations, and hover text.
type CodeIntelligenceRange struct {
	Range           types.Range
	Definitions     []Location
	References      []Location
	Implementations []Location
	HoverText       string
}

// referencesCursor stores (enough of) the state of a previous References request used to
// calculate the offset into the result set to be returned by the current request.
type ReferencesCursor struct {
	CursorsToVisibleUploads []CursorToVisibleUpload        `json:"adjustedUploads"`
	OrderedMonikers         []precise.QualifiedMonikerData `json:"orderedMonikers"`
	Phase                   string                         `json:"phase"`
	LocalCursor             LocalCursor                    `json:"localCursor"`
	RemoteCursor            RemoteCursor                   `json:"remoteCursor"`
}

// ImplementationsCursor stores (enough of) the state of a previous Implementations request used to
// calculate the offset into the result set to be returned by the current request.
type ImplementationsCursor struct {
	CursorsToVisibleUploads       []CursorToVisibleUpload        `json:"visibleUploads"`
	OrderedImplementationMonikers []precise.QualifiedMonikerData `json:"orderedImplementationMonikers"`
	OrderedExportMonikers         []precise.QualifiedMonikerData `json:"orderedExportMonikers"`
	Phase                         string                         `json:"phase"`
	LocalCursor                   LocalCursor                    `json:"localCursor"`
	RemoteCursor                  RemoteCursor                   `json:"remoteCursor"`
}

// cursorAdjustedUpload
type CursorToVisibleUpload struct {
	DumpID                int            `json:"dumpID"`
	TargetPath            string         `json:"adjustedPath"`
	TargetPosition        types.Position `json:"adjustedPosition"`
	TargetPathWithoutRoot string         `json:"adjustedPathInBundle"`
}

// localCursor is an upload offset and a location offset within that upload.
type LocalCursor struct {
	UploadOffset int `json:"uploadOffset"`
	// The location offset within the associated upload.
	LocationOffset int `json:"locationOffset"`
}

// RemoteCursor is an upload offset, the current batch of uploads, and a location offset within the batch of uploads.
type RemoteCursor struct {
	UploadOffset   int   `json:"batchOffset"`
	UploadBatchIDs []int `json:"uploadBatchIDs"`
	// The location offset within the associated batch of uploads.
	LocationOffset int `json:"locationOffset"`
}
