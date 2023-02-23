package embed

import (
	"context"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/errors"

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
	readFile := func(fileName string) ([]byte, error) {
		content, ok := mockFiles[fileName]
		if !ok {
			return nil, errors.Newf("file %s not found", fileName)
		}
		return content, nil
	}

	t.Run("no files", func(t *testing.T) {
		index, err := EmbedRepo(ctx, repoName, revision, []string{}, client, splitOptions, readFile)
		if err != nil {
			t.Fatal(err)
		}
		if index.CodeIndex != nil {
			t.Fatal("expected code index to be nil")
		}
		if index.TextIndex != nil {
			t.Fatal("expected text index to be nil")
		}
	})

	t.Run("code files only", func(t *testing.T) {
		index, err := EmbedRepo(ctx, repoName, revision, []string{"a.go"}, client, splitOptions, readFile)
		if err != nil {
			t.Fatal(err)
		}
		if index.TextIndex != nil {
			t.Fatal("expected text index to be nil")
		}
		if index.CodeIndex == nil {
			t.Fatal("expected code index to be non-nil")
		}
		if len(index.CodeIndex.RowMetadata) != 2 {
			t.Fatal("expected 2 embedding rows")
		}
	})

	t.Run("text files only", func(t *testing.T) {
		index, err := EmbedRepo(ctx, repoName, revision, []string{"b.md"}, client, splitOptions, readFile)
		if err != nil {
			t.Fatal(err)
		}
		if index.CodeIndex != nil {
			t.Fatal("expected code index to be nil")
		}
		if index.TextIndex == nil {
			t.Fatal("expected text index to be non-nil")
		}
		if len(index.TextIndex.RowMetadata) != 2 {
			t.Fatal("expected 2 embedding rows")
		}
	})

	t.Run("mixed code and text files", func(t *testing.T) {
		files := []string{"a.go", "b.md", "c.java", "autogen.py", "empty.rb", "lines_too_long.c", "binary.bin"}
		index, err := EmbedRepo(ctx, repoName, revision, files, client, splitOptions, readFile)
		if err != nil {
			t.Fatal(err)
		}
		if index.CodeIndex == nil {
			t.Fatal("expected code index to be non-nil")
		}
		if index.TextIndex == nil {
			t.Fatal("expected text index to be non-nil")
		}
		if len(index.CodeIndex.RowMetadata) != 5 {
			t.Fatalf("expected 5 embedding rows in code index, got %d", len(index.CodeIndex.RowMetadata))
		}
		if len(index.TextIndex.RowMetadata) != 2 {
			t.Fatalf("expected 2 embedding rows in text index, got %d", len(index.TextIndex.RowMetadata))
		}
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
