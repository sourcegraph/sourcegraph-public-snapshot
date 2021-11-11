package main

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol/reader"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/tools/lsif-flat/proto"
)

func main() {
	fmt.Println(convertFlatToGraph(compile()))
}

type graph struct {
	ID       int
	Elements []reader.Element
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
func (g *graph) AddEdge(label string, Payload interface{}) int {
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
		// TODO: emit resultSet (if missing)
		switch occ.Role {
		case proto.MonikerOccurrence_ROLE_DEFINITION:
			// definitionResult -> item edge -> rangeID
		case proto.MonikerOccurrence_ROLE_REFERENCE:
			// referenceResult -> item edge -> rangeID
		default:
		}
	}
	g.AddEdge("contains", reader.Edge{OutV: documentID, InVs: rangeIDs})
}
func (g *graph) AddMoniker(doc *proto.Moniker) {}
func convertFlatToGraph(vals *proto.LsifValues) []reader.Element {
	g := graph{ID: 0, Elements: []reader.Element{}}
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
