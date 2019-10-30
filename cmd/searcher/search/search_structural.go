package search

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/comby"
)

// XXX TODO: fileMatchLimit
func structuralSearch(ctx context.Context, pattern string, zipPath string, fileMatchLimit int, onlyFiles []string) (matches []protocol.FileMatch, limitHit bool, err error) {

	b := new(bytes.Buffer)
	w := bufio.NewWriter(b)

	// XXX only do structural if we have onlyFiles to search. But remove this later to enable unindexed search
	if len(onlyFiles) == 0 {
		return matches, limitHit, err
	}

	fmt.Printf("Only files: %s\n", strings.Join(onlyFiles, ","))

	args := comby.Args{
		Input:         comby.ZipPath(zipPath),
		MatchTemplate: pattern,
		MatchOnly:     true,
		FilePatterns:  onlyFiles,
		NumWorkers:    numWorkers,
	}
	err = comby.PipeTo(args, w)
	if err != nil {
		return nil, false, err
	}

	scanner := bufio.NewScanner(b)
	scanner.Buffer(make([]byte, 100), 10*bufio.MaxScanTokenSize)
	var combyMatches []comby.FileMatch
	for scanner.Scan() {
		b := scanner.Bytes()
		if err := scanner.Err(); err != nil {
			fmt.Printf("Error scanning: %v", err)
			// skip scanner errors
			continue
		}
		var r *comby.FileMatch
		fmt.Printf("Received: %s\n", string(b))
		if err := json.Unmarshal(b, &r); err != nil {
			fmt.Printf("Error unmarshaling: %v", err)
			// skip decode errors
			continue
		}
		combyMatches = append(combyMatches, *r)
	}

	fmt.Printf("Done searching. Num matches: %d\n", len(combyMatches))

	for _, m := range combyMatches {
		var lineMatches []protocol.LineMatch
		for _, r := range m.Matches {
			lineMatch := protocol.LineMatch{
				LineNumber: r.Range.Start.Line - 1,
				// XXX sigh. assume one match per line.
				OffsetAndLengths: [][2]int{{r.Range.Start.Column - 1, r.Range.Start.Column + len(r.Matched) - 1}},
			}
			lineMatches = append(lineMatches, lineMatch)
		}
		matches = append(matches,
			protocol.FileMatch{
				Path:        m.URI,
				LimitHit:    false,
				LineMatches: lineMatches,
			})
	}

	return matches, false, err
}
