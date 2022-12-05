package types

import "github.com/sourcegraph/scip/bindings/go/scip"

// TODO - document
func FlattenDocuments(documents []*scip.Document) []*scip.Document {
	documentMap := make(map[string]*scip.Document, len(documents))
	for _, document := range documents {
		if _, ok := documentMap[document.RelativePath]; !ok {
			documentMap[document.RelativePath] = document
			continue
		}

		// TODO - merge
	}

	flattened := make([]*scip.Document, 0, len(documentMap))
	for _, document := range documentMap {
		flattened = append(flattened, document)
	}

	return flattened
}

// TODO - document
func FlattenSymbols(symbols []*scip.SymbolInformation) []*scip.SymbolInformation {
	symbolMap := make(map[string]*scip.SymbolInformation, len(symbols))
	for _, symbol := range symbols {
		if _, ok := symbolMap[symbol.Symbol]; !ok {
			symbolMap[symbol.Symbol] = symbol
			continue
		}

		// TODO - merge
	}

	flattened := make([]*scip.SymbolInformation, 0, len(symbolMap))
	for _, symbol := range symbolMap {
		flattened = append(flattened, symbol)
	}

	return flattened
}

// TODO - document
func FlattenOccurrences(occurrences []*scip.Occurrence) []*scip.Occurrence {
	// TODO - implement
	return occurrences
}
