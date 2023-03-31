package shared

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// Location is an LSP-like location scoped to a dump.
type Location struct {
	DumpID int
	Path   string
	Range  types.Range
}

// Diagnostic describes diagnostic information attached to a location within a
// particular dump.
type Diagnostic struct {
	DumpID int
	Path   string
	precise.DiagnosticData
}

// CodeIntelligenceRange pairs a range with its definitions, references, implementations, and hover text.
type CodeIntelligenceRange struct {
	Range           types.Range
	Definitions     []Location
	References      []Location
	Implementations []Location
	HoverText       string
}
