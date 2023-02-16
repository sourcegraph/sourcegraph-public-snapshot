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

type EmbeddableChunk struct {
	FileName  string
	StartLine int
	EndLine   int
	Content   string
}

func SplitIntoEmbeddableChunks(text string, fileName string, splitOptions SplitOptions) []EmbeddableChunk {
	chunks := []EmbeddableChunk{}
	startLine, tokensSum := 0, 0
	lines := strings.Split(text, "\n")

	addChunk := func(endLine int) {
		content := strings.Join(lines[startLine:endLine], "\n")
		if len(content) > 0 {
			chunks = append(chunks, EmbeddableChunk{FileName: fileName, StartLine: startLine, EndLine: endLine, Content: content})
		}
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

	return chunks
}
