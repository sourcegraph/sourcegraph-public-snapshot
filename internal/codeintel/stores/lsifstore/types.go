package lsifstore

import "github.com/sourcegraph/sourcegraph/lib/codeintel/precise"

// Location is an LSP-like location scoped to a dump.
type Location struct {
	DumpID int
	Path   string
	Range  Range
}

// Range is an inclusive bounds within a file.
type Range struct {
	Start Position
	End   Position
}

// Position is a unique position within a file.
type Position struct {
	Line      int
	Character int
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
	Range           Range
	Definitions     []Location
	References      []Location
	Implementations []Location
	HoverText       string
}
