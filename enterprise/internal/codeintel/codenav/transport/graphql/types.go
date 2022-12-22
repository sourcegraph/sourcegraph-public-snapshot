package graphql

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// AdjustedLocation is a path and range pair from within a particular upload. The adjusted commit
// denotes the target commit for which the location was adjusted (the originally requested commit).
type AdjustedLocation struct {
	Dump           types.Dump
	Path           string
	AdjustedCommit string
	AdjustedRange  types.Range
}

// AdjustedCodeIntelligenceRange stores definition, reference, and hover information for all ranges
// within a block of lines. The definition and reference locations have been adjusted to fit the
// target (originally requested) commit.
type AdjustedCodeIntelligenceRange struct {
	Range           types.Range
	Definitions     []AdjustedLocation
	References      []AdjustedLocation
	Implementations []AdjustedLocation
	HoverText       string
}

// AdjustedDiagnostic is a diagnostic from within a particular upload. The adjusted commit denotes
// the target commit for which the location was adjusted (the originally requested commit).
type AdjustedDiagnostic struct {
	DumpID int
	Path   string
	precise.DiagnosticData

	Dump           types.Dump
	AdjustedCommit string
	AdjustedRange  types.Range
}
