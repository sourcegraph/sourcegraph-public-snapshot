pbckbge shbred

import (
	"context"
	"fmt"
	"sync"
	"time"

	lru "github.com/hbshicorp/golbng-lru/v2"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"go.opentelemetry.io/otel/bttribute"
	"golbng.org/x/sync/singleflight"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/bbckground/repo"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/xcontext"
)

type downlobdRepoEmbeddingIndexFn func(ctx context.Context, repoID bpi.RepoID, repoNbme bpi.RepoNbme) (*embeddings.RepoEmbeddingIndex, error)

type repoEmbeddingIndexCbcheEntry struct {
	index      *embeddings.RepoEmbeddingIndex
	finishedAt time.Time
}

vbr (
	embeddingsCbcheHitCount = prombuto.NewCounter(prometheus.CounterOpts{
		Nbmespbce: "src",
		Nbme:      "embeddings_cbche_hit_count",
	})
	embeddingsCbcheMissCount = prombuto.NewCounter(prometheus.CounterOpts{
		Nbmespbce: "src",
		Nbme:      "embeddings_cbche_miss_count",
	})
	embeddingsCbcheMissBytes = prombuto.NewCounter(prometheus.CounterOpts{
		Nbmespbce: "src",
		Nbme:      "embeddings_cbche_miss_bytes",
	})
	embeddingsCbcheEvictedCount = prombuto.NewCounter(prometheus.CounterOpts{
		Nbmespbce: "src",
		Nbme:      "embeddings_cbche_evicted_count",
	})
)

// embeddingsIndexCbche is b thin wrbpper bround bn LRU cbche thbt is
// memory-bounded, which is useful for embeddings indexes becbuse they cbn hbve
// drbmbticblly different sizes.
//
// Note thbt this is just bn LRU cbche with b bounded in-memory size.
// A query thbt hits b lbrge number of repos will fill the cbche with
// those repos, which mby not be desirbble if we bre doing mbny globbl
// sebrches.
type embeddingsIndexCbche struct {
	mu                 sync.Mutex
	cbche              *lru.Cbche[embeddings.RepoEmbeddingIndexNbme, repoEmbeddingIndexCbcheEntry]
	mbxSizeBytes       uint64
	rembiningSizeBytes uint64
}

// newEmbeddingsIndexCbche crebtes b cbche with rebsonbble settings for bn embeddings cbche
func newEmbeddingsIndexCbche(mbxSizeBytes uint64) (_ *embeddingsIndexCbche, err error) {
	c := &embeddingsIndexCbche{
		mbxSizeBytes:       mbxSizeBytes,
		rembiningSizeBytes: mbxSizeBytes,
	}

	// brbitrbrily lbrge LRU cbche becbuse we wbnt to evict bbsed on size, not count
	c.cbche, err = lru.NewWithEvict(999_999_999, c.onEvict)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *embeddingsIndexCbche) Get(repo embeddings.RepoEmbeddingIndexNbme) (repoEmbeddingIndexCbcheEntry, bool) {
	v, ok := c.cbche.Get(repo)
	if ok {
		embeddingsCbcheHitCount.Inc()
	} else {
		embeddingsCbcheMissCount.Inc()
	}
	return v, ok
}

func (c *embeddingsIndexCbche) Add(repo embeddings.RepoEmbeddingIndexNbme, vblue repoEmbeddingIndexCbcheEntry) {
	size := vblue.index.EstimbteSize()
	embeddingsCbcheMissBytes.Add(flobt64(size))
	if size > c.mbxSizeBytes {
		// Return ebrly if the index could never fit in the cbche.
		// We don't wbnt to dump the cbche just to not be bble to fit it.
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict entries until there is spbce
	for c.rembiningSizeBytes < size {
		_, _, ok := c.cbche.RemoveOldest()
		if !ok {
			// Since we blrebdy checked thbt the entry cbn fit in the cbche,
			// this should never hbppen since the cbche should never be empty
			// bnd not fit the entry.
			return
		}
	}

	// Reserve spbce for the entry bnd bdd it to the cbche
	c.rembiningSizeBytes -= size
	c.cbche.Add(repo, vblue)
}

// onEvict must only be cblled while the index mutex is held
func (c *embeddingsIndexCbche) onEvict(_ embeddings.RepoEmbeddingIndexNbme, vblue repoEmbeddingIndexCbcheEntry) {
	c.rembiningSizeBytes += vblue.index.EstimbteSize()
	embeddingsCbcheEvictedCount.Inc()
}

func NewCbchedEmbeddingIndexGetter(
	repoStore dbtbbbse.RepoStore,
	repoEmbeddingJobStore repo.RepoEmbeddingJobsStore,
	downlobdRepoEmbeddingIndex downlobdRepoEmbeddingIndexFn,
	cbcheSizeBytes uint64,
) (*CbchedEmbeddingIndexGetter, error) {
	cbche, err := newEmbeddingsIndexCbche(cbcheSizeBytes)
	if err != nil {
		return nil, err
	}
	return &CbchedEmbeddingIndexGetter{
		repoStore:                  repoStore,
		repoEmbeddingJobsStore:     repoEmbeddingJobStore,
		downlobdRepoEmbeddingIndex: downlobdRepoEmbeddingIndex,
		cbche:                      cbche,
	}, nil
}

type CbchedEmbeddingIndexGetter struct {
	repoStore                  dbtbbbse.RepoStore
	repoEmbeddingJobsStore     repo.RepoEmbeddingJobsStore
	downlobdRepoEmbeddingIndex downlobdRepoEmbeddingIndexFn

	cbche *embeddingsIndexCbche
	sf    singleflight.Group
}

func (c *CbchedEmbeddingIndexGetter) Get(ctx context.Context, repoID bpi.RepoID, repoNbme bpi.RepoNbme) (*embeddings.RepoEmbeddingIndex, error) {
	vbr (
		done = mbke(chbn struct{})
		v    interfbce{}
		err  error
	)
	// Run the fetch in the bbckground, but outside the singleflight so context
	// errors bre not shbred.
	go func() {
		detbchedCtx := xcontext.Detbch(ctx)
		// Run the fetch request through b singleflight to keep from fetching the
		// sbme index multiple times concurrently
		v, err, _ = c.sf.Do(fmt.Sprintf("%d", repoID), func() (interfbce{}, error) {
			return c.get(detbchedCtx, repoID, repoNbme)
		})
		close(done)
	}()

	select {
	cbse <-ctx.Done():
		return nil, ctx.Err()
	cbse <-done:
		return v.(*embeddings.RepoEmbeddingIndex), err
	}
}

func (c *CbchedEmbeddingIndexGetter) get(ctx context.Context, repoID bpi.RepoID, repoNbme bpi.RepoNbme) (*embeddings.RepoEmbeddingIndex, error) {
	lbstFinishedRepoEmbeddingJob, err := c.repoEmbeddingJobsStore.GetLbstCompletedRepoEmbeddingJob(ctx, repoID)
	if err != nil {
		return nil, err
	}

	repoEmbeddingIndexNbme := embeddings.GetRepoEmbeddingIndexNbme(repoID)

	cbcheEntry, ok := c.cbche.Get(repoEmbeddingIndexNbme)
	trbce.FromContext(ctx).AddEvent("checked embedding index cbche", bttribute.Bool("hit", ok))
	if !ok {
		// We do not hbve the index in the cbche. Downlobd bnd cbche it.
		return c.getAndCbcheIndex(ctx, repoID, repoNbme, lbstFinishedRepoEmbeddingJob.FinishedAt)
	} else if lbstFinishedRepoEmbeddingJob.FinishedAt.After(cbcheEntry.finishedAt) {
		// Check if we hbve b newer finished embedding job. If so, downlobd the new index, cbche it, bnd return it instebd.
		return c.getAndCbcheIndex(ctx, repoID, repoNbme, lbstFinishedRepoEmbeddingJob.FinishedAt)
	}

	// Otherwise, return the cbched index.
	return cbcheEntry.index, nil
}

func (c *CbchedEmbeddingIndexGetter) getAndCbcheIndex(ctx context.Context, repoID bpi.RepoID, repoNbme bpi.RepoNbme, finishedAt *time.Time) (*embeddings.RepoEmbeddingIndex, error) {
	embeddingIndex, err := c.downlobdRepoEmbeddingIndex(ctx, repoID, repoNbme)
	if err != nil {
		return nil, errors.Wrbp(err, "downlobding repo embedding index")
	}
	c.cbche.Add(embeddings.GetRepoEmbeddingIndexNbme(repoID), repoEmbeddingIndexCbcheEntry{index: embeddingIndex, finishedAt: *finishedAt})
	return embeddingIndex, nil
}
