package embeddings

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type noOpUploadStore struct{}

func newNoOpUploadStore() uploadstore.Store {
	return &noOpUploadStore{}
}

func (s *noOpUploadStore) Init(ctx context.Context) error {
	return nil
}

func (s *noOpUploadStore) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	return nil, nil
}

func (s *noOpUploadStore) Upload(ctx context.Context, key string, r io.Reader) (int64, error) {
	p := make([]byte, 1024)
	totalRead := 0
	for {
		n, err := r.Read(p)
		if err == io.EOF {
			break
		}
		totalRead += n
		if err != nil {
			return int64(totalRead), err
		}
	}

	return int64(totalRead), nil
}

func (s *noOpUploadStore) Compose(ctx context.Context, destination string, sources ...string) (int64, error) {
	return 0, nil
}

func (s *noOpUploadStore) Delete(ctx context.Context, key string) error {
	return nil
}

func (s *noOpUploadStore) ExpireObjects(ctx context.Context, prefix string, maxAge time.Duration) error {
	return nil
}

type mockUploadStore struct {
	files map[string][]byte
}

func newMockUploadStore() uploadstore.Store {
	return &mockUploadStore{files: map[string][]byte{}}
}

func (s *mockUploadStore) Init(ctx context.Context) error {
	return nil
}

func (s *mockUploadStore) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	file, ok := s.files[key]
	if !ok {
		return nil, errors.Newf("file %s not found", key)
	}
	return io.NopCloser(bytes.NewReader(file)), nil
}

func (s *mockUploadStore) Upload(ctx context.Context, key string, r io.Reader) (int64, error) {
	file, err := io.ReadAll(r)
	if err != nil {
		return -1, errors.Newf("error reading file %s", key)
	}
	s.files[key] = file
	return int64(len(file)), nil
}

func (s *mockUploadStore) Compose(ctx context.Context, destination string, sources ...string) (int64, error) {
	return 0, nil
}

func (s *mockUploadStore) Delete(ctx context.Context, key string) error {
	return nil
}

func (s *mockUploadStore) ExpireObjects(ctx context.Context, prefix string, maxAge time.Duration) error {
	return nil
}

func TestRepoEmbeddingIndexStorage(t *testing.T) {
	index := &RepoEmbeddingIndex{
		RepoName: api.RepoName("repo"),
		Revision: api.CommitID("commit"),
		CodeIndex: EmbeddingIndex{
			Embeddings:      []int8{0, 1, 2},
			ColumnDimension: 3,
			RowMetadata:     []RepoEmbeddingRowMetadata{{FileName: "a.go", StartLine: 0, EndLine: 1}},
		},
		TextIndex: EmbeddingIndex{
			Embeddings:      []int8{10, 21, 32},
			ColumnDimension: 3,
			RowMetadata:     []RepoEmbeddingRowMetadata{{FileName: "b.py", StartLine: 0, EndLine: 1}},
		},
	}

	ctx := context.Background()
	uploadStore := newMockUploadStore()

	err := UploadRepoEmbeddingIndex(ctx, uploadStore, "index", index)
	require.NoError(t, err)

	downloadedIndex, err := DownloadRepoEmbeddingIndex(ctx, uploadStore, "index")
	require.NoError(t, err)

	require.Equal(t, index, downloadedIndex)
}

func TestRepoEmbeddingVersionMismatch(t *testing.T) {
	index := &OldRepoEmbeddingIndex{
		RepoName: api.RepoName("repo"),
		Revision: api.CommitID("commit"),
		CodeIndex: OldEmbeddingIndex{
			Embeddings:      []float32{0, 1, 2},
			ColumnDimension: 3,
			RowMetadata:     []RepoEmbeddingRowMetadata{{FileName: "a.go", StartLine: 0, EndLine: 1}},
		},
		TextIndex: OldEmbeddingIndex{
			Embeddings:      []float32{10, 21, 32},
			ColumnDimension: 3,
			RowMetadata:     []RepoEmbeddingRowMetadata{{FileName: "b.py", StartLine: 0, EndLine: 1}},
		},
	}

	ctx := context.Background()
	uploadStore := newMockUploadStore()

	// Upload the index using the "old" function.
	err := UploadIndex(ctx, uploadStore, "index", index)
	require.NoError(t, err)

	// Download the index using the new, custom function.
	downloadedIndex, err := DownloadRepoEmbeddingIndex(ctx, uploadStore, "index")
	require.NoError(t, err)

	require.Equal(t, index.ToNewIndex(), downloadedIndex)
}

func getMockEmbeddingIndex(nRows int, columnDimension int) EmbeddingIndex {
	embeddings := make([]int8, nRows*columnDimension)
	for idx := range embeddings {
		embeddings[idx] = int8(rand.Int())
	}

	rowMetadata := make([]RepoEmbeddingRowMetadata, nRows)
	for i := range rowMetadata {
		rowMetadata[i].StartLine = rand.Int()
		rowMetadata[i].EndLine = rand.Int()
		rowMetadata[i].FileName = fmt.Sprintf("path/to/file/%d_%d.go", rowMetadata[i].StartLine, rowMetadata[i].EndLine)
	}

	return EmbeddingIndex{
		Embeddings:      embeddings,
		ColumnDimension: columnDimension,
		RowMetadata:     rowMetadata,
	}
}

func BenchmarkRepoEmbeddingIndexUpload(b *testing.B) {
	// Roughly the size of the sourcegraph/sourcegraph index.
	index := &RepoEmbeddingIndex{
		RepoName:  api.RepoName("repo"),
		Revision:  api.CommitID("commit"),
		CodeIndex: getMockEmbeddingIndex(40_000, 1536),
		TextIndex: getMockEmbeddingIndex(10_000, 1536),
	}

	ctx := context.Background()
	uploadStore := newNoOpUploadStore()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := UploadIndex(ctx, uploadStore, "index", index)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCustomRepoEmbeddingIndexUpload(b *testing.B) {
	// Roughly the size of the sourcegraph/sourcegraph index.
	index := &RepoEmbeddingIndex{
		RepoName:  api.RepoName("repo"),
		Revision:  api.CommitID("commit"),
		CodeIndex: getMockEmbeddingIndex(40_000, 1536),
		TextIndex: getMockEmbeddingIndex(10_000, 1536),
	}

	ctx := context.Background()
	uploadStore := newNoOpUploadStore()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := UploadRepoEmbeddingIndex(ctx, uploadStore, "index", index)
		if err != nil {
			b.Fatal(err)
		}
	}
}
