package keyword

import (
	"sort"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

type matchGroup struct {
	fileMatch *result.FileMatch
	group     result.ChunkMatches
	// fileScore is the pre-calculated score based on file metadata (e.g., file name).
	fileScore float64
	// distinctMatchesRatio is the ratio between the number of distinct pattern matches in the group and
	// the number of patterns in the query. It is the most basic measure of group relevancy. A relevant match group
	// should contain a sufficent ratio of matches as defined by `DISTINCT_MATCHES_RATIO_THRESHOLD`.
	distinctMatchesRatio float64
	// distinctMatchesPerLineRatio is the ratio between the sum of distinct pattern matches per line and
	// the total possible number of pattern matches. It is used to penalize long groups of matches with few distinct matches
	// (small distinctMatchesPerLineRatio).
	distinctMatchesPerLineRatio float64
	// keywordsPerLineRatio is the ratio of lines in the group that start with a keyword.
	keywordsPerLineRatio float64
}

const (
	DISTINCT_MATCHES_RATIO_THRESHOLD    = 0.75
	DISTINCT_MATCHES_PER_LINE_THRESHOLD = 0.66
)

func (g matchGroup) IsRelevant() bool {
	return g.distinctMatchesRatio >= DISTINCT_MATCHES_RATIO_THRESHOLD && g.distinctMatchesPerLineRatio >= DISTINCT_MATCHES_PER_LINE_THRESHOLD
}

func (g matchGroup) Score() float64 {
	return g.distinctMatchesRatio + g.distinctMatchesPerLineRatio + g.keywordsPerLineRatio + g.fileScore
}

func newChunkMatchGroup(fileMatch *result.FileMatch, fileScore float64, group result.ChunkMatches, numPatterns float64) matchGroup {
	distinctMatches := stringSet{}
	distinctMatchesPerLineCount := 0
	keywordCount := 0
	for _, chunkMatch := range group {
		distinctMatchesPerLine := stringSet{}

		// TODO(novoselrok): We should use symbols data if possible.
		if hasKeywordPrefix(chunkMatch.Content) {
			keywordCount += 1
		}

		for _, match := range chunkMatch.MatchedContent() {
			matchLowerCase := strings.ToLower(match)
			distinctMatches.Add(matchLowerCase)
			distinctMatchesPerLine.Add(matchLowerCase)
		}

		distinctMatchesPerLineCount += len(distinctMatchesPerLine)
	}

	lineCount := float64(len(group))
	distinctMatchesRatio := float64(len(distinctMatches)) / numPatterns
	distinctMatchesPerLineRatio := float64(distinctMatchesPerLineCount) / (lineCount * numPatterns)
	keywordsPerLine := float64(keywordCount) / lineCount

	return matchGroup{fileMatch, group, fileScore, distinctMatchesRatio, distinctMatchesPerLineRatio, keywordsPerLine}
}

func groupChunkMatches(fileMatch *result.FileMatch, fileScore float64, chunkMatches result.ChunkMatches, numPatterns float64) []matchGroup {
	// Sort chunks by line number
	sort.Slice(chunkMatches, func(i, j int) bool {
		return chunkMatches[i].ContentStart.Line < chunkMatches[j].ContentStart.Line
	})

	// Group chunk matches if they are within two lines of each other
	groups := []matchGroup{}
	startIndex := 0
	for i := 0; i < len(chunkMatches)-1; i++ {
		if chunkMatches[i+1].ContentStart.Line-chunkMatches[i].ContentStart.Line > 2 {
			groups = append(groups, newChunkMatchGroup(fileMatch, fileScore, chunkMatches[startIndex:i+1], numPatterns))
			startIndex = i + 1
		}
	}
	groups = append(groups, newChunkMatchGroup(fileMatch, fileScore, chunkMatches[startIndex:], numPatterns))
	return groups
}
