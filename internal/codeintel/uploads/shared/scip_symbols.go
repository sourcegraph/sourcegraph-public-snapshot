package shared

import (
	"sort"

	"github.com/sourcegraph/scip/bindings/go/scip"
)

type InvertedRangeIndex struct {
	SymbolName           string
	DefinitionRanges     []int32
	ReferenceRanges      []int32
	ImplementationRanges []int32
	TypeDefinitionRanges []int32
}

// ExtractSymbolIndexes creates the inverse index of symbol uses to sets of ranges within the
// given document.
func ExtractSymbolIndexes(document *scip.Document) []InvertedRangeIndex {
	rangesBySymbol := make(map[string]struct {
		definitionRanges     []*scip.Range
		referenceRanges      []*scip.Range
		implementationRanges []*scip.Range
		typeDefinitionRanges []*scip.Range
	}, len(document.Occurrences))

	for _, occurrence := range document.Occurrences {
		if occurrence.Symbol == "" || scip.IsLocalSymbol(occurrence.Symbol) {
			continue
		}

		// Get (or create) a rangeSet for this key
		rangeSet := rangesBySymbol[occurrence.Symbol]
		{
			r := scip.NewRange(occurrence.Range)

			if isDefinition := scip.SymbolRole_Definition.Matches(occurrence); isDefinition {
				rangeSet.definitionRanges = append(rangeSet.definitionRanges, r)
			} else {
				rangeSet.referenceRanges = append(rangeSet.referenceRanges, r)
			}
		}
		// Insert or update rangeSet
		rangesBySymbol[occurrence.Symbol] = rangeSet
	}

	for _, symbol := range document.Symbols {
		definitionRanges := rangesBySymbol[symbol.Symbol].definitionRanges
		if len(definitionRanges) == 0 {
			continue
		}

		for _, relationship := range symbol.Relationships {
			if !(relationship.IsImplementation || relationship.IsTypeDefinition) {
				continue
			}

			rangeSet := rangesBySymbol[relationship.Symbol]
			{
				if relationship.IsImplementation {
					rangeSet.implementationRanges = append(rangeSet.implementationRanges, definitionRanges...)
				}
				if relationship.IsTypeDefinition {
					rangeSet.typeDefinitionRanges = append(rangeSet.typeDefinitionRanges, definitionRanges...)
				}
			}
			rangesBySymbol[relationship.Symbol] = rangeSet
		}
	}

	invertedRangeIndexes := make([]InvertedRangeIndex, 0, len(rangesBySymbol))
	for symbolName, rangeSet := range rangesBySymbol {
		invertedRangeIndexes = append(invertedRangeIndexes, InvertedRangeIndex{
			SymbolName:           symbolName,
			DefinitionRanges:     collapseRanges(rangeSet.definitionRanges),
			ReferenceRanges:      collapseRanges(rangeSet.referenceRanges),
			ImplementationRanges: collapseRanges(rangeSet.implementationRanges),
			TypeDefinitionRanges: collapseRanges(rangeSet.typeDefinitionRanges),
		})
	}
	sort.Slice(invertedRangeIndexes, func(i, j int) bool {
		return invertedRangeIndexes[i].SymbolName < invertedRangeIndexes[j].SymbolName
	})

	return invertedRangeIndexes
}

// collapseRanges returns a flattened sequence of int32 components encoding the given ranges.
// The output is a concatenation of quads suitable for `types.EncodeRanges`. The output ranges
// are sorted by ascending starting position, so range sequences are also in canonical form.
func collapseRanges(ranges []*scip.Range) []int32 {
	if len(ranges) == 0 {
		return nil
	}

	rangeComponents := make([]int32, 0, len(ranges)*4)
	for _, r := range scip.SortRanges(ranges) {
		rangeComponents = append(rangeComponents, r.Start.Line, r.Start.Character, r.End.Line, r.End.Character)
	}

	return rangeComponents
}
