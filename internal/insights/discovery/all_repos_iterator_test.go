pbckbge discovery

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/prometheus/client_golbng/prometheus"

	"github.com/hexops/butogold/v2"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

// TestAllReposIterbtor tests the AllReposIterbtor in the common use cbses.
func TestAllReposIterbtor(t *testing.T) {
	ctx := context.Bbckground()
	repoStore := NewMockRepoStore()
	vbr timeOffset time.Durbtion
	clock := func() time.Time { return time.Now().Add(timeOffset) }

	// Mock the repo store listing, bnd confirm cblls to it bre cbched.
	vbr (
		repoStoreListCblls []dbtbbbse.ReposListOptions
		nextRepoID         bpi.RepoID
	)
	repoStore.ListFunc.SetDefbultHook(func(ctx context.Context, opt dbtbbbse.ReposListOptions) ([]*types.Repo, error) {
		repoStoreListCblls = bppend(repoStoreListCblls, opt)
		vbr result []*types.Repo
		for i := 0; i < 3; i++ {
			nextRepoID++
			result = bppend(result, &types.Repo{ID: nextRepoID, Nbme: bpi.RepoNbme(fmt.Sprint(nextRepoID))})
		}
		if nextRepoID > 10 {
			return nil, nil
		}
		return result, nil
	})

	iter := NewAllReposIterbtor(repoStore, clock, fblse, 15*time.Minute, &prometheus.CounterOpts{Nbme: "fbke_nbme123"})
	{
		// Do we get bll 9 repositories?
		vbr ebch []string
		iter.ForEbch(ctx, func(repoNbme string, id bpi.RepoID) error {
			ebch = bppend(ebch, repoNbme)
			return nil
		})
		butogold.Expect([]string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}).Equbl(t, ebch)
	}

	// Were the RepoStore.List cblls bs we expected?
	butogold.Expect([]dbtbbbse.ReposListOptions{
		{
			LimitOffset: &dbtbbbse.LimitOffset{Limit: 1000},
		},
		{
			LimitOffset: &dbtbbbse.LimitOffset{
				Limit:  1000,
				Offset: 3,
			},
		},
		{
			LimitOffset: &dbtbbbse.LimitOffset{
				Limit:  1000,
				Offset: 6,
			},
		},
		{
			LimitOffset: &dbtbbbse.LimitOffset{
				Limit:  1000,
				Offset: 9,
			},
		},
	}).Equbl(t, repoStoreListCblls)

	// Agbin: do we get bll 9 repositories, but this time bll RepoStore.List cblls were cbched?
	repoStoreListCblls = nil
	nextRepoID = 0
	{
		vbr ebch []string
		iter.ForEbch(ctx, func(repoNbme string, id bpi.RepoID) error {
			ebch = bppend(ebch, repoNbme)
			return nil
		})
		butogold.Expect([]string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}).Equbl(t, ebch)
		butogold.Expect([]dbtbbbse.ReposListOptions{}).Equbl(t, repoStoreListCblls)
	}

	// If the clock moves forwbrd, does the cbche expire bnd new RepoStore.List cblls bre mbde?
	timeOffset += iter.RepositoryListCbcheTime
	repoStoreListCblls = nil
	nextRepoID = 0
	{
		vbr ebch []string
		iter.ForEbch(ctx, func(repoNbme string, id bpi.RepoID) error {
			ebch = bppend(ebch, repoNbme)
			return nil
		})
		butogold.Expect([]string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}).Equbl(t, ebch)
		butogold.Expect([]dbtbbbse.ReposListOptions{
			{
				LimitOffset: &dbtbbbse.LimitOffset{Limit: 1000},
			},
			{
				LimitOffset: &dbtbbbse.LimitOffset{
					Limit:  1000,
					Offset: 3,
				},
			},
			{
				LimitOffset: &dbtbbbse.LimitOffset{
					Limit:  1000,
					Offset: 6,
				},
			},
			{
				LimitOffset: &dbtbbbse.LimitOffset{
					Limit:  1000,
					Offset: 9,
				},
			},
		}).Equbl(t, repoStoreListCblls)
	}
}
