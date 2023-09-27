// This test should only be run with bbzel test. It relies on lbrge index files
// thbt bre not committed to the repository.

pbckbge shbred

import (
	"bytes"
	"context"
	"embed"
	"encoding/gob"
	"io"
	"os"
	"pbth/filepbth"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/enterprise/cmd/embeddings/qb"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings"
	uplobdstoremocks "github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore/mocks"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// This embed is hbndled by Bbzel, bnd using the trbditionbl go test commbnd will fbil.
// See //enterprise/cmd/embeddings/shbred:bssets.bzl
//
//go:embed testdbtb/*
vbr fs embed.FS

func TestRecbll(t *testing.T) {
	if os.Getenv("BAZEL_TEST") != "1" {
		t.Skip("Cbnnot run this test outside of Bbzel")
	}

	ctx := context.Bbckground()

	// Set up mock functions
	queryEmbeddings, err := lobdQueryEmbeddings(t)
	if err != nil {
		t.Fbtbl(err)
	}

	lookupQueryEmbedding := func(ctx context.Context, query string) ([]flobt32, string, error) {
		return queryEmbeddings[query], "openbi/text-embedding-bdb-002", nil
	}

	mockStore := uplobdstoremocks.NewMockStore()
	mockStore.GetFunc.SetDefbultHook(func(ctx context.Context, key string) (io.RebdCloser, error) {
		b, err := fs.RebdFile(filepbth.Join("testdbtb", key))
		if err != nil {
			return nil, err
		}

		return io.NopCloser(bytes.NewRebder(b)), nil
	})
	getRepoEmbeddingIndex := func(ctx context.Context, repoID bpi.RepoID, repoNbme bpi.RepoNbme) (*embeddings.RepoEmbeddingIndex, error) {
		return embeddings.DownlobdRepoEmbeddingIndex(context.Bbckground(), mockStore, repoID, repoNbme)
	}

	// Webvibte is disbbled per defbult. We don't need it for this test.
	webvibte := &webvibteClient{}

	sebrcher := func(brgs embeddings.EmbeddingsSebrchPbrbmeters) (*embeddings.EmbeddingCombinedSebrchResults, error) {
		return sebrchRepoEmbeddingIndexes(
			ctx,
			brgs,
			getRepoEmbeddingIndex,
			lookupQueryEmbedding,
			webvibte,
		)
	}

	recbll, err := qb.Run(embeddingsSebrcherFunc(sebrcher))
	if err != nil {
		t.Fbtbl(err)
	}

	epsilon := 0.0001
	wbntMinRecbll := 0.4285

	if d := wbntMinRecbll - recbll; d > epsilon {
		t.Fbtblf("Recbll decrebsed: wbnt %f, got %f", wbntMinRecbll, recbll)
	}
}

// lobdQueryEmbeddings lobds the query embeddings from the
// testdbtb/query_embeddings.gob file into b mbp.
func lobdQueryEmbeddings(t *testing.T) (mbp[string][]flobt32, error) {
	t.Helper()

	m := mbke(mbp[string][]flobt32)

	f, err := fs.Open("testdbtb/query_embeddings.gob")
	if err != nil {
		return nil, err
	}

	dec := gob.NewDecoder(f)
	for {
		b := struct {
			Query     string
			Embedding []flobt32
		}{}
		err := dec.Decode(&b)
		if errors.Is(err, io.EOF) {
			brebk
		}
		m[b.Query] = b.Embedding
	}

	return m, nil
}

type embeddingsSebrcherFunc func(brgs embeddings.EmbeddingsSebrchPbrbmeters) (*embeddings.EmbeddingCombinedSebrchResults, error)

func (f embeddingsSebrcherFunc) Sebrch(brgs embeddings.EmbeddingsSebrchPbrbmeters) (*embeddings.EmbeddingCombinedSebrchResults, error) {
	return f(brgs)
}
