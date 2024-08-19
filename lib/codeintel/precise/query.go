package precise

import "github.com/sourcegraph/sourcegraph/lib/errors"

type QueryResult struct {
	Definitions []LocationData
	References  []LocationData
	Hover       string
	Monikers    []QualifiedMonikerData
}

func Query(bundle *GroupedBundleDataMaps, path string, line, character int) ([]QueryResult, error) {
	document, exists := bundle.Documents[path]
	if !exists {
		return nil, errors.New("path does not exist in bundle")
	}

	var result []QueryResult
	for _, rng := range FindRanges(document.Ranges, line, character) {
		result = append(result, Resolve(bundle, document, rng))
	}

	return result, nil
}

func Resolve(bundle *GroupedBundleDataMaps, document DocumentData, rng RangeData) QueryResult {
	hover := document.HoverResults[rng.HoverResultID]
	var monikers []QualifiedMonikerData
	for _, monikerID := range rng.MonikerIDs {
		moniker := document.Monikers[monikerID]
		monikers = append(monikers, QualifiedMonikerData{
			MonikerData:            moniker,
			PackageInformationData: document.PackageInformation[moniker.PackageInformationID],
		})
	}

	return QueryResult{
		Definitions: resolveLocations(bundle, rng.DefinitionResultID),
		References:  resolveLocations(bundle, rng.ReferenceResultID),
		Hover:       hover,
		Monikers:    monikers,
	}
}

func resolveLocations(bundle *GroupedBundleDataMaps, resultID ID) []LocationData {
	var locations []LocationData
	docIDRngIDs, chunk := getDefRef(resultID, bundle.Meta, bundle.ResultChunks)
	for _, docIDRngID := range docIDRngIDs {
		path := chunk.DocumentPaths[docIDRngID.DocumentID]
		rng := bundle.Documents[path].Ranges[docIDRngID.RangeID]
		locations = append(locations, LocationData{
			DocumentPath:   path,
			StartLine:      rng.StartLine,
			StartCharacter: rng.StartCharacter,
			EndLine:        rng.EndLine,
			EndCharacter:   rng.EndCharacter,
		})
	}
	return locations
}

func getDefRef(resultID ID, meta MetaData, resultChunks map[int]ResultChunkData) ([]DocumentIDRangeID, ResultChunkData) {
	chunkID := HashKey(resultID, meta.NumResultChunks)
	chunk := resultChunks[chunkID]
	docRngIDs := chunk.DocumentIDRangeIDs[resultID]
	return docRngIDs, chunk
}
