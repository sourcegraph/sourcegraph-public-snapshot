package shared

import (
	"fmt"

	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// Location is an LSP-like location scoped to a dump.
type Location struct {
	UploadID int
	Path     core.UploadRelPath
	Range    Range
}

// UsageKind is a more compact representation for SymbolUsageKind
// in the GraphQL API
type UsageKind int32

const (
	UsageKindDefinition UsageKind = iota
	UsageKindReference
	UsageKindImplementation
	UsageKindSuper
)

// RangesColumnName represents the column name in the codeintel_scip_symbols table
// that should be used when searching for a certain kind of usage.
func (k UsageKind) RangesColumnName() string {
	switch k {
	case UsageKindDefinition:
		return "definition_ranges"
	case UsageKindReference:
		return "reference_ranges"
	case UsageKindImplementation:
		return "implementation_ranges"
	case UsageKindSuper:
		return "definition_ranges"
		// For supers, we're looking for definitions of interfaces/super-class methods.
	default:
		panic(fmt.Sprintf("unhandled case for UsageKind: %v", k))
	}
}

func (k UsageKind) String() string {
	switch k {
	case UsageKindDefinition:
		return "definition"
	case UsageKindReference:
		return "reference"
	case UsageKindImplementation:
		return "implementation"
	case UsageKindSuper:
		return "super"
	default:
		panic(fmt.Sprintf("unhandled case for UsageKind: %d", int32(k)))
	}
}

// Diagnostic describes diagnostic information attached to a location within a
// particular dump.
type Diagnostic[PathType any] struct {
	UploadID int
	Path     PathType
	precise.DiagnosticData
}

func AdjustDiagnostic(d Diagnostic[core.UploadRelPath], upload shared.CompletedUpload) Diagnostic[core.RepoRelPath] {
	return Diagnostic[core.RepoRelPath]{
		UploadID:       d.UploadID,
		Path:           core.NewRepoRelPath(upload, d.Path),
		DiagnosticData: d.DiagnosticData,
	}
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
	Path         core.RepoRelPath
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

func (p Position) ToSCIPPosition() scip.Position {
	return scip.Position{
		Line:      int32(p.Line),
		Character: int32(p.Character),
	}
}

func TranslatePosition(r scip.Position) Position {
	return Position{
		Line:      int(r.Line),
		Character: int(r.Character),
	}
}

func NewRange(startLine, startCharacter, endLine, endCharacter int) Range {
	return Range{
		Start: Position{
			Line:      startLine,
			Character: startCharacter,
		},
		End: Position{
			Line:      endLine,
			Character: endCharacter,
		},
	}
}

func TranslateRange(r scip.Range) Range {
	return NewRange(int(r.Start.Line), int(r.Start.Character), int(r.End.Line), int(r.End.Character))
}

func (r Range) ToSCIPRange() scip.Range {
	return scip.NewRangeUnchecked([]int32{
		int32(r.Start.Line), int32(r.Start.Character),
		int32(r.End.Line), int32(r.End.Character),
	})
}
