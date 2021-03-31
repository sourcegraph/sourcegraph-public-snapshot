package semantic

import (
	"context"
	"sort"
)

// FindRanges filters the given ranges and returns those that contain the position constructed
// from line and character. The order of the output slice is "outside-in", so that earlier
// ranges properly enclose later ranges.
func FindRanges(ranges map[ID]RangeData, line, character int) []RangeData {
	var filtered []RangeData
	for _, r := range ranges {
		if ComparePosition(r, line, character) == 0 {
			filtered = append(filtered, r)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return ComparePosition(filtered[i], filtered[j].StartLine, filtered[j].StartCharacter) != 0
	})

	return filtered
}

// FindRangesInWIndow filters the given ranges and returns those that intersect with the
// given window of lines. Ranges are returned in reading order (top-down/left-right).
func FindRangesInWindow(ranges map[ID]RangeData, startLine, endLine int) []RangeData {
	var filtered []RangeData
	for _, r := range ranges {
		if RangeIntersectsSpan(r, startLine, endLine) {
			filtered = append(filtered, r)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return CompareRanges(filtered[i], filtered[j]) < 0
	})

	return filtered
}

// CompareRanges compares two ranges.
// Returns -1 if the range A starts before range B, or starts at the same place but ends earlier.
// Returns 0 if they're exactly equal. Returns 1 otherwise.
func CompareRanges(a RangeData, b RangeData) int {
	if a.StartLine < b.StartLine {
		return -1
	}

	if a.StartLine > b.StartLine {
		return 1
	}

	if a.StartCharacter < b.StartCharacter {
		return -1
	}

	if a.StartCharacter > b.StartCharacter {
		return 1
	}

	if a.EndLine < b.EndLine {
		return -1
	}

	if a.EndLine > b.EndLine {
		return 1
	}

	if a.EndCharacter < b.EndCharacter {
		return -1
	}

	if a.EndCharacter > b.EndCharacter {
		return 1
	}

	return 0
}

// CompareLocations compares two locations.
// Returns -1 if the range A starts before range B, or starts at the same place but ends earlier.
// Returns 0 if they're exactly equal. Returns 1 otherwise.
func CompareLocations(a LocationData, b LocationData) int {
	if a.StartLine < b.StartLine {
		return -1
	}

	if a.StartLine > b.StartLine {
		return 1
	}

	if a.StartCharacter < b.StartCharacter {
		return -1
	}

	if a.StartCharacter > b.StartCharacter {
		return 1
	}

	if a.EndLine < b.EndLine {
		return -1
	}

	if a.EndLine > b.EndLine {
		return 1
	}

	if a.EndCharacter < b.EndCharacter {
		return -1
	}

	if a.EndCharacter > b.EndCharacter {
		return 1
	}

	return 0
}

// ComparePosition compares the range r with the position constructed from line and character.
// Returns -1 if the position occurs before the range, +1 if it occurs after, and 0 if the
// position is inside of the range.
func ComparePosition(r RangeData, line, character int) int {
	if line < r.StartLine {
		return 1
	}

	if line > r.EndLine {
		return -1
	}

	if line == r.StartLine && character < r.StartCharacter {
		return 1
	}

	if line == r.EndLine && character >= r.EndCharacter {
		return -1
	}

	return 0
}

// RangeIntersectsSpan determines if the given range falls within the window denoted by the
// given start and end lines.
func RangeIntersectsSpan(r RangeData, startLine, endLine int) bool {
	return (startLine <= r.StartLine && r.StartLine < endLine) || (startLine <= r.EndLine && r.EndLine < endLine)
}

// CAUTION: Data is not deep copied.
func GroupedBundleDataMapsToChans(ctx context.Context, maps *GroupedBundleDataMaps) *GroupedBundleDataChans {
	documentChan := make(chan KeyedDocumentData, len(maps.Documents))
	go func() {
		defer close(documentChan)
		for path, doc := range maps.Documents {
			select {
			case documentChan <- KeyedDocumentData{
				Path:     path,
				Document: doc,
			}:
			case <-ctx.Done():
				return
			}
		}
	}()
	resultChunkChan := make(chan IndexedResultChunkData, len(maps.ResultChunks))
	go func() {
		defer close(resultChunkChan)

		for idx, chunk := range maps.ResultChunks {
			select {
			case resultChunkChan <- IndexedResultChunkData{
				Index:       idx,
				ResultChunk: chunk,
			}:
			case <-ctx.Done():
				return
			}
		}
	}()
	monikerDefsChan := make(chan MonikerLocations)
	go func() {
		defer close(monikerDefsChan)

		for scheme, identMap := range maps.Definitions {
			for ident, locations := range identMap {
				select {
				case monikerDefsChan <- MonikerLocations{
					Scheme:     scheme,
					Identifier: ident,
					Locations:  locations,
				}:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	monikerRefsChan := make(chan MonikerLocations)
	go func() {
		defer close(monikerRefsChan)

		for scheme, identMap := range maps.References {
			for ident, locations := range identMap {
				select {
				case monikerRefsChan <- MonikerLocations{
					Scheme:     scheme,
					Identifier: ident,
					Locations:  locations,
				}:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return &GroupedBundleDataChans{
		Meta:              maps.Meta,
		Documents:         documentChan,
		ResultChunks:      resultChunkChan,
		Definitions:       monikerDefsChan,
		References:        monikerRefsChan,
		Packages:          maps.Packages,
		PackageReferences: maps.PackageReferences,
	}
}

// CAUTION: Data is not deep copied.
func GroupedBundleDataChansToMaps(chans *GroupedBundleDataChans) *GroupedBundleDataMaps {
	documentMap := make(map[string]DocumentData)
	for keyedDocumentData := range chans.Documents {
		documentMap[keyedDocumentData.Path] = keyedDocumentData.Document
	}
	resultChunkMap := make(map[int]ResultChunkData)
	for indexedResultChunk := range chans.ResultChunks {
		resultChunkMap[indexedResultChunk.Index] = indexedResultChunk.ResultChunk
	}
	monikerDefsMap := make(map[string]map[string][]LocationData)
	for monikerDefs := range chans.Definitions {
		if _, exists := monikerDefsMap[monikerDefs.Scheme]; !exists {
			monikerDefsMap[monikerDefs.Scheme] = make(map[string][]LocationData)
		}
		monikerDefsMap[monikerDefs.Scheme][monikerDefs.Identifier] = monikerDefs.Locations
	}
	monikerRefsMap := make(map[string]map[string][]LocationData)
	for monikerRefs := range chans.References {
		if _, exists := monikerRefsMap[monikerRefs.Scheme]; !exists {
			monikerRefsMap[monikerRefs.Scheme] = make(map[string][]LocationData)
		}
		monikerRefsMap[monikerRefs.Scheme][monikerRefs.Identifier] = monikerRefs.Locations
	}

	return &GroupedBundleDataMaps{
		Meta:              chans.Meta,
		Documents:         documentMap,
		ResultChunks:      resultChunkMap,
		Definitions:       monikerDefsMap,
		References:        monikerRefsMap,
		Packages:          chans.Packages,
		PackageReferences: chans.PackageReferences,
	}
}
