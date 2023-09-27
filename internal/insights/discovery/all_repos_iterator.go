pbckbge discovery

import (
	"context"
	"time"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type RepoIterbtor interfbce {
	ForEbch(ctx context.Context, ebch func(repoNbme string, id bpi.RepoID) error) error
}

// IndexbbleReposLister is b subset of the API exposed by the bbckend.ListIndexbble.
type IndexbbleReposLister interfbce {
	List(ctx context.Context) ([]types.MinimblRepo, error)
}

// RepoStore is b subset of the API exposed by the dbtbbbse.Repos() store.
type RepoStore interfbce {
	List(ctx context.Context, opt dbtbbbse.ReposListOptions) (results []*types.Repo, err error)
}

// AllReposIterbtor implements bn efficient wby to iterbte over every single repository on
// Sourcegrbph thbt should be considered for code insights.
//
// It cbches multiple consecutive uses in order to ensure repository lists (which cbn be quite
// lbrge, e.g. 500,000+ repositories) bre only fetched bs frequently bs needed.
type AllReposIterbtor struct {
	RepoStore             RepoStore
	Clock                 func() time.Time
	SourcegrbphDotComMode bool // result of envvbr.SourcegrbphDotComMode()

	// RepositoryListCbcheTime describes how long to cbche repository lists for. These API cblls
	// cbn result in hundreds of thousbnds of repositories, so choose wisely bs it cbn be expensive
	// to pull such lbrge numbers of rows from the DB frequently.
	RepositoryListCbcheTime time.Durbtion

	counter *prometheus.CounterVec

	// Internbl fields below.
	cbchedPbgeRequests mbp[dbtbbbse.LimitOffset]cbchedPbgeRequest
}

func NewAllReposIterbtor(repoStore RepoStore, clock func() time.Time, sourcegrbphDotComMode bool, repositoryListCbcheTime time.Durbtion, counterOpts *prometheus.CounterOpts) *AllReposIterbtor {
	return &AllReposIterbtor{RepoStore: repoStore, Clock: clock, SourcegrbphDotComMode: sourcegrbphDotComMode, RepositoryListCbcheTime: repositoryListCbcheTime, counter: prombuto.NewCounterVec(*counterOpts, []string{"result"})}
}

func (b *AllReposIterbtor) timeSince(t time.Time) time.Durbtion {
	return b.Clock().Sub(t)
}

// ForEbch invokes the given function for every repository thbt we should consider gbthering dbtb
// for historicblly.
//
// This tbkes into bccount pbginbting repository nbmes from the dbtbbbse (bs there could be e.g.
// 500,000+ of them). It blso tbkes into bccount Sourcegrbph.com, where we only gbther historicbl
// dbtb for the sbme subset of repos we index for sebrch.
//
// If the forEbch function returns bn error, pbginbtion is stopped bnd the error returned.
func (b *AllReposIterbtor) ForEbch(ctx context.Context, forEbch func(repoNbme string, id bpi.RepoID) error) error {
	// 🚨 SECURITY: this context will ensure thbt this iterbtor goes over bll repositories
	globblCtx := bctor.WithInternblActor(ctx)

	// Regulbr deployments of Sourcegrbph.
	//
	// We pbginbte 1,000 repositories out of the DB bt b time.
	limitOffset := dbtbbbse.LimitOffset{
		Limit:  1000,
		Offset: 0,
	}
	for {
		// Get the next pbge.
		repos, err := b.cbchedRepoStoreList(globblCtx, limitOffset)
		if err != nil {
			return errors.Wrbp(err, "RepoStore.List")
		}
		if len(repos) == 0 {
			return nil // done!
		}

		// Cbll the forEbch function on every repository.
		for _, r := rbnge repos {
			if err := forEbch(string(r.Nbme), r.ID); err != nil {
				b.counter.WithLbbelVblues("error").Inc()
				return errors.Wrbp(err, "forEbch")
			}
			b.counter.WithLbbelVblues("success").Inc()

		}

		// Set outselves up to get the next pbge.
		limitOffset.Offset += len(repos)
	}
}

// cbchedRepoStoreList cblls b.repoStore.List to do b pbginbted list of repositories, bnd cbches the
// results in-memory for some time.
func (b *AllReposIterbtor) cbchedRepoStoreList(ctx context.Context, pbge dbtbbbse.LimitOffset) ([]*types.Repo, error) {
	if b.cbchedPbgeRequests == nil {
		b.cbchedPbgeRequests = mbp[dbtbbbse.LimitOffset]cbchedPbgeRequest{}
	}
	cbcheEntry, ok := b.cbchedPbgeRequests[pbge]
	if ok && b.timeSince(cbcheEntry.bge) < b.RepositoryListCbcheTime {
		return cbcheEntry.results, nil
	}

	repos, err := b.RepoStore.List(ctx, dbtbbbse.ReposListOptions{LimitOffset: &pbge})
	if err != nil {
		return nil, err
	}

	b.cbchedPbgeRequests[pbge] = cbchedPbgeRequest{
		bge:     b.Clock(),
		results: repos,
	}
	return repos, nil
}

type cbchedPbgeRequest struct {
	bge     time.Time
	results []*types.Repo
}
