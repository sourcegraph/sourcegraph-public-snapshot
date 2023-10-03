// This test should only be run with bazel test. It relies on large index files
// that are not committed to the repository.

package shared

import (
	"bytes"
	"context"
	"embed"
	"encoding/gob"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/embeddings/qa"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/embeddings"
	uploadstoremocks "github.com/sourcegraph/sourcegraph/internal/uploadstore/mocks"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// This embed is handled by Bazel, and using the traditional go test command will fail.
// See //cmd/embeddings/shared:assets.bzl
//
//go:embed testdata/*
var fs embed.FS

func TestRecall(t *testing.T) {
	if os.Getenv("BAZEL_TEST") != "1" {
		t.Skip("Cannot run this test outside of Bazel")
	}

	ctx := context.Background()

	// Set up mock functions
	queryEmbeddings, err := loadQueryEmbeddings(t)
	if err != nil {
		t.Fatal(err)
	}

	lookupQueryEmbedding := func(ctx context.Context, query string) ([]float32, string, error) {
		return queryEmbeddings[query], "openai/text-embedding-ada-002", nil
	}

	mockStore := uploadstoremocks.NewMockStore()
	mockStore.GetFunc.SetDefaultHook(func(ctx context.Context, key string) (io.ReadCloser, error) {
		b, err := fs.ReadFile(filepath.Join("testdata", key))
		if err != nil {
			return nil, err
		}

		return io.NopCloser(bytes.NewReader(b)), nil
	})
	getRepoEmbeddingIndex := func(ctx context.Context, repoID api.RepoID, repoName api.RepoName) (*embeddings.RepoEmbeddingIndex, error) {
		return embeddings.DownloadRepoEmbeddingIndex(context.Background(), mockStore, repoID, repoName)
	}

	searcher := func(args embeddings.EmbeddingsSearchParameters) (*embeddings.EmbeddingCombinedSearchResults, error) {
		return searchRepoEmbeddingIndexes(
			ctx,
			args,
			getRepoEmbeddingIndex,
			lookupQueryEmbedding,
		)
	}

	recall, err := qa.Run(embeddingsSearcherFunc(searcher))
	if err != nil {
		t.Fatal(err)
	}

	epsilon := 0.0001
	wantMinRecall := 0.4285

	if d := wantMinRecall - recall; d > epsilon {
		t.Fatalf("Recall decreased: want %f, got %f", wantMinRecall, recall)
	}
}

// loadQueryEmbeddings loads the query embeddings from the
// testdata/query_embeddings.gob file into a map.
func loadQueryEmbeddings(t *testing.T) (map[string][]float32, error) {
	t.Helper()

	m := make(map[string][]float32)

	f, err := fs.Open("testdata/query_embeddings.gob")
	if err != nil {
		return nil, err
	}

	dec := gob.NewDecoder(f)
	for {
		a := struct {
			Query     string
			Embedding []float32
		}{}
		err := dec.Decode(&a)
		if errors.Is(err, io.EOF) {
			break
		}
		m[a.Query] = a.Embedding
	}

	return m, nil
}

type embeddingsSearcherFunc func(args embeddings.EmbeddingsSearchParameters) (*embeddings.EmbeddingCombinedSearchResults, error)

func (f embeddingsSearcherFunc) Search(args embeddings.EmbeddingsSearchParameters) (*embeddings.EmbeddingCombinedSearchResults, error) {
	return f(args)
}
