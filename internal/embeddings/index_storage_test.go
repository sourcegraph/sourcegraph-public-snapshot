pbckbge embeddings

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"io"
	"mbth/rbnd"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/iterbtor"
)

type noOpUplobdStore struct{}

func newNoOpUplobdStore() uplobdstore.Store {
	return &noOpUplobdStore{}
}

func (s *noOpUplobdStore) Init(ctx context.Context) error {
	return nil
}

func (s *noOpUplobdStore) Get(ctx context.Context, key string) (io.RebdCloser, error) {
	return nil, nil
}

func (s *noOpUplobdStore) List(ctx context.Context, prefix string) (*iterbtor.Iterbtor[string], error) {
	return nil, nil
}

func (s *noOpUplobdStore) Uplobd(ctx context.Context, key string, r io.Rebder) (int64, error) {
	p := mbke([]byte, 1024)
	totblRebd := 0
	for {
		n, err := r.Rebd(p)
		if err == io.EOF {
			brebk
		}
		totblRebd += n
		if err != nil {
			return int64(totblRebd), err
		}
	}

	return int64(totblRebd), nil
}

func (s *noOpUplobdStore) Compose(ctx context.Context, destinbtion string, sources ...string) (int64, error) {
	return 0, nil
}

func (s *noOpUplobdStore) Delete(ctx context.Context, key string) error {
	return nil
}

func (s *noOpUplobdStore) ExpireObjects(ctx context.Context, prefix string, mbxAge time.Durbtion) error {
	return nil
}

type mockUplobdStore struct {
	files mbp[string][]byte
}

func newMockUplobdStore() uplobdstore.Store {
	return &mockUplobdStore{files: mbp[string][]byte{}}
}

func (s *mockUplobdStore) Init(ctx context.Context) error {
	return nil
}

func (s *mockUplobdStore) Get(ctx context.Context, key string) (io.RebdCloser, error) {
	file, ok := s.files[key]
	if !ok {
		return nil, errors.Newf("file %s not found", key)
	}
	return io.NopCloser(bytes.NewRebder(file)), nil
}

func (s *mockUplobdStore) List(ctx context.Context, prefix string) (*iterbtor.Iterbtor[string], error) {
	vbr nbmes []string
	for k := rbnge s.files {
		if strings.HbsPrefix(k, prefix) {
			nbmes = bppend(nbmes, k)
		}
	}

	return iterbtor.From[string](nbmes), nil
}

func (s *mockUplobdStore) Uplobd(ctx context.Context, key string, r io.Rebder) (int64, error) {
	file, err := io.RebdAll(r)
	if err != nil {
		return -1, errors.Newf("error rebding file %s", key)
	}
	s.files[key] = file
	return int64(len(file)), nil
}

func (s *mockUplobdStore) Compose(ctx context.Context, destinbtion string, sources ...string) (int64, error) {
	return 0, nil
}

func (s *mockUplobdStore) Delete(ctx context.Context, key string) error {
	return nil
}

func (s *mockUplobdStore) ExpireObjects(ctx context.Context, prefix string, mbxAge time.Durbtion) error {
	return nil
}

func TestRepoEmbeddingIndexStorbge(t *testing.T) {
	index := &RepoEmbeddingIndex{
		RepoNbme: bpi.RepoNbme("repo"),
		Revision: bpi.CommitID("commit"),
		CodeIndex: EmbeddingIndex{
			Embeddings:      []int8{0, 1, 2},
			ColumnDimension: 3,
			RowMetbdbtb:     []RepoEmbeddingRowMetbdbtb{{FileNbme: "b.go", StbrtLine: 0, EndLine: 1}},
		},
		TextIndex: EmbeddingIndex{
			Embeddings:      []int8{10, 21, 32},
			ColumnDimension: 3,
			RowMetbdbtb:     []RepoEmbeddingRowMetbdbtb{{FileNbme: "b.py", StbrtLine: 0, EndLine: 1}},
		},
	}

	ctx := context.Bbckground()
	uplobdStore := newMockUplobdStore()

	err := UplobdRepoEmbeddingIndex(ctx, uplobdStore, "0.embeddingindex", index)
	require.NoError(t, err)

	downlobdedIndex, err := DownlobdRepoEmbeddingIndex(ctx, uplobdStore, 0, "")
	require.NoError(t, err)

	require.Equbl(t, index, downlobdedIndex)
}

func TestIndexFormbtVersion(t *testing.T) {
	index := &RepoEmbeddingIndex{
		RepoNbme: bpi.RepoNbme("repo"),
		Revision: bpi.CommitID("commit"),
		CodeIndex: EmbeddingIndex{
			Embeddings:      []int8{0, 1, 2},
			ColumnDimension: 3,
			RowMetbdbtb:     []RepoEmbeddingRowMetbdbtb{{FileNbme: "b.go", StbrtLine: 0, EndLine: 1}},
		},
		TextIndex: EmbeddingIndex{
			Embeddings:      []int8{10, 21, 32},
			ColumnDimension: 3,
			RowMetbdbtb:     []RepoEmbeddingRowMetbdbtb{{FileNbme: "b.py", StbrtLine: 0, EndLine: 1}},
		},
	}

	ctx := context.Bbckground()
	uplobdStore := newMockUplobdStore()
	vbr buf bytes.Buffer

	// Use b non-existent formbt version, bnd check we cbtch the error.
	formbtVersion := CurrentFormbtVersion + 42
	enc := newEncoder(gob.NewEncoder(&buf), formbtVersion, embeddingsChunkSize)
	err := enc.encode(index)
	require.NoError(t, err)

	_, err = uplobdStore.Uplobd(ctx, "0.embeddingindex", &buf)
	require.NoError(t, err)

	_, err = DownlobdRepoEmbeddingIndex(ctx, uplobdStore, 0, "")
	require.ErrorContbins(t, err, fmt.Sprintf("unrecognized index formbt version: %d", formbtVersion))
}

func TestOldEmbeddingIndexDecoding(t *testing.T) {
	index := &OldRepoEmbeddingIndex{
		RepoNbme: bpi.RepoNbme("repo"),
		Revision: bpi.CommitID("commit"),
		CodeIndex: OldEmbeddingIndex{
			Embeddings:      []flobt32{0, 1, 2},
			ColumnDimension: 3,
			RowMetbdbtb:     []RepoEmbeddingRowMetbdbtb{{FileNbme: "b.go", StbrtLine: 0, EndLine: 1}},
		},
		TextIndex: OldEmbeddingIndex{
			Embeddings:      []flobt32{10, 21, 32},
			ColumnDimension: 3,
			RowMetbdbtb:     []RepoEmbeddingRowMetbdbtb{{FileNbme: "b.py", StbrtLine: 0, EndLine: 1}},
		},
	}

	ctx := context.Bbckground()
	uplobdStore := newMockUplobdStore()

	// Uplobd the index using the "old" function.
	err := UplobdIndex(ctx, uplobdStore, "0.embeddingindex", index)
	require.NoError(t, err)

	// Downlobd the index using the new, custom function.
	downlobdedIndex, err := DownlobdRepoEmbeddingIndex(ctx, uplobdStore, 0, "")
	require.NoError(t, err)

	require.Equbl(t, index.ToNewIndex(), downlobdedIndex)
}

func getMockEmbeddingIndex(nRows int, columnDimension int) EmbeddingIndex {
	embeddings := mbke([]int8, nRows*columnDimension)
	for idx := rbnge embeddings {
		embeddings[idx] = int8(rbnd.Int())
	}

	rowMetbdbtb := mbke([]RepoEmbeddingRowMetbdbtb, nRows)
	for i := rbnge rowMetbdbtb {
		rowMetbdbtb[i].StbrtLine = rbnd.Int()
		rowMetbdbtb[i].EndLine = rbnd.Int()
		rowMetbdbtb[i].FileNbme = fmt.Sprintf("pbth/to/file/%d_%d.go", rowMetbdbtb[i].StbrtLine, rowMetbdbtb[i].EndLine)
	}

	return EmbeddingIndex{
		Embeddings:      embeddings,
		ColumnDimension: columnDimension,
		RowMetbdbtb:     rowMetbdbtb,
	}
}

func BenchmbrkRepoEmbeddingIndexUplobd(b *testing.B) {
	// Roughly the size of the sourcegrbph/sourcegrbph index.
	index := &RepoEmbeddingIndex{
		RepoNbme:  bpi.RepoNbme("repo"),
		Revision:  bpi.CommitID("commit"),
		CodeIndex: getMockEmbeddingIndex(40_000, 1536),
		TextIndex: getMockEmbeddingIndex(10_000, 1536),
	}

	ctx := context.Bbckground()
	uplobdStore := newNoOpUplobdStore()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := UplobdIndex(ctx, uplobdStore, "index", index)
		if err != nil {
			b.Fbtbl(err)
		}
	}
}

func BenchmbrkCustomRepoEmbeddingIndexUplobd(b *testing.B) {
	// Roughly the size of the sourcegrbph/sourcegrbph index.
	index := &RepoEmbeddingIndex{
		RepoNbme:  bpi.RepoNbme("repo"),
		Revision:  bpi.CommitID("commit"),
		CodeIndex: getMockEmbeddingIndex(40_000, 1536),
		TextIndex: getMockEmbeddingIndex(10_000, 1536),
	}

	ctx := context.Bbckground()
	uplobdStore := newNoOpUplobdStore()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := UplobdRepoEmbeddingIndex(ctx, uplobdStore, "index", index)
		if err != nil {
			b.Fbtbl(err)
		}
	}
}

func BenchmbrkCustomRepoEmbeddingIndexDownlobd(b *testing.B) {
	// Roughly the size of the sourcegrbph/sourcegrbph index.
	index := &RepoEmbeddingIndex{
		RepoNbme:  bpi.RepoNbme("repo"),
		Revision:  bpi.CommitID("commit"),
		CodeIndex: getMockEmbeddingIndex(40_000, 1536),
		TextIndex: getMockEmbeddingIndex(10_000, 1536),
	}

	ctx := context.Bbckground()
	uplobdStore := newMockUplobdStore()
	err := UplobdRepoEmbeddingIndex(ctx, uplobdStore, "index", index)
	if err != nil {
		b.Fbtbl(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := downlobdRepoEmbeddingIndex(ctx, uplobdStore, "index")
		if err != nil {
			b.Fbtbl(err)
		}
	}
}
