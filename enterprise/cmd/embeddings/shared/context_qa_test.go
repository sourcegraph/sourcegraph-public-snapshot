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

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/embeddings/qa"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/api"
	uploadstoremocks "github.com/sourcegraph/sourcegraph/internal/uploadstore/mocks"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// This embed is handled by Bazel, and using the traditional go test command will fail.
// See //enterprise/cmd/embeddings/shared:assets.bzl
//
//go:embed testdata/*
var fs embed.FS

func TestRecall(t *testing.T) {
	if os.Getenv("BAZEL_TEST") != "1" {
		t.Skip("Cannot run this test outside of Bazel")
	}

	ctx := context.Background()
	logger := log.NoOp()

	// Set up mock functions
	queryEmbeddings, err := loadQueryEmbeddings(t)
	if err != nil {
		t.Fatal(err)
	}
	getQueryEmbedding := func(ctx context.Context, query string) ([]float32, error) {
		return queryEmbeddings[query], nil
	}

	mockStore := uploadstoremocks.NewMockStore()
	mockStore.GetFunc.SetDefaultHook(func(ctx context.Context, key string) (io.ReadCloser, error) {
		b, err := fs.ReadFile(filepath.Join("testdata", key))
		if err != nil {
			return nil, err
		}

		return io.NopCloser(bytes.NewReader(b)), nil
	})
	getRepoEmbeddingIndex := func(ctx context.Context, repo api.RepoName) (*embeddings.RepoEmbeddingIndex, error) {
		key := embeddings.GetRepoEmbeddingIndexName(repo)
		return embeddings.DownloadRepoEmbeddingIndex(context.Background(), mockStore, string(key))
	}

	// We only care about the file names in this test.
	mockReadFile := func(ctx context.Context, repoName api.RepoName, revision api.CommitID, fileName string) ([]byte, error) {
		return []byte{}, nil
	}

	// Weaviate is disabled per default. We don't need it for this test.
	weaviate := &weaviateClient{}

	searcher := func(args embeddings.EmbeddingsSearchParameters) (*embeddings.EmbeddingSearchResults, error) {
		return searchRepoEmbeddingIndex(
			ctx,
			logger,
			args,
			mockReadFile,
			getRepoEmbeddingIndex,
			getQueryEmbedding,
			weaviate,
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

type embeddingsSearcherFunc func(args embeddings.EmbeddingsSearchParameters) (*embeddings.EmbeddingSearchResults, error)

func (f embeddingsSearcherFunc) Search(args embeddings.EmbeddingsSearchParameters) (*embeddings.EmbeddingSearchResults, error) {
	return f(args)
}
