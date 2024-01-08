package visualization

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/grafana/regexp"

	protocolReader "github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol/reader"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/reader"
)

var quoteRe = regexp.MustCompile(`(^|[^\\]?)(")`)

type Visualizer struct {
	Context *VisualizationContext
}

func (v *Visualizer) Visualize(indexFile io.Reader, fromID, subgraphDepth int, exclude []string) error {
	if err := reader.Read(indexFile, v.Context.Stasher, nil, nil); err != nil {
		return err
	}

	forwardEdges := buildForwardGraph(v.Context.Stasher)
	backwardEdges := invertEdges(forwardEdges)
	vertices := map[int]struct{}{}
	getReachableVerticesAtDepth(fromID, forwardEdges, backwardEdges, subgraphDepth, vertices)

	fmt.Printf("digraph G {\n")

	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	_ = v.Context.Stasher.Vertices(func(lineContext reader.LineContext) bool {
		if _, ok := vertices[lineContext.Element.ID]; !ok {
			return true
		}

		if slices.Contains(exclude, lineContext.Element.Label) {
			return true
		}

		if lineContext.Element.Payload != nil {
			if err := enc.Encode(lineContext.Element.Payload); err != nil {
				fmt.Println(":bomb emoji:")
				return true
			}
			payloadStr := b.String()
			payloadStr = quoteRe.ReplaceAllString(payloadStr, `$1\"`)
			payloadStr = strings.ReplaceAll(payloadStr, "\\\\\"", "\\\"")
			payloadStr = strings.TrimSpace(payloadStr)

			fmt.Printf("\tv%d [label=\"(%d) %s %s\"];\n", lineContext.Element.ID, lineContext.Element.ID, lineContext.Element.Label, payloadStr)
			b.Reset()
		} else {
			fmt.Printf("\tv%d [label=\"(%d) %s\"];\n", lineContext.Element.ID, lineContext.Element.ID, lineContext.Element.Label)
		}
		return true
	})

	_ = v.Context.Stasher.Edges(func(lineContext reader.LineContext, edge protocolReader.Edge) bool {
		if _, ok := vertices[edge.OutV]; !ok {
			return true
		}

		vertex, _ := v.Context.Stasher.Vertex(edge.OutV)
		if slices.Contains(exclude, vertex.Element.Label) {
			return true
		}

		return forEachInV(edge, func(inV int) bool {
			if _, ok := vertices[inV]; ok {
				vertex, _ = v.Context.Stasher.Vertex(inV)
				if slices.Contains(exclude, vertex.Element.Label) {
					return true
				}
				fmt.Printf("\tv%d -> v%d [label=\"(%d) %s\"];\n", edge.OutV, inV, lineContext.Element.ID, lineContext.Element.Label)
			}

			return true
		})
	})

	fmt.Printf("}\n")
	return nil
}

func getReachableVerticesAtDepth(from int, forwardEdges, backwardEdges map[int][]int, depth int, vertices map[int]struct{}) {
	if _, ok := vertices[from]; ok || depth == 0 {
		return
	}

	vertices[from] = struct{}{}

	for _, v := range forwardEdges[from] {
		getReachableVerticesAtDepth(v, forwardEdges, backwardEdges, depth-1, vertices)
	}
	for _, v := range backwardEdges[from] {
		getReachableVerticesAtDepth(v, forwardEdges, backwardEdges, depth-1, vertices)
	}
}
