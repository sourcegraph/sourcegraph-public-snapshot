package main

import (
	"context"
	"io"

	"github.com/dcadenas/pagerank"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/conversion"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/conversion/datastructures"
)

const (
	// Chance of following a link rather than jumping randomly.
	FOLLOW_LINK_CHANCE = 0.85
	// Smaller number here yields a more exact result; more CPU cycles required.
	TOLERANCE = 0.00001
)

type Document struct {
	docId    int
	filePath string
	rank     float64
}

// Read LSIF data from the given reader and return the files from the
// index along with their pagerank. The results are not sorted.
func PageRankLsif(indexFile io.Reader) (*[]Document, error) {
	// Build a conversion.State representing the input lsif index.
	state, err := conversion.CorrelateInMemory(context.TODO(), indexFile, "", nil)
	if err != nil {
		return nil, err
	}

	edges := addReferenceEdges(state)

	if *addImplEdges {
		addImplementationEdges(state, edges)
	}

	return runPageRanker(state, edges), nil
}

func addReferenceEdges(state *conversion.State) *map[int]int {
	// First, we need a map of range id -> doc id for the whole index.
	// Even for a very large index it's likely only to be a few million entries.
	// Since we're the only use case needing this lookup, we build it ephemerally.
	// TODO(stevey):  Can we get this info from State without preprocessing?
	rangeToDoc := make(map[int]int)
	state.Contains.Each(func(docId int, rangeIDs *datastructures.IDSet) {
		rangeIDs.Each(func(rangeId int) {
			rangeToDoc[rangeId] = docId
		})
	})

	// Walk the references and find each one's file, and also the file(s)
	// containing the definition being referenced. These two files make an edge.
	// We omit links from files referencing themselves because PageRank ignores them.
	edges := make(map[int]int)
	for _, documentMap := range state.ReferenceData {
		documentMap.Each(func(docId int, ranges *datastructures.IDSet) {
			ranges.Each(func(rangeId int) {
				// Pagerank source node is doc ID for this reference range.
				refDocId := rangeToDoc[rangeId]

				// Pagerank dest nodes are doc IDs for the associated definition range(s).
				// I.e., if a definition is split across files, they each get an edge for now.
				if data, ok := state.DefinitionData[state.RangeData[rangeId].DefinitionResultID]; ok {
					for _, defDocId := range data.UnorderedKeys() {
						// Insert a link for the PageRank calculator.
						if refDocId != defDocId {
							edges[refDocId] = defDocId
						}
					}
				}
			})
		})
	}
	return &edges
}

// Treat implementations as references to the type being implemented.
//
// Note: These edges seem to have a dramatic impact on the results. This adds in millions
// of edges and tends to push up interfaces and classes with lots of implementations. This
// leads to their effect being a bit overwhelming, compared to when only call/use references
// are included as pagerank edges. Moreover, some very strange results seem to be bubbling
// up for Java indexes (about 20% don't look like they should be in the top results).
//
// So for now this is disabled by default.
// TODO(stevey): pagerank.go shouldn't read the flag; it should be passed as a config option.
func addImplementationEdges(state *conversion.State, edges *map[int]int) {
	graph := *edges
	for _, docMap := range state.ImplementationData {
		docMap.Each(func(docId int, ranges *datastructures.IDSet) {
			// We interpret the destination vertex as the thing being implemented
			// (e.g., the definition of an interface or abstract class).
			destNode := docId
			ranges.Each(func(rangeId int) {
				if data, ok := state.ImplementationData[state.RangeData[rangeId].ImplementationResultID]; ok {
					for _, implDocId := range data.UnorderedKeys() {
						if destNode != implDocId { // skip self-references for pagerank
							graph[implDocId] = destNode
						}
					}
				}
			})
		})
	}
}

// The API to this PageRank package is that you get one shot at seeing the results.
// Rank the graph, and toss each file/rank pair into the result set.
func runPageRanker(state *conversion.State, edges *map[int]int) *[]Document {
	graph := pagerank.New()

	for srcDocId, targetDocId := range *edges {
		graph.Link(srcDocId, targetDocId)
	}

	rankings := make([]Document, 0, len(state.DocumentData))

	graph.Rank(FOLLOW_LINK_CHANCE, TOLERANCE, func(docId int, rank float64) {
		rankings = append(rankings, Document{docId: docId, filePath: state.DocumentData[docId], rank: rank})
	})
	return &rankings
}
