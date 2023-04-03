package embed

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/split"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

func mockFile(lines ...string) []byte {
	return []byte(strings.Join(lines, "\n"))
}

func TestEmbedRepo(t *testing.T) {
	ctx := context.Background()
	repoName := api.RepoName("repo/name")
	revision := api.CommitID("deadbeef")
	client := NewMockEmbeddingsClient()
	splitOptions := split.SplitOptions{ChunkTokensThreshold: 8}
	mockFiles := map[string][]byte{
		// 2 embedding chunks (based on split options above)
		"a.go": mockFile(
			strings.Repeat("a", 32),
			"",
			strings.Repeat("b", 32),
		),
		// 2 embedding chunks
		"b.md": mockFile(
			"# "+strings.Repeat("a", 32),
			"",
			"## "+strings.Repeat("b", 32),
		),
		// 3 embedding chunks
		"c.java": mockFile(
			strings.Repeat("a", 32),
			"",
			strings.Repeat("b", 32),
			"",
			strings.Repeat("c", 32),
		),
		// Should be excluded
		"autogen.py": mockFile(
			"# "+strings.Repeat("a", 32),
			"// Do not edit",
		),
		// Should be excluded
		"lines_too_long.c": mockFile(
			strings.Repeat("a", 2049),
			strings.Repeat("b", 2049),
			strings.Repeat("c", 2049),
		),
		// Should be excluded
		"empty.rb": mockFile(""),
		// Should be excluded (binary file),
		"binary.bin": {0xFF, 0xF, 0xF, 0xF, 0xFF, 0xF, 0xF, 0xA},
	}

	mockRanks := map[string]float64{
		"a.go":             0.1,
		"b.md":             0.2,
		"c.java":           0.3,
		"autogen.py":       0.4,
		"lines_too_long.c": 0.5,
		"empty.rb":         0.6,
		"binary.bin":       0.7,
	}

	t.Run("no files", func(t *testing.T) {
		index, err := EmbedRepo(ctx, repoName, revision, client, splitOptions, func(f func(fileName string, fileReader io.Reader) error) error {
			return nil
		}, func(f func(fileName string, fileReader io.Reader) error) error {
			return nil
		}, types.RepoPathRanks{
			MeanRank: 0,
			Paths:    mockRanks,
		})
		require.NoError(t, err)
		require.Len(t, index.CodeIndex.Embeddings, 0)
		require.Len(t, index.TextIndex.Embeddings, 0)
	})

	t.Run("code files only", func(t *testing.T) {
		index, err := EmbedRepo(ctx, repoName, revision, client, splitOptions,
			func(f func(fileName string, fileReader io.Reader) error) error {
				if err := f("a.go", bytes.NewReader(mockFiles["a.go"])); err != nil {
					return err
				}
				return nil
			},
			func(f func(fileName string, fileReader io.Reader) error) error {
				return nil
			},
			types.RepoPathRanks{
				MeanRank: 0,
				Paths:    mockRanks,
			})
		require.NoError(t, err)
		require.Len(t, index.TextIndex.Embeddings, 0)
		require.Len(t, index.CodeIndex.Embeddings, 6)
		require.Len(t, index.CodeIndex.RowMetadata, 2)
		require.Len(t, index.CodeIndex.Ranks, 2)
	})

	t.Run("text files only", func(t *testing.T) {
		index, err := EmbedRepo(ctx, repoName, revision, client, splitOptions,
			func(f func(fileName string, fileReader io.Reader) error) error {
				return nil
			},
			func(f func(fileName string, fileReader io.Reader) error) error {
				if err := f("b.md", bytes.NewReader(mockFiles["b.md"])); err != nil {
					return err
				}
				return nil
			},
			types.RepoPathRanks{
				MeanRank: 0,
				Paths:    mockRanks,
			})
		require.NoError(t, err)
		require.Len(t, index.CodeIndex.Embeddings, 0)
		require.Len(t, index.TextIndex.Embeddings, 6)
		require.Len(t, index.TextIndex.RowMetadata, 2)
		require.Len(t, index.TextIndex.Ranks, 2)
	})

	t.Run("mixed code and text files", func(t *testing.T) {
		files := []string{"a.go", "b.md", "c.java", "autogen.py", "empty.rb", "lines_too_long.c", "binary.bin"}
		index, err := EmbedRepo(ctx, repoName, revision, client, splitOptions,
			func(f func(fileName string, fileReader io.Reader) error) error {
				for _, file := range files {
					if !IsValidTextFile(file) {
						if err := f(file, bytes.NewReader(mockFiles[file])); err != nil {
							return err
						}
					}
				}
				return nil
			},
			func(f func(fileName string, fileReader io.Reader) error) error {
				for _, file := range files {
					if IsValidTextFile(file) {
						if err := f(file, bytes.NewReader(mockFiles[file])); err != nil {
							return err
						}
					}
				}
				return nil
			},
			types.RepoPathRanks{
				MeanRank: 0,
				Paths:    mockRanks,
			})
		require.NoError(t, err)
		require.Len(t, index.CodeIndex.Embeddings, 15)
		require.Len(t, index.CodeIndex.RowMetadata, 5)
		require.Len(t, index.CodeIndex.Ranks, 5)
		require.Len(t, index.TextIndex.Embeddings, 6)
		require.Len(t, index.TextIndex.RowMetadata, 2)
		require.Len(t, index.TextIndex.Ranks, 2)
	})
}

func NewMockEmbeddingsClient() EmbeddingsClient {
	return &mockEmbeddingsClient{}
}

type mockEmbeddingsClient struct{}

func (c *mockEmbeddingsClient) GetDimensions() (int, error) {
	return 3, nil
}

func (c *mockEmbeddingsClient) GetEmbeddingsWithRetries(texts []string, maxRetries int) ([]float32, error) {
	dimensions, err := c.GetDimensions()
	if err != nil {
		return nil, err
	}
	return make([]float32, len(texts)*dimensions), nil
}
