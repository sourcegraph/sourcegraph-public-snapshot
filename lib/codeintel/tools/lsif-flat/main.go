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
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol/reader"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/tools/lsif-flat/lib"
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
		els := lib.ConvertFlatToGraph(&values)
		lib.WriteNDJSON(lib.ElementsToEmptyInterfaces(els), os.Stdout)
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
	state, err := conversion.CorrelateFromReader(context.Background(), reader.Dump{Reader: r, Format: reader.StandardFormat}, "")
	if err != nil {
		return nil, err
	}
	conversion.Canonicalize(state)

	// Helpers
	vals := []*proto.LsifValue{}
	emitMoniker := func(v *proto.Moniker) {
		vals = append(vals, &proto.LsifValue{Value: &proto.LsifValue_Moniker{Moniker: v}})
	}
	emitDocument := func(v *proto.Document) {
		vals = append(vals, &proto.LsifValue{Value: &proto.LsifValue_Document{Document: v}})
	}
	emitPackage := func(v *proto.Package) {
		vals = append(vals, &proto.LsifValue{Value: &proto.LsifValue_Package{Package: v}})
	}
	f_mkMonikerID := func(uri string, start protocol.Pos, end protocol.Pos) string {
		return fmt.Sprintf("%s:%d:%d-%d:%d", uri, start.Line, start.Character, end.Line, end.Character)
	}
	f_mkPackageID := func(pkg conversion.PackageInformation) string {
		return fmt.Sprintf("%s@%s", pkg.Name, pkg.Version)
	}
	_ = emitMoniker
	_ = emitDocument
	_ = emitPackage

	// Emit proto
	// Packages
	for _, g_pkg := range state.PackageInformationData {
		emitPackage(&proto.Package{
			Id:      f_mkPackageID(g_pkg),
			Name:    g_pkg.Name,
			Version: g_pkg.Version,
			Manager: "go",
		})
	}
	// Monikers
	nmonikers := 0
	for _, g_moniker := range state.MonikerData {
		nmonikers += 1
		emitMoniker(&proto.Moniker{
			Id:        g_moniker.Identifier,
			Scheme:    g_moniker.Scheme,
			PackageId: f_mkPackageID(state.PackageInformationData[g_moniker.PackageInformationID]),
		})
	}
	// Occurrences of definitions
	ndefs := 0
	nrefs := 0
	g_doc_to_f_occs := map[string][]*proto.MonikerOccurrence{}
	for _, g_docToRanges := range state.DefinitionData {
		g_docToRanges.Each(func(g_docid int, g_ranges *datastructures.IDSet) {
			g_uri := state.DocumentData[g_docid]
			g_ranges.Each(func(g_range int) {
				f_occ := &proto.MonikerOccurrence{
					MonikerId: f_mkMonikerID(g_uri, state.RangeData[g_range].Start, state.RangeData[g_range].End),
					Role:      proto.MonikerOccurrence_ROLE_DEFINITION,
					Range: &proto.Range{
						Start: &proto.Position{Line: int32(state.RangeData[g_range].Start.Line), Character: int32(state.RangeData[g_range].Start.Character)},
						End:   &proto.Position{Line: int32(state.RangeData[g_range].End.Line), Character: int32(state.RangeData[g_range].End.Character)},
					},
				}
				ndefs += 1
				if _, ok := g_doc_to_f_occs[g_uri]; !ok {
					g_doc_to_f_occs[g_uri] = []*proto.MonikerOccurrence{}
				}
				g_doc_to_f_occs[g_uri] = append(g_doc_to_f_occs[g_uri], f_occ)
			})
		})
	}
	// Occurrences of references
	for _, g_docToRanges := range state.ReferenceData {
		g_docToRanges.Each(func(g_docid int, g_ranges *datastructures.IDSet) {
			g_uri := state.DocumentData[g_docid]
			g_ranges.Each(func(g_range int) {
				f_defMonikerId := ""
				g_defid := state.RangeData[g_range].DefinitionResultID
				if g_defid == 0 {
					// no def, check monikers
					g_set := state.Monikers.Get(g_range)
					if g_set == nil {
						fmt.Println("no definition for uri", g_uri, "line", state.RangeData[g_range].Start.Line, state.RangeData[g_range].Start.Character)
						return
					}
					g_set.Each(func(monikerid int) {
						if f_defMonikerId != "" {
							fmt.Println("multiple monikers at def for", g_uri, "rng", g_range)
						} else if state.MonikerData[monikerid].Kind == "import" {
							f_defMonikerId = state.MonikerData[monikerid].Identifier
						}
					})
				} else {
					state.DefinitionData[g_defid].Each(func(defdocid int, defrnges *datastructures.IDSet) {
						defrnges.Each(func(g_defrng int) {
							if f_defMonikerId != "" {
								fmt.Println("multiple defs for", state.DocumentData[defdocid], "rng", g_defrng)
							} else {
								f_defMonikerId = f_mkMonikerID(state.DocumentData[defdocid], state.RangeData[g_defrng].Start, state.RangeData[g_defrng].End)
							}
						})
					})
				}
				if f_defMonikerId == "" {
					fmt.Println("missing definition for uri", g_uri, "rng", g_range)
					return
				}
				f_occ := &proto.MonikerOccurrence{
					MonikerId: f_defMonikerId,
					Role:      proto.MonikerOccurrence_ROLE_REFERENCE,
					Range: &proto.Range{
						Start: &proto.Position{Line: int32(state.RangeData[g_range].Start.Line), Character: int32(state.RangeData[g_range].Start.Character)},
						End:   &proto.Position{Line: int32(state.RangeData[g_range].End.Line), Character: int32(state.RangeData[g_range].End.Character)},
					},
				}
				nrefs += 1
				if _, ok := g_doc_to_f_occs[g_uri]; !ok {
					g_doc_to_f_occs[g_uri] = []*proto.MonikerOccurrence{}
				}
				g_doc_to_f_occs[g_uri] = append(g_doc_to_f_occs[g_uri], f_occ)
			})
		})
	}
	// Documents
	ndocs := 0
	for g_uri, f_occs := range g_doc_to_f_occs {
		ndocs += 1
		emitDocument(&proto.Document{Uri: g_uri, Occurrences: f_occs})
	}
	fmt.Println("Stats:")
	fmt.Println("- docs", ndocs)
	fmt.Println("- defs", ndefs)
	fmt.Println("- refs", nrefs)
	fmt.Println("- monikers", nmonikers)

	// Return
	return &proto.LsifValues{Values: vals}, nil
}
