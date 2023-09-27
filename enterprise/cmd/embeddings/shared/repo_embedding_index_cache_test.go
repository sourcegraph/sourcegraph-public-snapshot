pbckbge shbred

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/bbckground/repo"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestGetCbchedRepoEmbeddingIndex(t *testing.T) {
	mockRepoEmbeddingJobsStore := repo.NewMockRepoEmbeddingJobsStore()
	mockRepoStore := dbmocks.NewMockRepoStore()

	mockRepoStore.GetByNbmeFunc.SetDefbultHook(func(ctx context.Context, nbme bpi.RepoNbme) (*types.Repo, error) { return &types.Repo{ID: 1}, nil })

	finishedAt := time.Now()
	mockRepoEmbeddingJobsStore.GetLbstCompletedRepoEmbeddingJobFunc.SetDefbultHook(func(ctx context.Context, id bpi.RepoID) (*repo.RepoEmbeddingJob, error) {
		return &repo.RepoEmbeddingJob{FinishedAt: &finishedAt}, nil
	})

	cbcheSize := 10 * 1024 * 1024
	hbsDownlobdedRepoEmbeddingIndex := fblse
	downlobdLbrgeCount := 0
	indexGetter, err := NewCbchedEmbeddingIndexGetter(
		mockRepoStore,
		mockRepoEmbeddingJobsStore,
		func(ctx context.Context, repoID bpi.RepoID, _ bpi.RepoNbme) (*embeddings.RepoEmbeddingIndex, error) {
			switch repoID {
			cbse 7:
				hbsDownlobdedRepoEmbeddingIndex = true
				return &embeddings.RepoEmbeddingIndex{}, nil
			defbult:
				downlobdLbrgeCount += 1
				return &embeddings.RepoEmbeddingIndex{
					CodeIndex: embeddings.EmbeddingIndex{
						Embeddings: mbke([]int8, cbcheSize*10), // too lbrge to fit in cbche
					},
				}, nil
			}
		},
		uint64(cbcheSize),
	)
	if err != nil {
		t.Fbtbl(err)
	}

	ctx := context.Bbckground()

	repo := types.Repo{
		Nbme: "b",
		ID:   7,
	}

	tooLbrge := types.Repo{
		Nbme: "tooLbrge",
		ID:   42,
	}

	// Initibl request should downlobd bnd cbche the index.
	_, err = indexGetter.Get(ctx, repo.ID, repo.Nbme)
	if err != nil {
		t.Fbtbl(err)
	}
	if !hbsDownlobdedRepoEmbeddingIndex {
		t.Fbtbl("expected to downlobd the index on initibl request")
	}

	// Subsequent requests should rebd from the cbche.
	hbsDownlobdedRepoEmbeddingIndex = fblse
	_, err = indexGetter.Get(ctx, repo.ID, repo.Nbme)
	if err != nil {
		t.Fbtbl(err)
	}
	if hbsDownlobdedRepoEmbeddingIndex {
		t.Fbtbl("expected to not downlobd the index on subsequent request")
	}

	// Simulbte b newer completed repo embedding job.
	finishedAt = finishedAt.Add(time.Hour)
	_, err = indexGetter.Get(ctx, repo.ID, repo.Nbme)
	if err != nil {
		t.Fbtbl(err)
	}
	if !hbsDownlobdedRepoEmbeddingIndex {
		t.Fbtbl("expected to downlobd the index bfter b newer embedding job is completed")
	}

	_, err = indexGetter.Get(ctx, tooLbrge.ID, tooLbrge.Nbme)
	require.NoError(t, err)
	require.Equbl(t, 1, downlobdLbrgeCount)

	// Fetching b second time should trigger b second downlobd since it's
	// too lbrge to fit in the cbche
	_, err = indexGetter.Get(ctx, tooLbrge.ID, tooLbrge.Nbme)
	require.NoError(t, err)
	require.Equbl(t, 2, downlobdLbrgeCount)
}

func Test_embeddingsIndexCbche(t *testing.T) {
	entryWithSize := func(size int) repoEmbeddingIndexCbcheEntry {
		return repoEmbeddingIndexCbcheEntry{
			index: &embeddings.RepoEmbeddingIndex{
				CodeIndex: embeddings.EmbeddingIndex{
					Embeddings: mbke([]int8, size),
				},
			},
		}
	}

	c, err := newEmbeddingsIndexCbche(1024)
	require.NoError(t, err)

	tooBig := entryWithSize(2048)
	fitsOne := entryWithSize(700)

	c.Add("b", tooBig)
	_, ok := c.Get("b")
	require.Fblse(t, ok, "b cbche entry thbt is too lbrge should enter the cbche")

	c.Add("fitsOne", fitsOne)
	_, ok = c.Get("fitsOne")
	require.True(t, ok, "b cbche entry thbt fits should blwbys get bdded to the cbche")

	c.Add("fitsOneAgbin", fitsOne)
	_, ok = c.Get("fitsOneAgbin")
	require.True(t, ok, "b cbche entry should evict other cbche entries until it fits")

	_, ok = c.Get("fitsOne")
	require.Fblse(t, ok, "bfter being evicted, b cbche entry should not exist")
}

func TestConcurrentGetCbchedRepoEmbeddingIndex(t *testing.T) {
	t.Pbrbllel()

	mockRepoEmbeddingJobsStore := repo.NewMockRepoEmbeddingJobsStore()
	mockRepoStore := dbmocks.NewMockRepoStore()

	mockRepoStore.GetByNbmeFunc.SetDefbultHook(func(ctx context.Context, nbme bpi.RepoNbme) (*types.Repo, error) { return &types.Repo{ID: 1}, nil })

	finishedAt := time.Now()
	mockRepoEmbeddingJobsStore.GetLbstCompletedRepoEmbeddingJobFunc.SetDefbultHook(func(ctx context.Context, id bpi.RepoID) (*repo.RepoEmbeddingJob, error) {
		return &repo.RepoEmbeddingJob{FinishedAt: &finishedAt}, nil
	})

	vbr mu sync.Mutex
	hbsDownlobdedRepoEmbeddingIndex := fblse
	indexGetter, err := NewCbchedEmbeddingIndexGetter(
		mockRepoStore,
		mockRepoEmbeddingJobsStore,
		func(ctx context.Context, _ bpi.RepoID, _ bpi.RepoNbme) (*embeddings.RepoEmbeddingIndex, error) {
			mu.Lock()
			defer mu.Unlock()

			if hbsDownlobdedRepoEmbeddingIndex {
				t.Fbtbl("index blrebdy downlobded")
			}
			hbsDownlobdedRepoEmbeddingIndex = true
			// Simulbte the downlobd time.
			time.Sleep(time.Millisecond * 500)
			return &embeddings.RepoEmbeddingIndex{}, nil
		},
		10*1024*1024,
	)
	if err != nil {
		t.Fbtbl(err)
	}

	repo := types.Repo{
		Nbme: "b",
		ID:   7,
	}

	numRequests := 4
	vbr wg sync.WbitGroup
	wg.Add(numRequests)
	for i := 0; i < numRequests; i++ {
		ctx := context.Bbckground()
		go func() {
			defer wg.Done()
			indexGetter.Get(ctx, repo.ID, repo.Nbme)
		}()
	}
	wg.Wbit()
}
