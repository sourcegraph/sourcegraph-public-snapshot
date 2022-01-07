package main

import (
	"bytes"
	"compress/gzip"
	"fmt"

	"github.com/gogo/protobuf/jsonpb"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif_typed"
)

func main() {
		fmt.Println("max", "compressed-json", "compressed-proto", "compressed-json-proto", "optimal")
	for _, max := range []int{10, 100, 1_000, 10_000, 100_000} {
		doc := &lsif_typed.Document{}
		doc2 := &lsif_typed.Document{}
		for i := 0; i < max; i++ {
			occ := &lsif_typed.MonikerOccurrence{
				Range: &lsif_typed.Range{
					Start: &lsif_typed.Position{Line: int32(i), Character: int32(1)},
					End:   &lsif_typed.Position{Line: int32(i), Character: int32(11)},
				},
				Role:      lsif_typed.MonikerOccurrence_ROLE_DEFINITION,
				Highlight: lsif_typed.MonikerOccurrence_HIGHLIGHT_IDENTIFIER,
			}
			doc.Occurrences = append(doc.Occurrences, occ)
			occ2 := &lsif_typed.MonikerOccurrence{
				RangeCompressed: []int32{int32(i), 1, 10},
				Role:            lsif_typed.MonikerOccurrence_ROLE_DEFINITION,
				Highlight:       lsif_typed.MonikerOccurrence_HIGHLIGHT_IDENTIFIER,
			}
			doc2.Occurrences = append(doc2.Occurrences, occ2)
		}
		json1 := jsonEncode(doc)
		json2 := jsonEncode(doc2)
		proto1 := protoEncode(doc)
		proto2 := protoEncode(doc2)
		fmt.Println(
			max,
			float64(json1)/float64(json2),
			float64(proto1)/float64(proto2),
			float64(json2)/float64(proto2),
			float64(json2)/float64(proto2),
		)
	}
}

func jsonEncode(doc *lsif_typed.Document) int {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	marshaller := &jsonpb.Marshaler{EnumsAsInts: true, EmitDefaults: false}
	marshaller.Marshal(writer, doc)
	writer.Close()
	return buf.Len()
}

func protoEncode(doc *lsif_typed.Document) int {
	protoBytes, err := proto.Marshal(doc)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	writer.Write(protoBytes)
	writer.Close()
	return buf.Len()
}
