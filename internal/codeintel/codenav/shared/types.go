package shared

import (
	"cmp"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// Location is an LSP-like location scoped to a dump.
type Location struct {
	UploadID int
	Path     string
	Range    Range
}

// Diagnostic describes diagnostic information attached to a location within a
// particular dump.
type Diagnostic struct {
	UploadID int
	Path     string
	precise.DiagnosticData
}

// CodeIntelligenceRange pairs a range with its definitions, references, implementations, and hover text.
type CodeIntelligenceRange struct {
	Range           Range
	Definitions     []Location
	References      []Location
	Implementations []Location
	HoverText       string
}

// UploadLocation is a path and range pair from within a particular upload. The target commit
// denotes the target commit for which the location was set (the originally requested commit).
type UploadLocation struct {
	Upload       shared.CompletedUpload
	Path         string
	TargetCommit string
	TargetRange  Range
}

type SnapshotData struct {
	DocumentOffset int
	Symbol         string
	AdditionalData []string
}

type Range struct {
	Start Position
	End   Position
}

// Position is a unique position within a file.
type Position struct {
	Line      int
	Character int
}

func (p Position) Compare(other Position) int {
	return cmp.Or(cmp.Compare(p.Line, other.Line), cmp.Compare(p.Character, other.Character))
}

// Contains checks if position is within the range inclusively for both start and end.
func (r Range) Contains(position Position) bool {
	return r.Start.Compare(position) <= 0 && r.End.Compare(position) >= 0
}

// Intersects checks if two ranges intersect inclusively for both start and end.
func (r Range) Intersects(other Range) bool {
	return r.Start.Compare(other.End) <= 0 && r.End.Compare(other.Start) >= 0
}

// Compare returns the relative order of two ranges, ranges that intersect are considered equal.
func (r Range) Compare(other Range) int {
	if r.Intersects(other) {
		return 0
	}
	return r.Start.Compare(other.Start)
}
