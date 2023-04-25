package embed

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
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

	getDocumentRanks := func(ctx context.Context, repoName string) (types.RepoPathRanks, error) {
		return types.RepoPathRanks{
			MeanRank: 0,
			Paths:    mockRanks,
		}, nil
	}

	reader := funcReader(func(_ context.Context, fileName string) ([]byte, error) {
		content, ok := mockFiles[fileName]
		if !ok {
			return nil, errors.Newf("file %s not found", fileName)
		}
		return content, nil
	})

	newReadLister := func(fileNames ...string) FileReadLister {
		fileEntries := make([]FileEntry, len(fileNames))
		for i, fileName := range fileNames {
			fileEntries[i] = FileEntry{Name: fileName, Size: 350}
		}
		return listReader{
			FileReader: reader,
			FileLister: staticLister(fileEntries),
		}
	}

	excludedGlobPatterns := GetDefaultExcludedFilePathPatterns()

	t.Run("no files", func(t *testing.T) {
		index, stats, err := EmbedRepo(ctx, repoName, revision, excludedGlobPatterns, client, splitOptions, newReadLister(), getDocumentRanks)
		require.NoError(t, err)
		require.Len(t, index.CodeIndex.Embeddings, 0)
		require.Len(t, index.TextIndex.Embeddings, 0)

		expectedStats := &embeddings.EmbedRepoStats{
			HasRanks: true,
			CodeIndexStats: embeddings.EmbedFilesStats{
				SkippedByteCounts: map[string]int{},
				SkippedCounts:     map[string]int{},
			},
			TextIndexStats: embeddings.EmbedFilesStats{
				SkippedByteCounts: map[string]int{},
				SkippedCounts:     map[string]int{},
			},
		}
		// ignore durations
		stats.Duration = 0
		stats.CodeIndexStats.Duration = 0
		stats.TextIndexStats.Duration = 0
		require.Equal(t, expectedStats, stats)
	})

	t.Run("code files only", func(t *testing.T) {
		index, stats, err := EmbedRepo(ctx, repoName, revision, excludedGlobPatterns, client, splitOptions, newReadLister("a.go"), getDocumentRanks)
		require.NoError(t, err)
		require.Len(t, index.TextIndex.Embeddings, 0)
		require.Len(t, index.CodeIndex.Embeddings, 6)
		require.Len(t, index.CodeIndex.RowMetadata, 2)
		require.Len(t, index.CodeIndex.Ranks, 2)

		expectedStats := &embeddings.EmbedRepoStats{
			HasRanks: true,
			CodeIndexStats: embeddings.EmbedFilesStats{
				EmbeddedFileCount:  1,
				EmbeddedChunkCount: 2,
				EmbeddedBytes:      65,
				SkippedByteCounts:  map[string]int{},
				SkippedCounts:      map[string]int{},
			},
			TextIndexStats: embeddings.EmbedFilesStats{
				SkippedByteCounts: map[string]int{},
				SkippedCounts:     map[string]int{},
			},
		}
		// ignore durations
		stats.Duration = 0
		stats.CodeIndexStats.Duration = 0
		stats.TextIndexStats.Duration = 0
		require.Equal(t, expectedStats, stats)
	})

	t.Run("text files only", func(t *testing.T) {
		index, stats, err := EmbedRepo(ctx, repoName, revision, excludedGlobPatterns, client, splitOptions, newReadLister("b.md"), getDocumentRanks)
		require.NoError(t, err)
		require.Len(t, index.CodeIndex.Embeddings, 0)
		require.Len(t, index.TextIndex.Embeddings, 6)
		require.Len(t, index.TextIndex.RowMetadata, 2)
		require.Len(t, index.TextIndex.Ranks, 2)

		expectedStats := &embeddings.EmbedRepoStats{
			HasRanks: true,
			CodeIndexStats: embeddings.EmbedFilesStats{
				SkippedByteCounts: map[string]int{},
				SkippedCounts:     map[string]int{},
			},
			TextIndexStats: embeddings.EmbedFilesStats{
				EmbeddedFileCount:  1,
				EmbeddedChunkCount: 2,
				EmbeddedBytes:      70,
				SkippedByteCounts:  map[string]int{},
				SkippedCounts:      map[string]int{},
			},
		}
		// ignore durations
		stats.Duration = 0
		stats.CodeIndexStats.Duration = 0
		stats.TextIndexStats.Duration = 0
		require.Equal(t, expectedStats, stats)
	})

	t.Run("mixed code and text files", func(t *testing.T) {
		rl := newReadLister("a.go", "b.md", "c.java", "autogen.py", "empty.rb", "lines_too_long.c", "binary.bin")
		index, stats, err := EmbedRepo(ctx, repoName, revision, excludedGlobPatterns, client, splitOptions, rl, getDocumentRanks)
		require.NoError(t, err)
		require.Len(t, index.CodeIndex.Embeddings, 15)
		require.Len(t, index.CodeIndex.RowMetadata, 5)
		require.Len(t, index.CodeIndex.Ranks, 5)
		require.Len(t, index.TextIndex.Embeddings, 6)
		require.Len(t, index.TextIndex.RowMetadata, 2)
		require.Len(t, index.TextIndex.Ranks, 2)

		expectedStats := &embeddings.EmbedRepoStats{
			HasRanks: true,
			CodeIndexStats: embeddings.EmbedFilesStats{
				EmbeddedFileCount:  2,
				EmbeddedChunkCount: 5,
				EmbeddedBytes:      163,
				SkippedByteCounts: map[string]int{
					"autogenerated": 49,
					"binary":        8,
					"longLine":      6149,
					"small":         0,
				},
				SkippedCounts: map[string]int{
					"autogenerated": 1,
					"binary":        1,
					"longLine":      1,
					"small":         1,
				},
			},
			TextIndexStats: embeddings.EmbedFilesStats{
				EmbeddedFileCount:  1,
				EmbeddedChunkCount: 2,
				EmbeddedBytes:      70,
				SkippedByteCounts:  map[string]int{},
				SkippedCounts:      map[string]int{},
			},
		}
		// ignore durations
		stats.Duration = 0
		stats.CodeIndexStats.Duration = 0
		stats.TextIndexStats.Duration = 0
		require.Equal(t, expectedStats, stats)
	})
}

func NewMockEmbeddingsClient() EmbeddingsClient {
	return &mockEmbeddingsClient{}
}

type mockEmbeddingsClient struct{}

func (c *mockEmbeddingsClient) GetDimensions() (int, error) {
	return 3, nil
}

func (c *mockEmbeddingsClient) GetEmbeddingsWithRetries(_ context.Context, texts []string, maxRetries int) ([]float32, error) {
	dimensions, err := c.GetDimensions()
	if err != nil {
		return nil, err
	}
	return make([]float32, len(texts)*dimensions), nil
}

type funcReader func(ctx context.Context, fileName string) ([]byte, error)

func (f funcReader) Read(ctx context.Context, fileName string) ([]byte, error) {
	return f(ctx, fileName)
}

type staticLister []FileEntry

func (l staticLister) List(_ context.Context) ([]FileEntry, error) {
	return l, nil
}

type listReader struct {
	FileReader
	FileLister
}
