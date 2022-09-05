package conversion

import (
	"sort"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/conversion/datastructures"
)

// canonicalize deduplicates data in the raw correlation state and collapses range,
// result set, and moniker data that form chains via next edges.
func canonicalize(state *State) {
	fns := []func(state *State){
		canonicalizeDocuments,
		canonicalizeReferenceResults,
		canonicalizeResultSets,
		canonicalizeRanges,
	}

	for _, fn := range fns {
		fn(state)
	}
}

// canonicalizeDocuments determines if multiple documents are defined with the same URI. This can
// happen in some indexers (such as lsif-tsc) that index dependent projects into the same index
// as the target project. For each set of documents that share a path, we choose one document to
// be the canonical representative and merge the contains, definition, and reference data into the
// unique canonical document. This function guarantees that duplicate document IDs are removed from
// the correlation state.
func canonicalizeDocuments(state *State) {
	documentIDs := map[string][]int{}
	for documentID, uri := range state.DocumentData {
		documentIDs[uri] = append(documentIDs[uri], documentID)
	}
	for _, v := range documentIDs {
		sort.Ints(v)
	}

	canonicalIDs := make(map[int]int, len(state.DocumentData))
	for documentID, uri := range state.DocumentData {
		// Choose canonical document alphabetically
		if canonicalID := documentIDs[uri][0]; documentID != canonicalID {
			canonicalIDs[documentID] = canonicalID
		}
	}

	// Replace references to each document with the canonical references
	canonicalizeDocumentsInDefinitionReferences(state.DefinitionData, canonicalIDs)
	canonicalizeDocumentsInDefinitionReferences(state.ReferenceData, canonicalIDs)
	canonicalizeDocumentsInDefinitionReferences(state.ImplementationData, canonicalIDs)

	for documentID, canonicalID := range canonicalIDs {
		// Move ranges and diagnostics into the canonical document
		state.Contains.UnionIDSet(canonicalID, state.Contains.Get(documentID))
		state.Diagnostics.UnionIDSet(canonicalID, state.Diagnostics.Get(documentID))

		// Remove non-canonical documents
		delete(state.DocumentData, documentID)
		state.Contains.Delete(documentID)
		state.Diagnostics.Delete(documentID)
	}
}

// canonicalizeDocumentsInDefinitionReferences moves definition, reference, and implementation result
// data from a document to its canonical document (if they differ) and removes all references to the
// non-canonical document.
func canonicalizeDocumentsInDefinitionReferences(definitionReferenceData map[int]*datastructures.DefaultIDSetMap, canonicalIDs map[int]int) {
	for _, documentRanges := range definitionReferenceData {
		// The length of documentRanges will always be less than or equal to
		// the length of canonicalIDs, since canonicalIDs will have one entry
		// for each document. So iterate over documentRanges instead of
		// iterating over canonicalIDs.

		// Copy out keys first instead of (incorrectly) iterating over documentRanges while modifying it
		var documentIDs = documentRanges.UnorderedKeys()
		for _, documentID := range documentIDs {
			canonicalID, ok := canonicalIDs[documentID]
			if !ok {
				continue
			}
			// Remove def/ref data from non-canonical document...
			rangeIDs := documentRanges.Pop(documentID)
			// ...and merge it with the data for the canonical document.
			documentRanges.UnionIDSet(canonicalID, rangeIDs)
		}
	}
}

// canonicalizeReferenceResults determines which reference results refer to another reference result.
// We denormalize the data so that all ranges reachable from set A are also reachable from set B when
// B is linked to A via an item edge.
func canonicalizeReferenceResults(state *State) {
	visited := map[int]struct{}{}

	var visit func(state *State, id int)
	visit = func(state *State, id int) {
		if _, ok := visited[id]; ok {
			return
		}
		visited[id] = struct{}{}

		nextIDs, ok := state.LinkedReferenceResults[id]
		if !ok {
			return
		}

		for _, nextID := range nextIDs {
			visit(state, nextID)

			// Copy data from the referenced to the referencing set
			state.ReferenceData[nextID].Each(func(documentID int, rangeIDs *datastructures.IDSet) {
				state.ReferenceData[id].UnionIDSet(documentID, rangeIDs)
			})
		}
	}

	for id := range state.ReferenceData {
		visit(state, id)
	}
}

// canonicalizeResultSets runs canonicalizeResultSet on each result set in the correlation state.
// This will collapse result sets down recursively so that if a result set's next element also has
// a next element, then both sets merge down into the original result set.
func canonicalizeResultSets(state *State) {
	for resultSetID, resultSetData := range state.ResultSetData {
		canonicalizeResultSetData(state, resultSetID, resultSetData)
	}

	for resultSetID := range state.ResultSetData {
		state.Monikers.UnionIDSet(resultSetID, gatherMonikers(state, state.Monikers.Get(resultSetID)))
	}
}

// canonicalizeResultSets "merges down" the definition, reference, and hover result identifiers
// from the element's "next" result set if such an element exists and the identifier is not already.
// defined. This also merges down the moniker ids by unioning the sets.
//
// This method is assumed to be invoked only after canonicalizeResultSets, otherwise the next element
// of a range may not have all of the necessary data to perform this canonicalization step.
func canonicalizeRanges(state *State) {
	for rangeID, rangeData := range state.RangeData {
		if nextID, nextItem, ok := next(state, rangeID); ok {
			// Merge range and next element
			rangeData = mergeNextRangeData(state, rangeID, rangeData, nextID, nextItem)
			// Delete next data to prevent us from re-performing this step
			delete(state.NextData, rangeID)
		}

		state.RangeData[rangeID] = rangeData
		state.Monikers.UnionIDSet(rangeID, gatherMonikers(state, state.Monikers.Get(rangeID)))
	}
}

// canonicalizeResultSets "merges down" the definition, reference, and hover result identifiers
// from the element's "next" result set if such an element exists and the identifier is not
// already defined. This also merges down the moniker ids by unioning the sets.
func canonicalizeResultSetData(state *State, id int, item ResultSet) ResultSet {
	if nextID, nextItem, ok := next(state, id); ok {
		// Recursively canonicalize the next element
		nextItem = canonicalizeResultSetData(state, nextID, nextItem)
		// Merge result set and canonicalized next element
		item = mergeNextResultSetData(state, id, item, nextID, nextItem)
		// Delete next data to prevent us from re-performing this step
		delete(state.NextData, id)
	}

	state.ResultSetData[id] = item
	return item
}

// mergeNextResultSetData merges the definition, reference, and hover result identifiers from
// nextItem into item when not already defined. The moniker identifiers of nextItem are unioned
// into the moniker identifiers of item.
func mergeNextResultSetData(state *State, itemID int, item ResultSet, nextID int, nextItem ResultSet) ResultSet {
	if item.DefinitionResultID == 0 {
		item = item.SetDefinitionResultID(nextItem.DefinitionResultID)
	}
	if item.ReferenceResultID == 0 {
		item = item.SetReferenceResultID(nextItem.ReferenceResultID)
	}
	if item.ImplementationResultID == 0 {
		item = item.SetImplementationResultID(nextItem.ImplementationResultID)
	}
	if item.HoverResultID == 0 {
		item = item.SetHoverResultID(nextItem.HoverResultID)
	}

	state.Monikers.UnionIDSet(itemID, state.Monikers.Get(nextID))
	return item
}

// mergeNextRangeData merges the definition, reference, and hover result identifiers from nextItem
// into item when not already defined. The moniker identifiers of nextItem are unioned into the
// moniker identifiers of item.
func mergeNextRangeData(state *State, itemID int, item Range, nextID int, nextItem ResultSet) Range {
	if item.DefinitionResultID == 0 {
		item = item.SetDefinitionResultID(nextItem.DefinitionResultID)
	}
	if item.ReferenceResultID == 0 {
		item = item.SetReferenceResultID(nextItem.ReferenceResultID)
	}
	if item.ImplementationResultID == 0 {
		item = item.SetImplementationResultID(nextItem.ImplementationResultID)
	}
	if item.HoverResultID == 0 {
		item = item.SetHoverResultID(nextItem.HoverResultID)
	}

	state.Monikers.UnionIDSet(itemID, state.Monikers.Get(nextID))
	return item
}

// gatherMonikers returns a new set of moniker identifiers based off the given set. The returned
// set will additionall contain the transitive closure of all moniker identifiers linked to any
// moniker identifier in the original set. This ignores adding any local-kind monikers to the new
// set.
func gatherMonikers(state *State, source *datastructures.IDSet) *datastructures.IDSet {
	if source == nil || source.Len() == 0 {
		return nil
	}

	monikers := datastructures.NewIDSet()

	source.Each(func(sourceID int) {
		state.LinkedMonikers.ExtractSet(sourceID).Each(func(id int) {
			if state.MonikerData[id].Kind != "local" {
				monikers.Add(id)
			}
		})
	})

	return monikers
}

// next returns the "next" identifier and result set element for the given identifier, if one exists.
func next(state *State, id int) (int, ResultSet, bool) {
	nextID, ok := state.NextData[id]
	if !ok {
		return 0, ResultSet{}, false
	}

	return nextID, state.ResultSetData[nextID], true
}
