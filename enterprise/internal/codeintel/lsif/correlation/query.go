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
		return nil, errors.New("path does not exist in bundle")
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

	return QueryResult{
		Definitions: resolveLocations(bundle, rng.DefinitionResultID),
		References:  resolveLocations(bundle, rng.ReferenceResultID),
		Hover:       hover,
		Monikers:    monikers,
	}
}

func resolveLocations(bundle *GroupedBundleDataMaps, resultID lsifstore.ID) []lsifstore.LocationData {
	var locations []lsifstore.LocationData
	docIDRngIDs, chunk := getDefRef(resultID, bundle.Meta, bundle.ResultChunks)
	for _, docIDRngID := range docIDRngIDs {
		path := chunk.DocumentPaths[docIDRngID.DocumentID]
		rng := bundle.Documents[path].Ranges[docIDRngID.RangeID]
		locations = append(locations, lsifstore.LocationData{
			URI:            path,
			StartLine:      rng.StartLine,
			StartCharacter: rng.StartCharacter,
			EndLine:        rng.EndLine,
			EndCharacter:   rng.EndCharacter,
		})
	}
	return locations
}
