package shared

import (
	"context"
	"testing"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

func TestFilterAndHydrateContent_emptyFile(t *testing.T) {
	// Set up test data
	repoName := api.RepoName("example/repo")
	revision := api.CommitID("abc123")
	debug := false
	unfiltered := []embeddings.SimilaritySearchResult{
		{
			RepoEmbeddingRowMetadata: embeddings.RepoEmbeddingRowMetadata{
				FileName:  "file.txt",
				StartLine: 5,
				EndLine:   20,
			},
		},
	}

	// Define a mock readFile function that returns an empty string
	readFile := func(ctx context.Context, repoName api.RepoName, revision api.CommitID, fileName string) ([]byte, error) {
		return []byte{}, nil
	}

	filtered := filterAndHydrateContent(context.Background(), log.NoOp(), repoName, revision, readFile, debug, unfiltered)

	if len(filtered) != 1 {
		t.Errorf("Expected 1 filtered result, but got %d elements", len(filtered))
	}
}
