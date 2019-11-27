package search

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/comby"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func ToFileMatch(combyMatches []comby.FileMatch) (matches []protocol.FileMatch) {
	for _, m := range combyMatches {
		var lineMatches []protocol.LineMatch
		for _, r := range m.Matches {
			lineMatch := protocol.LineMatch{
				LineNumber: r.Range.Start.Line - 1,
				OffsetAndLengths: [][2]int{
					{
						r.Range.Start.Column - 1,
						r.Range.End.Column - r.Range.Start.Column,
					},
				},
				Preview: r.Matched,
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

func structuralSearch(ctx context.Context, zipPath, pattern string, includePatterns []string, repo api.RepoName) (matches []protocol.FileMatch, limitHit bool, err error) {
	log15.Info("structural search", "repo", string(repo))

	args := comby.Args{
		Input:         comby.ZipPath(zipPath),
		MatchTemplate: pattern,
		MatchOnly:     true,
		FilePatterns:  includePatterns,
		NumWorkers:    numWorkers,
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
