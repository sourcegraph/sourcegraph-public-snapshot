package background

import (
	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
)

// canonicalizeDocument ensures that the fields of the given document are ordered in a
// deterministic manner (when it would not otherwise affect the data semantics). This pass
// has a two-fold benefit:
//
// (1) equivalent document payloads will share a canonical form, so they will hash to the
// same value when being inserted into the codeintel-db, and
// (2) consumers of canonical-form documents can rely on order of fields for quicker access,
// such as binary search through symbol names or occurrence ranges.
func canonicalizeDocument(document *scip.Document, externalSymbolsByName map[string]*scip.SymbolInformation) {
	// We store the relative path outside of the document payload so that renames do not
	// necessarily invalidate the document payload. When returning a SCIP document to the
	// consumer of a codeintel API, we reconstruct this relative path.
	document.RelativePath = ""

	// Denormalize external symbols into each referencing document
	injectExternalSymbols(document, externalSymbolsByName)

	// Order the remaining fields deterministically
	_ = types.CanonicalizeDocument(document)
}

// injectExternalSymbols adds symbol information objects from the external symbols into the document
// if there is an occurrence that references the external symbol name and no local symbol information
// exists.
func injectExternalSymbols(document *scip.Document, externalSymbolsByName map[string]*scip.SymbolInformation) {
	symbolNames := make(map[string]struct{}, len(document.Symbols))
	for _, symbol := range document.Symbols {
		symbolNames[symbol.Symbol] = struct{}{}
	}

	for _, occurrence := range document.Occurrences {
		if occurrence.Symbol == "" || scip.IsLocalSymbol(occurrence.Symbol) {
			continue
		}

		// Ensure we only add each symbol once
		if _, ok := symbolNames[occurrence.Symbol]; ok {
			continue
		}
		symbolNames[occurrence.Symbol] = struct{}{}

		if symbol, ok := externalSymbolsByName[occurrence.Symbol]; ok {
			document.Symbols = append(document.Symbols, symbol)
		}
	}
}
