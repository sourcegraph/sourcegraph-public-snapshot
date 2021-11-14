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
	f_mkMonikerID := func(m conversion.Moniker) string {
		return m.Kind + ":" + m.Scheme + ":" + m.Identifier
	}
	f_mkMonikerIDFromMoniker := func(g_docid int, g_range int) string {
		if monikers := state.Monikers.Get(g_range); monikers != nil {
			s := ""
			monikers.Each(func(g_moniker int) {
				switch state.MonikerData[g_moniker].Kind {
				case "import", "export":
					if s != "" {
						fmt.Printf("range %d has multiple import/export monikers, picked %s and will ignore the others\n", g_range, s)
						return
					}
					s = f_mkMonikerID(state.MonikerData[g_moniker])
				}
			})
			if s != "" {
				return s
			}
		}

		return ""
	}
	f_mkMonikerIDFromDef := func(g_docid int, g_range int) string {
		if s := f_mkMonikerIDFromMoniker(g_docid, g_range); s != "" {
			return s
		}

		return fmt.Sprintf(
			"%s:%d:%d", // file:line:character
			state.DocumentData[g_docid],
			state.RangeData[g_range].Start.Line,
			state.RangeData[g_range].Start.Character,
		)
	}
	f_mkMonikerIDFromRef := func(g_docid int, g_range int) string {
		if s := f_mkMonikerIDFromMoniker(g_docid, g_range); s != "" {
			return s
		}

		if g_defid := state.RangeData[g_range].DefinitionResultID; g_defid != 0 {
			s := ""
			state.DefinitionData[g_defid].Each(func(defdocid int, defrnges *datastructures.IDSet) {
				defrnges.Each(func(g_defrng int) {
					if s != "" {
						fmt.Println("multiple defs for", state.DocumentData[defdocid], "rng", g_defrng)
					} else {
						s = f_mkMonikerIDFromDef(defdocid, g_defrng)
					}
				})
			})
			if s != "" {
				return s
			}
		}

		fmt.Println("floating ref", g_docid, g_range)
		return "floating ref"
	}
	f_mkPackageID := func(pkg conversion.PackageInformation) string {
		return fmt.Sprintf("%s@%s", pkg.Name, pkg.Version)
	}
	f_mkRange := func(g_range int) *proto.Range {
		return &proto.Range{
			Start: &proto.Position{Line: int32(state.RangeData[g_range].Start.Line), Character: int32(state.RangeData[g_range].Start.Character)},
			End:   &proto.Position{Line: int32(state.RangeData[g_range].End.Line), Character: int32(state.RangeData[g_range].End.Character)},
		}
	}

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
			Kind:      g_moniker.Kind,
			Id:        f_mkMonikerID(g_moniker),
			Scheme:    g_moniker.Scheme,
			PackageId: f_mkPackageID(state.PackageInformationData[g_moniker.PackageInformationID]),
		})
	}
	// Occurrences of definitions
	g_doc_to_f_occs := map[string][]*proto.MonikerOccurrence{}
	gatherOccs := func(
		data map[int]*datastructures.DefaultIDSetMap,
		toMonikerID func(g_docid int, g_range int) string,
		role proto.MonikerOccurrence_Role,
	) {
		for _, g_docToRanges := range data {
			g_docToRanges.Each(func(g_docid int, g_ranges *datastructures.IDSet) {
				g_uri := state.DocumentData[g_docid]
				g_ranges.Each(func(g_range int) {
					f_occ := &proto.MonikerOccurrence{
						MonikerId:     toMonikerID(g_docid, g_range),
						Role:          role,
						Range:         f_mkRange(g_range),
						MarkdownHover: []string{state.HoverData[state.RangeData[g_range].HoverResultID]},
					}
					if _, ok := g_doc_to_f_occs[g_uri]; !ok {
						g_doc_to_f_occs[g_uri] = []*proto.MonikerOccurrence{}
					}
					g_doc_to_f_occs[g_uri] = append(g_doc_to_f_occs[g_uri], f_occ)
				})
			})
		}
	}
	gatherOccs(state.DefinitionData, f_mkMonikerIDFromDef, proto.MonikerOccurrence_ROLE_DEFINITION)
	gatherOccs(state.ReferenceData, f_mkMonikerIDFromRef, proto.MonikerOccurrence_ROLE_REFERENCE)
	// Documents
	for g_uri, f_occs := range g_doc_to_f_occs {
		emitDocument(&proto.Document{Uri: g_uri, Occurrences: f_occs})
	}

	// Return
	return &proto.LsifValues{Values: vals}, nil
}
