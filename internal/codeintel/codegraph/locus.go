package codegraph

import (
	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
)

// Locus represents a source range within a document as found in the DB.
//
// We will eventually rename this to Location once we get rid of the
// existing Location, LocationData etc. types.
type Locus struct {
	Path  core.UploadRelPath
	Range scip.Range
}

type UploadLoci struct {
	UploadID int
	Loci     []Locus
}

type UploadSymbolLoci struct {
	UploadID int
	Symbol   string
	Loci     []Locus
}
