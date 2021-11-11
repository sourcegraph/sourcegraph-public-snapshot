package main

import (
	"fmt"
	"os"

	pb "github.com/golang/protobuf/proto"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol/reader"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol/writer"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/tools/lsif-flat/proto"
)

func main() {
	// parse file into LsifValues proto
	file := os.Args[1]
	bytes, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}
	values := proto.LsifValues{}
	pb.Unmarshal(bytes, &values)
	fmt.Println("PROTOBUF", file, &values)
	g := convertFlatToGraph(&values)
	// bytes := proto.ReadFile(file)

	// g := convertFlatToGraph(compile())
	writeGraphToJSON(g, writer.NewJSONWriter(os.Stdout))
}

type ResultIDs struct {
	ResultSet        int
	DefinitionResult int
	ReferenceResult  int
}

type graph struct {
	ID       int
	Elements []reader.Element
	idCache  map[string]ResultIDs
}

func (g *graph) ResultIDs(moniker string) ResultIDs {
	ids, ok := g.idCache[moniker]
	if !ok {
		ids = ResultIDs{
			ResultSet:        g.AddVertex("resultSet", reader.ResultSet{}),
			DefinitionResult: g.AddVertex("definitionResult", nil),
			ReferenceResult:  g.AddVertex("referenceResult", nil),
		}
		g.AddEdge("textDocument/definition", reader.Edge{OutV: ids.ResultSet, InV: ids.DefinitionResult})
		g.AddEdge("textDocument/references", reader.Edge{OutV: ids.ResultSet, InV: ids.ReferenceResult})
		g.idCache[moniker] = ids
	}
	return ids
}
func (g *graph) Add(Type, Label string, Payload interface{}) int {
	g.ID++
	g.Elements = append(g.Elements, reader.Element{
		ID:      g.ID,
		Type:    Type,
		Label:   Label,
		Payload: Payload,
	})
	return g.ID
}
func (g *graph) AddVertex(label string, Payload interface{}) int {
	return g.Add("vertex", label, Payload)
}
func (g *graph) AddEdge(label string, Payload reader.Edge) int {
	return g.Add("edge", label, Payload)
}
func (g *graph) AddPackage(doc *proto.Package) {}
func (g *graph) AddDocument(doc *proto.Document) {
	documentID := g.AddVertex("document", doc.Uri)
	rangeIDs := []int{}
	for _, occ := range doc.Occurrences {
		rangeID := g.AddVertex("range", reader.Range{
			RangeData: protocol.RangeData{
				Start: protocol.Pos{
					Line:      int(occ.Range.Start.Line),
					Character: int(occ.Range.Start.Character),
				},
				End: protocol.Pos{
					Line:      int(occ.Range.End.Line),
					Character: int(occ.Range.End.Character),
				},
			},
		})
		rangeIDs = append(rangeIDs, rangeID)
		ids := g.ResultIDs(occ.MonikerId)
		switch occ.Role {
		case proto.MonikerOccurrence_ROLE_DEFINITION:
			g.AddEdge("item", reader.Edge{OutV: ids.DefinitionResult, InV: rangeID, Document: documentID})
		case proto.MonikerOccurrence_ROLE_REFERENCE:
			g.AddEdge("item", reader.Edge{OutV: ids.ReferenceResult, InV: rangeID, Document: documentID})
		default:
		}
	}
	g.AddEdge("contains", reader.Edge{OutV: documentID, InVs: rangeIDs})
}
func (g *graph) AddMoniker(doc *proto.Moniker) {}
func writeGraphToJSON(elements []reader.Element, writer writer.JSONWriter) {
	for _, e := range elements {
		// TODO: marshall with switch
		writer.Write(e)
	}
	writer.Flush()
}
func convertFlatToGraph(vals *proto.LsifValues) []reader.Element {
	g := graph{ID: 0, Elements: []reader.Element{}, idCache: map[string]ResultIDs{}}
	g.AddVertex(
		"metaData",
		reader.MetaData{
			Version:     "0.1.0",
			ProjectRoot: "file:///",
		},
	)
	for _, lsifValue := range vals.Values {
		switch value := lsifValue.Value.(type) {
		case *proto.LsifValue_Package:
			g.AddPackage(value.Package)
		case *proto.LsifValue_Document:
			g.AddDocument(value.Document)
		case *proto.LsifValue_Moniker:
			g.AddMoniker(value.Moniker)
		default:
		}

	}
	return g.Elements
}

func compile() *proto.LsifValues {
	vals := []*proto.LsifValue{}

	s := "miso cat miso "
	word := ""
	doc := proto.Document{}
	defs := map[string]int{}
	for i, char := range s {
		role := proto.MonikerOccurrence_ROLE_REFERENCE
		if char == ' ' {
			if _, ok := defs[word]; !ok {
				role = proto.MonikerOccurrence_ROLE_DEFINITION
				defs[word] = i
				vals = append(vals, &proto.LsifValue{Value: &proto.LsifValue_Moniker{Moniker: &proto.Moniker{
					Id:            word,
					MarkdownHover: []string{fmt.Sprintf("Good %s! 🥰", word)},
				}}})
			}
			// TODO proto stuff
			doc.Occurrences = append(doc.Occurrences, &proto.MonikerOccurrence{
				MonikerId: word,
				Role:      role,
				Range: &proto.Range{
					Start: &proto.Position{Line: 0, Character: int32(i - len(word))},
					End:   &proto.Position{Line: 0, Character: int32(i)},
				},
			})
			word = ""
		} else {
			word = word + string(char)
		}
	}
	vals = append(vals, &proto.LsifValue{Value: &proto.LsifValue_Document{Document: &doc}})
	return &proto.LsifValues{Values: vals}
}
