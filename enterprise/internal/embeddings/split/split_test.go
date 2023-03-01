package split

import (
	"testing"

	"github.com/hexops/autogold"
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
	autogold.Equal(t, chunks)
}
