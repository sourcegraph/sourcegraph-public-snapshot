package smart

import (
	"sort"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

type matchGroup struct {
	fileMatch                   *result.FileMatch
	group                       result.ChunkMatches
	fileScore                   float64
	distinctMatchesRatio        float64
	distinctMatchesPerLineRatio float64
	keywordsPerLine             float64
}

// TODO: Thresholds are best effort for now
func (g matchGroup) IsValid() bool {
	return g.distinctMatchesRatio >= 0.75 && g.distinctMatchesPerLineRatio >= 0.66
}

// TODO: This should be a property
func (g matchGroup) Score() float64 {
	return g.distinctMatchesRatio + g.distinctMatchesPerLineRatio + g.keywordsPerLine + g.fileScore
}

func newChunkMatchGroup(fileMatch *result.FileMatch, fileScore float64, group result.ChunkMatches, numPatterns float64) matchGroup {
	distinctMatches := stringSet{}
	distinctMatchesPerLineCount := 0
	keywordCount := 0
	for _, chunkMatch := range group {
		distinctMatchesPerLine := stringSet{}

		matches := chunkMatch.MatchedContent()
		linePreview := strings.TrimSpace(chunkMatch.AsLineMatches()[0].Preview)

		// TODO: Not the best. We really should be using symbols data.
		if hasKeywordPrefix(linePreview) {
			keywordCount += 1
		}

		for _, match := range matches {
			matchLowerCase := strings.ToLower(match)
			distinctMatches.Add(matchLowerCase)
			distinctMatchesPerLine.Add(matchLowerCase)
		}
		distinctMatchesPerLineCount += len(distinctMatchesPerLine)
	}

	nLines := float64(len(group))
	distinctMatchesRatio := float64(len(distinctMatches)) / numPatterns
	distinctMatchesPerLineRatio := float64(distinctMatchesPerLineCount) / (nLines * numPatterns)
	keywordsPerLine := float64(keywordCount) / nLines

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

func getFileScore(filePath string, patterns []string) float64 {
	filePathLowerCase := strings.ToLower(filePath)
	count := 0
	for _, pattern := range patterns {
		if strings.Contains(filePathLowerCase, pattern) {
			count += 1
		}
	}
	return float64(count) / float64(len(patterns))
}
