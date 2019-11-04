package search

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/comby"
)

func ToFileMatch(combyMatches []comby.FileMatch) (matches []protocol.FileMatch) {
	for _, m := range combyMatches {
		var lineMatches []protocol.LineMatch
		for _, r := range m.Matches {
			lineMatch := protocol.LineMatch{
				LineNumber: r.Range.Start.Line - 1,
				// XXX sigh. assume one match per line.
				OffsetAndLengths: [][2]int{{r.Range.Start.Column - 1, len(r.Matched)}},
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
	return matches
}

func structuralSearch(ctx context.Context, zipPath string, pattern string, includePatterns []string, repo string, fileMatchLimit int) (matches []protocol.FileMatch, limitHit bool, err error) {
	fmt.Printf("structural search, repo %s, only do files: %d [%s]\n", repo, len(includePatterns), strings.Join(includePatterns, ","))

	args := comby.Args{
		Input:         comby.ZipPath(zipPath),
		MatchTemplate: pattern,
		MatchOnly:     true,
		FilePatterns:  includePatterns,
		// XXX unhardcode
		Matcher:    ".go",
		NumWorkers: numWorkers,
	}

	combyMatches, err := comby.Matches(args)
	if err != nil {
		return nil, false, err
	}

	matches = ToFileMatch(combyMatches)
	if err != nil {
		return nil, false, err
	}
	return matches, false, err
}
