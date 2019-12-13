package search

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/comby"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// The Sourcegraph frontend and interface only allow LineMatches (matches on a
// single line) and it isn't possible to specify a line and column range
// spanning multiple lines for highlighting. This function chops up potentially
// multiline matches into multiple LineMatches.
func highlightMultipleLines(r *comby.Match) (matches []protocol.LineMatch) {
	lineSpan := r.Range.End.Line - r.Range.Start.Line + 1
	if lineSpan == 1 {
		return []protocol.LineMatch{
			{
				LineNumber: r.Range.Start.Line - 1,
				OffsetAndLengths: [][2]int{
					{
						r.Range.Start.Column - 1,
						r.Range.End.Column - r.Range.Start.Column,
					},
				},
				Preview: r.Matched,
			},
		}
	}

	contentLines := strings.Split(r.Matched, "\n")
	for i, line := range contentLines {
		var columnStart, columnEnd int
		if i == 0 {
			// First line.
			columnStart = r.Range.Start.Column - 1
			columnEnd = len(line)
		} else if i == (lineSpan - 1) {
			// Last line.
			columnStart = 0
			columnEnd = r.Range.End.Column - 1 // don't include trailing newline
		} else {
			// In between line.
			columnStart = 0
			columnEnd = len(line)
		}

		matches = append(matches, protocol.LineMatch{
			LineNumber: r.Range.Start.Line + i - 1,
			OffsetAndLengths: [][2]int{
				{
					columnStart,
					columnEnd,
				},
			},
			Preview: line,
		})
	}
	return matches
}

func ToFileMatch(combyMatches []comby.FileMatch) (matches []protocol.FileMatch) {
	for _, m := range combyMatches {
		var lineMatches []protocol.LineMatch
		for _, r := range m.Matches {
			lineMatches = append(lineMatches, highlightMultipleLines(&r)...)
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

func structuralSearch(ctx context.Context, zipPath, pattern string, rule string, includePatterns []string, repo api.RepoName) (matches []protocol.FileMatch, limitHit bool, err error) {
	log15.Info("structural search", "repo", string(repo))

	// Cap the number of forked processes to limit the size of zip contents being mapped to memory. Resolving #7133 could help to lift this restriction.
	numWorkers := 4

	args := comby.Args{
		Input:         comby.ZipPath(zipPath),
		MatchTemplate: pattern,
		MatchOnly:     true,
		FilePatterns:  includePatterns,
		Rule:          rule,
		NumWorkers:    numWorkers,
	}

	combyMatches, err := comby.Matches(ctx, args)
	if err != nil {
		return nil, false, err
	}

	matches = ToFileMatch(combyMatches)
	if err != nil {
		return nil, false, err
	}
	return matches, false, err
}
