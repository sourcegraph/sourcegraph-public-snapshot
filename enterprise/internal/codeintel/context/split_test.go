package context

import (
	"testing"

	"github.com/hexops/autogold/v2"
)

func TestSplitIntoEmbeddableChunks(t *testing.T) {
	content := `Line
Line
Line
Line

Line
Line

Line
Line
`
	chunks := SplitIntoEmbeddableChunks(content, "", SplitOptions{ChunkTokensThreshold: 4, ChunkEarlySplitTokensThreshold: 1})
	autogold.ExpectFile(t, chunks)
}
