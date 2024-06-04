package precise

import (
	"cmp"
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

// CompareLocations compares two locations.
// Returns -1 if the range A starts before range B, or starts at the same place but ends earlier.
// Returns 0 if they're exactly equal. Returns 1 otherwise.
func CompareLocations(a LocationData, b LocationData) int {
	if v := cmp.Compare(a.StartLine, b.StartLine); v != 0 {
		return v
	}

	if v := cmp.Compare(a.StartCharacter, b.StartCharacter); v != 0 {
		return v
	}

	if v := cmp.Compare(a.EndLine, b.EndLine); v != 0 {
		return v
	}

	return cmp.Compare(a.EndCharacter, b.EndCharacter)
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
func GroupedBundleDataChansToMaps(chans *GroupedBundleDataChans) *GroupedBundleDataMaps {
	documentMap := make(map[string]DocumentData)
	for keyedDocumentData := range chans.Documents {
		documentMap[keyedDocumentData.Path] = keyedDocumentData.Document
	}
	resultChunkMap := make(map[int]ResultChunkData)
	for indexedResultChunk := range chans.ResultChunks {
		resultChunkMap[indexedResultChunk.Index] = indexedResultChunk.ResultChunk
	}
	monikerDefsMap := make(map[string]map[string]map[string][]LocationData)
	for monikerDefs := range chans.Definitions {
		if _, exists := monikerDefsMap[monikerDefs.Kind]; !exists {
			monikerDefsMap[monikerDefs.Kind] = make(map[string]map[string][]LocationData)
		}
		if _, exists := monikerDefsMap[monikerDefs.Kind][monikerDefs.Scheme]; !exists {
			monikerDefsMap[monikerDefs.Kind][monikerDefs.Scheme] = make(map[string][]LocationData)
		}
		monikerDefsMap[monikerDefs.Kind][monikerDefs.Scheme][monikerDefs.Identifier] = monikerDefs.Locations
	}
	monikerRefsMap := make(map[string]map[string]map[string][]LocationData)
	for monikerRefs := range chans.References {
		if _, exists := monikerRefsMap[monikerRefs.Kind]; !exists {
			monikerRefsMap[monikerRefs.Kind] = make(map[string]map[string][]LocationData)
		}
		if _, exists := monikerRefsMap[monikerRefs.Kind][monikerRefs.Scheme]; !exists {
			monikerRefsMap[monikerRefs.Kind][monikerRefs.Scheme] = make(map[string][]LocationData)
		}
		monikerRefsMap[monikerRefs.Kind][monikerRefs.Scheme][monikerRefs.Identifier] = monikerRefs.Locations
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
