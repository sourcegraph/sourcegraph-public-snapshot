package conversion

import (
	"fmt"
	"strings"

	pb "github.com/golang/protobuf/proto"

	lsifTyped "github.com/sourcegraph/sourcegraph/lib/codeintel/lsif_typed"
)

func ToGraph(payload []byte) error {
	var index lsifTyped.Index
	if err := pb.Unmarshal(payload, &index); err != nil {
		return err
	}

	symbols := map[string]*lsifTyped.SymbolInformation{}
	for _, symbol := range index.ExternalSymbols {
		symbols[symbol.Symbol] = symbol
	}
	for _, document := range index.Document {
		for _, symbol := range document.Symbols {
			symbols[symbol.Symbol] = symbol
		}
	}

	fmt.Printf("> metadata vertex %v", "file:///"+index.Metadata.ProjectRoot) // TODO - emit (+ tool info)

	monikerIDs := map[string]int{}
	for name, info := range symbols {
		glbl++
		monikerID := glbl
		fmt.Printf("> moniker vertex %v %v", monikerID, name) // TODO - emit
		monikerIDs[name] = monikerID
		// TODO
		_ = info
		// if info.Package != "" {
		// TODO - emit package information
		// }
	}

	for _, document := range index.Document {
		fmt.Printf("> document vertex %v", "file:///"+document.RelativePath) // TODO - emit

		glbl++
		documentID := glbl
		rangeIDs := make([]int, 0, len(document.Occurrences))

		for _, occurrence := range document.Occurrences {
			glbl++
			rangeID := glbl
			fmt.Printf("> range vertex %v %v", rangeID, normalizeRange(occurrence.Range))
			rangeIDs = append(rangeIDs, rangeID)

			glbl++
			resultSetID := glbl
			fmt.Printf("> result set %v", resultSetID)            // TODO - emit
			fmt.Printf("> next edge %v %v", rangeID, resultSetID) // TODO - emit

			if documentation := symbols[occurrence.Symbol].Documentation; len(documentation) != 0 {
				glbl++
				hoverResultID := glbl
				fmt.Printf("> hoverResult vertex %v %v", hoverResultID, strings.Join(documentation, "\n\n")) // TODO - emit
				fmt.Printf("> textDocument/hover edge", resultSetID, hoverResultID)                          // TODO - emit
			}

			if occurrence.SymbolRoles&int32(lsifTyped.SymbolRole_Definition) != 0 {
				glbl++
				definitionResultID := glbl
				fmt.Printf("> definitionResult vertex %v", definitionResultID)                      // TODO - emit
				fmt.Printf("> textDocument/definitionResult edge", resultSetID, definitionResultID) // TODO - emit
				fmt.Printf("> item edge %v %v (%v)", definitionResultID, rangeID, documentID)       // TODO - emit
			} else {
				glbl++
				referenceResultID := glbl
				fmt.Printf("> referenceResult vertex %v", referenceResultID)                      // TODO - emit
				fmt.Printf("> textDocument/referenceResult edge", resultSetID, referenceResultID) // TODO
				fmt.Printf("> item edge %v %v (%v)", referenceResultID, rangeID, documentID)      // TODO - emit
			}

			if monikerID, ok := monikerIDs[occurrence.Symbol]; ok {
				fmt.Printf("> moniker edge", resultSetID, monikerID)
			}

			// TODO - condition
			if false {
				glbl++
				implementationResultID := glbl
				fmt.Printf("> implementationResult vertex %v", implementationResultID)                      // TODO - emit
				fmt.Printf("> textDocument/implementationResult edge", resultSetID, implementationResultID) // TODO - emit
				fmt.Printf("> item edge %v %v (%v)", implementationResultID, rangeID, documentID)           // TODO - emit

				// TODO - relationship not complete
			}
		}

		fmt.Printf("> contains edge %v %v", documentID, rangeIDs) // TODO - emit
	}

	fmt.Printf("PLAYGROUND\n")
	return nil
}

func normalizeRange(elements []int32) []int32 {
	if len(elements) == 3 {
		return []int32{elements[0], elements[1], elements[0], elements[2]}
	}

	return elements
}

var glbl = 1
