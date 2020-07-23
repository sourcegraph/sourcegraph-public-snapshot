package correlation

import (
	"math"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/correlation/datastructures"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/correlation/lsif"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bloomfilter"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/types"
)

// GroupedBundleData is a view of a correlation State that sorts data by it containing document
// and shared data into shareded result chunks. The fields of this type are what is written to
// persistent storage and what is read in the query path.
type GroupedBundleData struct {
	Meta              types.MetaData
	Documents         map[string]types.DocumentData
	ResultChunks      map[int]types.ResultChunkData
	Definitions       []types.MonikerLocations
	References        []types.MonikerLocations
	Packages          []types.Package
	PackageReferences []types.PackageReference
}

const MaxNumResultChunks = 1000
const ResultsPerResultChunk = 500

// groupBundleData converts a raw (but canonicalized) correlation State into a GroupedBundleData.
func groupBundleData(state *State, dumpID int) (*GroupedBundleData, error) {
	numResults := len(state.DefinitionData) + len(state.ReferenceData)
	numResultChunks := int(math.Min(
		MaxNumResultChunks,
		math.Max(
			1,
			math.Floor(float64(numResults)/ResultsPerResultChunk),
		),
	))

	meta := types.MetaData{NumResultChunks: numResultChunks}
	documents := serializeBundleDocuments(state)
	resultChunks := serializeResultChunks(state, numResultChunks)
	definitionRows := gatherMonikersLocations(state, state.DefinitionData, getDefinitionResultID)
	referenceRows := gatherMonikersLocations(state, state.ReferenceData, getReferenceResultID)
	packages := gatherPackages(state, dumpID)
	packageReferences, err := gatherPackageReferences(state, dumpID)
	if err != nil {
		return nil, err
	}

	return &GroupedBundleData{
		Meta:              meta,
		Documents:         documents,
		ResultChunks:      resultChunks,
		Definitions:       definitionRows,
		References:        referenceRows,
		Packages:          packages,
		PackageReferences: packageReferences,
	}, nil
}

func serializeBundleDocuments(state *State) map[string]types.DocumentData {
	out := make(map[string]types.DocumentData, len(state.DocumentData))
	for _, doc := range state.DocumentData {
		if strings.HasPrefix(doc.URI, "..") {
			continue
		}

		out[doc.URI] = serializeDocument(state, doc)
	}

	return out
}

func serializeDocument(state *State, doc lsif.Document) types.DocumentData {
	document := types.DocumentData{
		Ranges:             map[types.ID]types.RangeData{},
		HoverResults:       map[types.ID]string{},
		Monikers:           map[types.ID]types.MonikerData{},
		PackageInformation: map[types.ID]types.PackageInformationData{},
		Diagnostics:        []types.DiagnosticData{},
	}

	doc.Contains.Each(func(rangeID int) {
		k := rangeID
		v := state.RangeData[rangeID]

		monikerIDs := make([]types.ID, 0, v.MonikerIDs.Len())
		v.MonikerIDs.Each(func(m int) {
			monikerIDs = append(monikerIDs, toID(m))
		})

		document.Ranges[toID(k)] = types.RangeData{
			StartLine:          v.StartLine,
			StartCharacter:     v.StartCharacter,
			EndLine:            v.EndLine,
			EndCharacter:       v.EndCharacter,
			DefinitionResultID: toID(v.DefinitionResultID),
			ReferenceResultID:  toID(v.ReferenceResultID),
			HoverResultID:      toID(v.HoverResultID),
			MonikerIDs:         monikerIDs,
		}

		if v.HoverResultID != 0 {
			hoverData := state.HoverData[v.HoverResultID]
			document.HoverResults[toID(v.HoverResultID)] = hoverData
		}

		v.MonikerIDs.Each(func(monikerID int) {
			moniker := state.MonikerData[monikerID]
			document.Monikers[toID(monikerID)] = types.MonikerData{
				Kind:                 moniker.Kind,
				Scheme:               moniker.Scheme,
				Identifier:           moniker.Identifier,
				PackageInformationID: toID(moniker.PackageInformationID),
			}

			if moniker.PackageInformationID != 0 {
				packageInformation := state.PackageInformationData[moniker.PackageInformationID]
				document.PackageInformation[toID(moniker.PackageInformationID)] = types.PackageInformationData{
					Name:    packageInformation.Name,
					Version: packageInformation.Version,
				}
			}
		})
	})

	doc.Diagnostics.Each(func(diagnosticID int) {
		for _, diagnostic := range state.Diagnostics[diagnosticID].Result {
			document.Diagnostics = append(document.Diagnostics, types.DiagnosticData{
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

func serializeResultChunks(state *State, numResultChunks int) map[int]types.ResultChunkData {
	resultChunks := make([]types.ResultChunkData, 0, numResultChunks)
	for i := 0; i < numResultChunks; i++ {
		resultChunks = append(resultChunks, types.ResultChunkData{
			DocumentPaths:      map[types.ID]string{},
			DocumentIDRangeIDs: map[types.ID][]types.DocumentIDRangeID{},
		})
	}

	addToChunk(state, resultChunks, state.DefinitionData)
	addToChunk(state, resultChunks, state.ReferenceData)

	out := make(map[int]types.ResultChunkData, len(resultChunks))
	for id, resultChunk := range resultChunks {
		if len(resultChunk.DocumentPaths) == 0 && len(resultChunk.DocumentIDRangeIDs) == 0 {
			continue
		}

		out[id] = resultChunk
	}

	return out
}

func addToChunk(state *State, resultChunks []types.ResultChunkData, data map[int]datastructures.DefaultIDSetMap) {
	for id, documentRanges := range data {
		resultChunk := resultChunks[types.HashKey(toID(id), len(resultChunks))]

		if len(documentRanges) == 0 {
			// We may have pruned all document/ranges from a definition or reference result,
			// but we add a dummy set here so that we don't hit an unknown key during queries.
			// TODO(efritz) - remove these as part of the prune pass instead
			resultChunk.DocumentIDRangeIDs[toID(id)] = nil
		}

		for documentID, rangeIDs := range documentRanges {
			doc := state.DocumentData[documentID]
			resultChunk.DocumentPaths[toID(documentID)] = doc.URI

			rangeIDs.Each(func(rangeID int) {
				resultChunk.DocumentIDRangeIDs[toID(id)] = append(resultChunk.DocumentIDRangeIDs[toID(id)], types.DocumentIDRangeID{
					DocumentID: toID(documentID),
					RangeID:    toID(rangeID),
				})
			})
		}
	}
}

var (
	getDefinitionResultID = func(r lsif.Range) int { return r.DefinitionResultID }
	getReferenceResultID  = func(r lsif.Range) int { return r.ReferenceResultID }
)

func gatherMonikersLocations(state *State, data map[int]datastructures.DefaultIDSetMap, getResultID func(r lsif.Range) int) []types.MonikerLocations {
	monikers := datastructures.DefaultIDSetMap{}
	for _, r := range state.RangeData {
		resultID := getResultID(r)

		if resultID != 0 && r.MonikerIDs.Len() > 0 {
			monikers.GetOrCreate(resultID).Union(r.MonikerIDs)
		}
	}

	n := 0
	for id := range data {
		monikerIDs, ok := monikers[id]
		if !ok {
			continue
		}

		n += monikerIDs.Len()
	}

	uniques := make(map[string]types.MonikerLocations, n)
	for id, documentRanges := range data {
		monikerIDs, ok := monikers[id]
		if !ok {
			continue
		}

		monikerIDs.Each(func(monikerID int) {
			n := 0
			for _, rangeIDs := range documentRanges {
				n += rangeIDs.Len()
			}

			locations := make([]types.Location, 0, n)
			for documentID, rangeIDs := range documentRanges {
				document := state.DocumentData[documentID]
				if strings.HasPrefix(document.URI, "..") {
					continue
				}

				rangeIDs.Each(func(id int) {
					r := state.RangeData[id]

					locations = append(locations, types.Location{
						URI:            document.URI,
						StartLine:      r.StartLine,
						StartCharacter: r.StartCharacter,
						EndLine:        r.EndLine,
						EndCharacter:   r.EndCharacter,
					})
				})
			}

			moniker := state.MonikerData[monikerID]
			key := makeKey(moniker.Scheme, moniker.Identifier)
			uniques[key] = types.MonikerLocations{
				Scheme:     moniker.Scheme,
				Identifier: moniker.Identifier,
				Locations:  append(uniques[key].Locations, locations...),
			}
		})
	}

	monikerLocations := make([]types.MonikerLocations, 0, len(uniques))
	for _, v := range uniques {
		if len(v.Locations) > 0 {
			monikerLocations = append(monikerLocations, v)
		}
	}

	return monikerLocations
}

func gatherPackages(state *State, dumpID int) []types.Package {
	uniques := make(map[string]types.Package, state.ExportedMonikers.Len())
	state.ExportedMonikers.Each(func(id int) {
		source := state.MonikerData[id]
		packageInfo := state.PackageInformationData[source.PackageInformationID]

		uniques[makeKey(source.Scheme, packageInfo.Name, packageInfo.Version)] = types.Package{
			DumpID:  dumpID,
			Scheme:  source.Scheme,
			Name:    packageInfo.Name,
			Version: packageInfo.Version,
		}
	})

	packages := make([]types.Package, 0, len(uniques))
	for _, v := range uniques {
		packages = append(packages, v)
	}

	return packages
}

func gatherPackageReferences(state *State, dumpID int) ([]types.PackageReference, error) {
	type ExpandedPackageReference struct {
		Scheme      string
		Name        string
		Version     string
		Identifiers []string
	}

	uniques := make(map[string]ExpandedPackageReference, state.ImportedMonikers.Len())
	state.ImportedMonikers.Each(func(id int) {
		source := state.MonikerData[id]
		packageInfo := state.PackageInformationData[source.PackageInformationID]

		key := makeKey(source.Scheme, packageInfo.Name, packageInfo.Version)
		uniques[key] = ExpandedPackageReference{
			Scheme:      source.Scheme,
			Name:        packageInfo.Name,
			Version:     packageInfo.Version,
			Identifiers: append(uniques[key].Identifiers, source.Identifier),
		}
	})

	packageReferences := make([]types.PackageReference, 0, len(uniques))
	for _, v := range uniques {
		filter, err := bloomfilter.CreateFilter(v.Identifiers)
		if err != nil {
			return nil, errors.Wrap(err, "bloomfilter.CreateFilter")
		}

		packageReferences = append(packageReferences, types.PackageReference{
			DumpID:  dumpID,
			Scheme:  v.Scheme,
			Name:    v.Name,
			Version: v.Version,
			Filter:  filter,
		})
	}

	return packageReferences, nil
}

func makeKey(parts ...string) string {
	return strings.Join(parts, ":")
}

func toID(id int) types.ID {
	if id == 0 {
		return types.ID("")
	}

	return types.ID(strconv.FormatInt(int64(id), 10))
}
