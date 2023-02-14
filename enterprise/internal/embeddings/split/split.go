package split

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
)

var splittableLinePrefixes = []string{
	"//",
	"#",
	"/*",
	"func",
	"var",
	"const",
	"fn",
	"public",
	"private",
	"type",
}

func isSplittableLine(line string) bool {
	trimmedLine := strings.TrimSpace(line)
	if len(trimmedLine) == 0 {
		return true
	}
	for _, prefix := range splittableLinePrefixes {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}

type SplitOptions struct {
	ChunkTokensThreshold           int
	ChunkEarlySplitTokensThreshold int
}

func SplitIntoEmbeddableChunks(text string, fileName string, splitOptions SplitOptions) ([]string, []embeddings.RepoEmbeddingRowMetadata) {
	lines := strings.Split(text, "\n")
	chunks, metadata := []string{}, []embeddings.RepoEmbeddingRowMetadata{}
	startLine, tokensSum := 0, 0

	addChunk := func(endLine int) {
		chunks = append(chunks, strings.Join(lines[startLine:endLine], "\n"))
		metadata = append(metadata, embeddings.RepoEmbeddingRowMetadata{FileName: fileName, StartLine: startLine, EndLine: endLine})
		startLine, tokensSum = endLine, 0
	}

	for i := 0; i < len(lines); i++ {
		if tokensSum > splitOptions.ChunkTokensThreshold || (tokensSum > splitOptions.ChunkEarlySplitTokensThreshold && isSplittableLine(lines[i])) {
			addChunk(i)
		}
		tokensSum += embeddings.CountTokens(lines[i])
	}

	if tokensSum > 0 {
		addChunk(len(lines))
	}

	return chunks, metadata
}
