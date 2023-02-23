package conversion

import (
	"context"
	"math"
	"sort"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/conversion/datastructures"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// resultsPerResultChunk is the number of target keys in a single result chunk. This may
// not reflect the actual number of keys in a result sets, as result chunk identifiers
// are hashed into buckets based on the total number of result sets (and this value).
//
// This number does not prevent pathological cases where a single result chunk will have
// very large values, as only the number of keys (not total values within the keyspace)
// are used to determine the hashing scheme.
const resultsPerResultChunk = 512

// groupBundleData converts a raw (but canonicalized) correlation State into a GroupedBundleData.
func groupBundleData(ctx context.Context, state *State) *precise.GroupedBundleDataChans {
	numResults := len(state.DefinitionData) + len(state.ReferenceData) + len(state.ImplementationData)
	numResultChunks := int(math.Max(1, math.Floor(float64(numResults)/resultsPerResultChunk)))

	meta := precise.MetaData{NumResultChunks: numResultChunks}
	documents := serializeBundleDocuments(ctx, state)
	resultChunks := serializeResultChunks(ctx, state, numResultChunks)
	definitionRows := gatherMonikersLocations(ctx, state, state.DefinitionData, []string{"export"}, func(r Range) int { return r.DefinitionResultID })
	referenceRows := gatherMonikersLocations(ctx, state, state.ReferenceData, []string{"import", "export"}, func(r Range) int { return r.ReferenceResultID })
	implementationRows := gatherMonikersLocations(ctx, state, state.DefinitionData, []string{"implementation"}, func(r Range) int { return r.DefinitionResultID })
	packages := gatherPackages(state)
	packageReferences := gatherPackageReferences(state, packages)

	return &precise.GroupedBundleDataChans{
		ProjectRoot:       state.ProjectRoot,
		Meta:              meta,
		Documents:         documents,
		ResultChunks:      resultChunks,
		Definitions:       definitionRows,
		References:        referenceRows,
		Implementations:   implementationRows,
		Packages:          packages,
		PackageReferences: packageReferences,
	}
}

func serializeBundleDocuments(ctx context.Context, state *State) chan precise.KeyedDocumentData {
	ch := make(chan precise.KeyedDocumentData)

	go func() {
		defer close(ch)

		for documentID, uri := range state.DocumentData {
			if strings.HasPrefix(uri, "..") {
				continue
			}

			data := precise.KeyedDocumentData{
				Path:     uri,
				Document: serializeDocument(state, documentID),
			}

			select {
			case ch <- data:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch
}

func serializeDocument(state *State, documentID int) precise.DocumentData {
	document := precise.DocumentData{
		Ranges:             make(map[precise.ID]precise.RangeData, state.Contains.NumIDsForKey(documentID)),
		HoverResults:       map[precise.ID]string{},
		Monikers:           map[precise.ID]precise.MonikerData{},
		PackageInformation: map[precise.ID]precise.PackageInformationData{},
		Diagnostics:        make([]precise.DiagnosticData, 0, state.Diagnostics.NumIDsForKey(documentID)),
	}

	state.Contains.EachID(documentID, func(rangeID int) {
		rangeData := state.RangeData[rangeID]

		monikerIDs := make([]precise.ID, 0, state.Monikers.NumIDsForKey(rangeID))
		state.Monikers.EachID(rangeID, func(monikerID int) {
			moniker := state.MonikerData[monikerID]
			monikerIDs = append(monikerIDs, toID(monikerID))

			document.Monikers[toID(monikerID)] = precise.MonikerData{
				Kind:                 moniker.Kind,
				Scheme:               moniker.Scheme,
				Identifier:           moniker.Identifier,
				PackageInformationID: toID(moniker.PackageInformationID),
			}

			if moniker.PackageInformationID != 0 {
				packageInformation := state.PackageInformationData[moniker.PackageInformationID]
				document.PackageInformation[toID(moniker.PackageInformationID)] = precise.PackageInformationData{
					Manager: "",
					Name:    packageInformation.Name,
					Version: packageInformation.Version,
				}
			}
		})

		document.Ranges[toID(rangeID)] = precise.RangeData{
			StartLine:              rangeData.Start.Line,
			StartCharacter:         rangeData.Start.Character,
			EndLine:                rangeData.End.Line,
			EndCharacter:           rangeData.End.Character,
			DefinitionResultID:     toID(rangeData.DefinitionResultID),
			ReferenceResultID:      toID(rangeData.ReferenceResultID),
			ImplementationResultID: toID(rangeData.ImplementationResultID),
			HoverResultID:          toID(rangeData.HoverResultID),
			MonikerIDs:             monikerIDs,
		}

		if rangeData.HoverResultID != 0 {
			hoverData := state.HoverData[rangeData.HoverResultID]
			document.HoverResults[toID(rangeData.HoverResultID)] = hoverData
		}
	})

	state.Diagnostics.EachID(documentID, func(diagnosticID int) {
		for _, diagnostic := range state.DiagnosticResults[diagnosticID] {
			document.Diagnostics = append(document.Diagnostics, precise.DiagnosticData{
				Severity:       diagnostic.Severity,
				Code:           diagnostic.Code,
				Message:        diagnostic.Message,
				Source:         diagnostic.Source,
				StartLine:      diagnostic.StartLine,
				StartCharacter: diagnostic.StartCharacter,
				EndLine:        diagnostic.EndLine,
				EndCharacter:   diagnostic.EndCharacter,
			})
		}
	})

	return document
}

func serializeResultChunks(ctx context.Context, state *State, numResultChunks int) chan precise.IndexedResultChunkData {
	type entry struct {
		id     int
		ranges *datastructures.DefaultIDSetMap
	}
	chunkAssignments := make(map[int][]entry, numResultChunks)
	for id, ranges := range state.DefinitionData {
		index := precise.HashKey(toID(id), numResultChunks)
		chunkAssignments[index] = append(chunkAssignments[index], entry{id: id, ranges: ranges})
	}
	for id, ranges := range state.ReferenceData {
		index := precise.HashKey(toID(id), numResultChunks)
		chunkAssignments[index] = append(chunkAssignments[index], entry{id: id, ranges: ranges})
	}
	for id, ranges := range state.ImplementationData {
		index := precise.HashKey(toID(id), numResultChunks)
		chunkAssignments[index] = append(chunkAssignments[index], entry{id: id, ranges: ranges})
	}

	ch := make(chan precise.IndexedResultChunkData)

	go func() {
		defer close(ch)

		for index, entries := range chunkAssignments {
			if len(entries) == 0 {
				continue
			}

			documentPaths := map[precise.ID]string{}
			rangeIDsByResultID := make(map[precise.ID][]precise.DocumentIDRangeID, len(entries))

			for _, entry := range entries {
				rangeIDMap := map[precise.ID]int{}
				var documentIDRangeIDs []precise.DocumentIDRangeID

				entry.ranges.Each(func(documentID int, rangeIDs *datastructures.IDSet) {
					docID := toID(documentID)
					documentPaths[docID] = state.DocumentData[documentID]

					rangeIDs.Each(func(rangeID int) {
						rangeIDMap[toID(rangeID)] = rangeID

						documentIDRangeIDs = append(documentIDRangeIDs, precise.DocumentIDRangeID{
							DocumentID: docID,
							RangeID:    toID(rangeID),
						})
					})
				})

				// Sort locations by containing document path then by offset within the text
				// document (in reading order). This provides us with an obvious and deterministic
				// ordering of a result set over multiple API requests.

				sort.Sort(sortableDocumentIDRangeIDs{
					state:         state,
					documentPaths: documentPaths,
					rangeIDMap:    rangeIDMap,
					s:             documentIDRangeIDs,
				})

				rangeIDsByResultID[toID(entry.id)] = documentIDRangeIDs
			}

			data := precise.IndexedResultChunkData{
				Index: index,
				ResultChunk: precise.ResultChunkData{
					DocumentPaths:      documentPaths,
					DocumentIDRangeIDs: rangeIDsByResultID,
				},
			}

			select {
			case ch <- data:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch
}

// sortableDocumentIDRangeIDs implements sort.Interface for document/range id pairs.
type sortableDocumentIDRangeIDs struct {
	state         *State
	documentPaths map[precise.ID]string
	rangeIDMap    map[precise.ID]int
	s             []precise.DocumentIDRangeID
}

func (s sortableDocumentIDRangeIDs) Len() int      { return len(s.s) }
func (s sortableDocumentIDRangeIDs) Swap(i, j int) { s.s[i], s.s[j] = s.s[j], s.s[i] }
func (s sortableDocumentIDRangeIDs) Less(i, j int) bool {
	iDocumentID := s.s[i].DocumentID
	jDocumentID := s.s[j].DocumentID
	iRange := s.state.RangeData[s.rangeIDMap[s.s[i].RangeID]]
	jRange := s.state.RangeData[s.rangeIDMap[s.s[j].RangeID]]

	if s.documentPaths[iDocumentID] != s.documentPaths[jDocumentID] {
		return s.documentPaths[iDocumentID] <= s.documentPaths[jDocumentID]
	}

	if cmp := iRange.Start.Line - jRange.Start.Line; cmp != 0 {
		return cmp < 0
	}

	return iRange.Start.Character-jRange.Start.Character < 0
}

func gatherMonikersLocations(ctx context.Context, state *State, data map[int]*datastructures.DefaultIDSetMap, kinds []string, getResultID func(r Range) int) chan precise.MonikerLocations {
	monikers := datastructures.NewDefaultIDSetMap()
	for rangeID, r := range state.RangeData {
		if resultID := getResultID(r); resultID != 0 {
			monikers.UnionIDSet(resultID, state.Monikers.Get(rangeID))
		}
	}

	idsByKindBySchemeByIdentifier := map[string]map[string]map[string][]int{}
	for id := range data {
		monikerIDs := monikers.Get(id)
		if monikerIDs == nil {
			continue
		}

		monikerIDs.Each(func(monikerID int) {
			moniker := state.MonikerData[monikerID]
			found := false
			for _, kind := range kinds {
				if moniker.Kind == kind {
					found = true
					break
				}
			}
			if !found {
				return
			}
			idsBySchemeByIdentifier, ok := idsByKindBySchemeByIdentifier[moniker.Kind]
			if !ok {
				idsBySchemeByIdentifier = map[string]map[string][]int{}
				idsByKindBySchemeByIdentifier[moniker.Kind] = idsBySchemeByIdentifier
			}
			idsByIdentifier, ok := idsBySchemeByIdentifier[moniker.Scheme]
			if !ok {
				idsByIdentifier = map[string][]int{}
				idsBySchemeByIdentifier[moniker.Scheme] = idsByIdentifier
			}
			idsByIdentifier[moniker.Identifier] = append(idsByIdentifier[moniker.Identifier], id)
		})
	}

	ch := make(chan precise.MonikerLocations)

	go func() {
		defer close(ch)

		for kind, idsBySchemeByIdentifier := range idsByKindBySchemeByIdentifier {
			for scheme, idsByIdentifier := range idsBySchemeByIdentifier {
				for identifier, ids := range idsByIdentifier {
					var locations []precise.LocationData
					for _, id := range ids {
						data[id].Each(func(documentID int, rangeIDs *datastructures.IDSet) {
							uri := state.DocumentData[documentID]
							if strings.HasPrefix(uri, "..") {
								return
							}

							rangeIDs.Each(func(id int) {
								r := state.RangeData[id]

								locations = append(locations, precise.LocationData{
									URI:            uri,
									StartLine:      r.Start.Line,
									StartCharacter: r.Start.Character,
									EndLine:        r.End.Line,
									EndCharacter:   r.End.Character,
								})
							})
						})
					}

					if len(locations) == 0 {
						continue
					}

					// Sort locations by containing document path then by offset within the text
					// document (in reading order). This provides us with an obvious and deterministic
					// ordering of a result set over multiple API requests.

					sort.Sort(sortableLocations(locations))

					data := precise.MonikerLocations{
						Kind:       kind,
						Scheme:     scheme,
						Identifier: identifier,
						Locations:  locations,
					}

					select {
					case ch <- data:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return ch
}

// sortableLocations implements sort.Interface for locations.
type sortableLocations []precise.LocationData

func (s sortableLocations) Len() int      { return len(s) }
func (s sortableLocations) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s sortableLocations) Less(i, j int) bool {
	if s[i].URI != s[j].URI {
		return s[i].URI <= s[j].URI
	}

	if cmp := s[i].StartLine - s[j].StartLine; cmp != 0 {
		return cmp < 0
	}

	return s[i].StartCharacter < s[j].StartCharacter
}

func gatherPackages(state *State) []precise.Package {
	uniques := make(map[string]precise.Package, state.ExportedMonikers.Len())
	state.ExportedMonikers.Each(func(id int) {
		source := state.MonikerData[id]
		packageInfo := state.PackageInformationData[source.PackageInformationID]

		uniques[makeKey(source.Scheme, packageInfo.Name, packageInfo.Version)] = precise.Package{
			Scheme:  source.Scheme,
			Manager: "",
			Name:    packageInfo.Name,
			Version: packageInfo.Version,
		}
	})

	packages := make([]precise.Package, 0, len(uniques))
	for _, v := range uniques {
		packages = append(packages, v)
	}

	return packages
}

func gatherPackageReferences(state *State, packageDefinitions []precise.Package) []precise.PackageReference {
	type ExpandedPackageReference struct {
		Scheme      string
		Name        string
		Version     string
		Identifiers []string
	}

	packageDefinitionKeySet := make(map[string]struct{}, len(packageDefinitions))
	for _, pkg := range packageDefinitions {
		packageDefinitionKeySet[makeKey(pkg.Scheme, pkg.Name, pkg.Version)] = struct{}{}
	}

	uniques := make(map[string]ExpandedPackageReference, state.ImportedMonikers.Len())

	collect := func(monikers *datastructures.IDSet) {
		monikers.Each(func(id int) {
			source := state.MonikerData[id]
			packageInfo := state.PackageInformationData[source.PackageInformationID]
			key := makeKey(source.Scheme, packageInfo.Name, packageInfo.Version)

			if _, ok := packageDefinitionKeySet[key]; ok {
				// We use package definitions and references as a way to link an index
				// to its remote dependency. storing self-references is a waste of space
				// and complicates our data retention path when considering the set of
				// indexes that are referred to only by relevant/visible remote indexes.
				return
			}

			uniques[key] = ExpandedPackageReference{
				Scheme:      source.Scheme,
				Name:        packageInfo.Name,
				Version:     packageInfo.Version,
				Identifiers: append(uniques[key].Identifiers, source.Identifier),
			}
		})
	}

	collect(state.ImportedMonikers)
	collect(state.ImplementedMonikers)

	packageReferences := make([]precise.PackageReference, 0, len(uniques))
	for _, v := range uniques {
		packageReferences = append(packageReferences, precise.PackageReference{
			Package: precise.Package{
				Scheme:  v.Scheme,
				Manager: "",
				Name:    v.Name,
				Version: v.Version,
			},
		})
	}

	return packageReferences
}
