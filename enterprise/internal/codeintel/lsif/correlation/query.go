package correlation

import (
	"errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
)

type QueryResult struct {
	Definitions []lsifstore.LocationData
	References  []lsifstore.LocationData
	Hover       string
	Monikers    []lsifstore.MonikerData
}

func Query(bundle *GroupedBundleDataMaps, path string, line, character int) ([]QueryResult, error) {
	document, exists := bundle.Documents[path]
	if !exists {
		return nil, errors.New("Path does not exist in bundle.")
	}

	var result []QueryResult
	for _, rng := range lsifstore.FindRanges(document.Ranges, line, character) {
		result = append(result, Resolve(bundle, document, rng))
	}

	return result, nil
}

func Resolve(bundle *GroupedBundleDataMaps, document lsifstore.DocumentData, rng lsifstore.RangeData) QueryResult {
	hover := document.HoverResults[rng.HoverResultID]
	var monikers []lsifstore.MonikerData
	for _, monikerID := range rng.MonikerIDs {
		monikers = append(monikers, document.Monikers[monikerID])
	}

	var definitions []lsifstore.LocationData
	docIDRngIDs, chunk := getDefRef(rng.DefinitionResultID, bundle.Meta, bundle.ResultChunks)
	for _, docIDRngID := range docIDRngIDs {
		path := chunk.DocumentPaths[docIDRngID.DocumentID]
		def := bundle.Documents[path].Ranges[docIDRngID.RangeID]
		definitions = append(definitions, lsifstore.LocationData{
			URI:            path,
			StartLine:      def.StartLine,
			StartCharacter: def.StartCharacter,
			EndLine:        def.EndLine,
			EndCharacter:   def.EndCharacter,
		})
	}

	var references []lsifstore.LocationData
	docIDRngIDs, chunk = getDefRef(rng.ReferenceResultID, bundle.Meta, bundle.ResultChunks)
	for _, docIDRngID := range docIDRngIDs {
		path := chunk.DocumentPaths[docIDRngID.DocumentID]
		ref := bundle.Documents[path].Ranges[docIDRngID.RangeID]
		references = append(references, lsifstore.LocationData{
			URI:            path,
			StartLine:      ref.StartLine,
			StartCharacter: ref.StartCharacter,
			EndLine:        ref.EndLine,
			EndCharacter:   ref.EndCharacter,
		})
	}

	return QueryResult{
		Definitions: definitions,
		References:  references,
		Hover:       hover,
		Monikers:    monikers,
	}
}
