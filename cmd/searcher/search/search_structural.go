package search

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/comby"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
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

var languageExtensions = lazyregexp.New(`\.\w+`)

// inferExtension converts includePatterns like "\.c$|\.cats$|\.h$|\.idc$" to
// simply []string{".c",".cats",".h",".idc"} so that comby can infer the correct language
// matcher. The "\.c$" or "\.go$" format arises when we specify, for example,
// "lang:c" or "lang:go".
func inferExtensions(includePatterns []string) (new []string) {
	for _, s := range includePatterns {
		// Heuristically check that the pattern is not a filename by
		// checking whether it's terminated by $."
		if strings.HasSuffix(s, "$") {
			for _, extPat := range strings.Split(s, "|") {
				m := languageExtensions.FindString(extPat)
				if m != "" {
					new = append(new, m)
				}
			}
		} else {
			new = append(new, s)
		}
	}
	return new
}

func structuralSearch(ctx context.Context, zipPath, pattern string, includePatterns []string, repo api.RepoName) (matches []protocol.FileMatch, limitHit bool, err error) {
	log15.Info("structural search", "repo", string(repo))

	includePatterns = inferExtensions(includePatterns)

	args := comby.Args{
		Input:         comby.ZipPath(zipPath),
		MatchTemplate: pattern,
		MatchOnly:     true,
		FilePatterns:  includePatterns,
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
