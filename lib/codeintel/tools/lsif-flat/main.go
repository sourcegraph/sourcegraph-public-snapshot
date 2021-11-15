package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	pb "github.com/golang/protobuf/proto"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/conversion"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/conversion/datastructures"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol/reader"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/tools/lsif-flat/proto"
)

func main() {
	// parse file into LsifValues proto
	file := os.Args[1]
	fileReader, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	if strings.HasSuffix(file, ".lsif-flat.pb") {
		bytes, err := ioutil.ReadAll(fileReader)
		if err != nil {
			panic(err)
		}
		values := proto.LsifValues{}
		pb.Unmarshal(bytes, &values)
		els := reader.ConvertFlatToGraph(&values)
		reader.WriteNDJSON(reader.ElementsToEmptyInterfaces(els), os.Stdout)
	} else if strings.HasSuffix(file, ".lsif") {
		values, err := ConvertGraphToFlat(fileReader)
		if err != nil {
			panic(err)
		}
		bytes, err := pb.Marshal(values)
		if err != nil {
			panic(err)
		}
		prefix := strings.TrimSuffix(file, ".lsif")
		out, err := os.OpenFile(prefix+".lsif-flat.pb", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			panic(err)
		}
		_, err = out.Write(bytes)
		if err != nil {
			panic(err)
		}
	}
}

func ConvertGraphToFlat(r io.Reader) (*proto.LsifValues, error) {
	// Build state
	state, err := conversion.CorrelateFromReader(context.Background(), reader.Dump{Reader: r, Format: reader.GraphFormat}, "")
	if err != nil {
		return nil, err
	}
	conversion.Canonicalize(state)
	stateMonikers := map[int][]conversion.Moniker{} // range -> []moniker
	state.Monikers.Each(func(rnge int, ids *datastructures.IDSet) {
		stateMonikers[rnge] = []conversion.Moniker{}
		ids.Each(func(id int) {
			stateMonikers[rnge] = append(stateMonikers[rnge], state.MonikerData[id])
		})
	})
	stateDefs := map[int]map[int][]conversion.Range{} // definitionResult -> doc -> []range
	for defid, docidToRanges := range state.DefinitionData {
		stateDefs[defid] = map[int][]conversion.Range{}
		docidToRanges.Each(func(docid int, ranges *datastructures.IDSet) {
			stateDefs[defid][docid] = []conversion.Range{}
			ranges.Each(func(r int) {
				stateDefs[defid][docid] = append(stateDefs[defid][docid], state.RangeData[r])
			})
		})
	}

	// Helpers
	pkgId := func(pkg conversion.PackageInformation) string {
		return fmt.Sprintf("%s@%s", pkg.Name, pkg.Version)
	}
	// Grabs the import/export moniker and implementation monikers.
	findMoniker := func(docid int, rnge int) (*proto.Moniker, []*proto.Moniker) {
		hover := []string{state.HoverData[state.RangeData[rnge].HoverResultID]}

		implementations := []*proto.Moniker{}
		for _, m := range stateMonikers[rnge] {
			if m.Kind == "implementation" {
				implementations = append(implementations, &proto.Moniker{
					Id:        m.Identifier,
					Scheme:    m.Scheme,
					PackageId: pkgId(state.PackageInformationData[m.PackageInformationID]),
					Kind:      "implementation",
				})
			}
		}
		impls := []string{}
		for _, impl := range implementations {
			impls = append(impls, impl.Id)
		}

		for _, m := range stateMonikers[rnge] {
			if m.Kind == "import" || m.Kind == "export" {
				m := &proto.Moniker{
					Id:                     m.Identifier,
					Scheme:                 m.Scheme,
					MarkdownHover:          hover,
					ImplementationMonikers: impls,
					PackageId:              pkgId(state.PackageInformationData[m.PackageInformationID]),
					Kind:                   m.Kind,
				}
				return m, implementations
			}
		}

		if defid := state.RangeData[rnge].DefinitionResultID; defid != 0 {
			for defdocid, ranges := range stateDefs[defid] {
				for _, rnge := range ranges {
					m := &proto.Moniker{
						Id:                     fmt.Sprintf("%s:%d:%d", state.DocumentData[defdocid], rnge.Start.Line, rnge.Start.Character),
						MarkdownHover:          hover,
						ImplementationMonikers: impls,
					}
					return m, implementations
				}
			}
		}

		return nil, []*proto.Moniker{}
	}

	// Collect all ranges and their monikers.
	idToKindToMoniker := map[string]map[string]*proto.Moniker{}
	rangeToMoniker := map[int]*proto.Moniker{}
	collect := func(data map[int]*datastructures.DefaultIDSetMap) {
		for _, docToRanges := range data {
			docToRanges.Each(func(doc int, ranges *datastructures.IDSet) {
				ranges.Each(func(r int) {
					if m, implementations := findMoniker(doc, r); m != nil {
						if _, ok := idToKindToMoniker[m.Id]; !ok {
							idToKindToMoniker[m.Id] = map[string]*proto.Moniker{}
						}
						idToKindToMoniker[m.Id][m.Kind] = m
						rangeToMoniker[r] = m
						for _, implementation := range implementations {
							if _, ok := idToKindToMoniker[implementation.Id]; !ok {
								idToKindToMoniker[implementation.Id] = map[string]*proto.Moniker{}
							}
							idToKindToMoniker[implementation.Id]["implementation"] = implementation
						}
					} else {
						fmt.Printf("could not construct a moniker for range %d, ignoring it\n", r)
					}
				})
			})
		}
	}
	collect(state.ReferenceData)
	collect(state.DefinitionData)

	// Fill in local implementations.
	for rnge, moniker := range rangeToMoniker {
		if docToRanges := state.ImplementationData[state.RangeData[rnge].ImplementationResultID]; docToRanges != nil {
			docToRanges.Each(func(doc int, ranges *datastructures.IDSet) {
				ranges.Each(func(rnge int) {
					moniker.ImplementationMonikers = append(moniker.ImplementationMonikers, rangeToMoniker[rnge].Id)
				})
			})
		}
	}

	// Gather occurrences
	uriToOccs := map[string][]*proto.MonikerOccurrence{}
	gatherOccs := func(data map[int]*datastructures.DefaultIDSetMap, role proto.MonikerOccurrence_Role) {
		for _, docToRanges := range data {
			docToRanges.Each(func(docid int, ranges *datastructures.IDSet) {
				uri := state.DocumentData[docid]
				if _, ok := uriToOccs[uri]; !ok {
					uriToOccs[uri] = []*proto.MonikerOccurrence{}
				}
				ranges.Each(func(rnge int) {
					r := state.RangeData[rnge]
					uriToOccs[uri] = append(uriToOccs[uri], &proto.MonikerOccurrence{
						MonikerId: rangeToMoniker[rnge].Id,
						Role:      role,
						Range: &proto.Range{
							Start: &proto.Position{Line: int32(r.Start.Line), Character: int32(r.Start.Character)},
							End:   &proto.Position{Line: int32(r.End.Line), Character: int32(r.End.Character)},
						},
						MarkdownHover: []string{state.HoverData[state.RangeData[rnge].HoverResultID]},
					})
				})
			})
		}
	}
	gatherOccs(state.DefinitionData, proto.MonikerOccurrence_ROLE_DEFINITION)
	gatherOccs(state.ReferenceData, proto.MonikerOccurrence_ROLE_REFERENCE)

	// Emit
	vals := []*proto.LsifValue{}
	for _, pkg := range state.PackageInformationData {
		vals = append(vals, &proto.LsifValue{Value: &proto.LsifValue_Package{Package: &proto.Package{
			Id:      pkgId(pkg),
			Name:    pkg.Name,
			Version: pkg.Version,
			Manager: "go",
		}}})
	}
	for _, kindToMoniker := range idToKindToMoniker {
		for _, moniker := range kindToMoniker {
			vals = append(vals, &proto.LsifValue{Value: &proto.LsifValue_Moniker{Moniker: moniker}})
		}
	}
	for uri, occs := range uriToOccs {
		vals = append(vals, &proto.LsifValue{Value: &proto.LsifValue_Document{Document: &proto.Document{Uri: uri, Occurrences: occs}}})
	}

	// Return
	return &proto.LsifValues{Values: vals}, nil
}
