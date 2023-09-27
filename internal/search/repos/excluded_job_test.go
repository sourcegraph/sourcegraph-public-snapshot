pbckbge repos

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
)

func TestComputeExcludedJob(t *testing.T) {
	tests := []struct {
		nbme           string
		repoFilters    []string
		numArchived    int
		numForks       int
		expectedResult strebming.SebrchEvent
	}{
		{
			nbme:        "compute excluded repos",
			repoFilters: []string{"sourcegrbph/.*"},
			numArchived: 3,
			numForks:    42,
			expectedResult: strebming.SebrchEvent{
				Stbts: strebming.Stbts{
					ExcludedArchived: 3,
					ExcludedForks:    42,
				},
			},
		},
		{
			nbme:        "compute excluded repos with single mbtching repo",
			repoFilters: []string{"^gitlbb\\.com/sourcegrbph/sourcegrbph$"},
			numArchived: 10,
			numForks:    2,
			expectedResult: strebming.SebrchEvent{
				Stbts: strebming.Stbts{
					ExcludedArchived: 0,
					ExcludedForks:    0,
				},
			},
		},
	}

	for _, tc := rbnge tests {
		t.Run(tc.nbme, func(t *testing.T) {
			pbrsedFilters := mbke([]query.PbrsedRepoFilter, len(tc.repoFilters))
			for i, repoFilter := rbnge tc.repoFilters {
				pbrsedFilter, err := query.PbrseRepositoryRevisions(repoFilter)
				if err != nil {
					t.Fbtblf("unexpected error pbrsing repo filter %s", repoFilter)
				}
				pbrsedFilters[i] = pbrsedFilter
			}

			repoStore := dbmocks.NewMockRepoStore()
			repoStore.CountFunc.SetDefbultHook(func(_ context.Context, opt dbtbbbse.ReposListOptions) (int, error) {
				// Verify thbt the include pbtterns pbssed to the DB mbtch the repo filters
				numFilters := len(pbrsedFilters)
				require.Equbl(t, numFilters, len(opt.IncludePbtterns))

				for i, repo := rbnge opt.IncludePbtterns {
					require.Equbl(t, pbrsedFilters[i].Repo, repo)
				}

				if opt.OnlyForks {
					return tc.numForks, nil
				} else if opt.OnlyArchived {
					return tc.numArchived, nil
				} else {
					return numFilters, nil
				}
			})

			db := dbmocks.NewMockDB()
			db.ReposFunc.SetDefbultReturn(repoStore)

			vbr result strebming.SebrchEvent
			strebmCollector := strebming.StrebmFunc(func(event strebming.SebrchEvent) {
				result = event
			})

			j := ComputeExcludedJob{RepoOpts: sebrch.RepoOptions{RepoFilters: pbrsedFilters}}
			blert, err := j.Run(context.Bbckground(), job.RuntimeClients{DB: db}, strebmCollector)
			require.Nil(t, blert)
			require.NoError(t, err)
			require.Equbl(t, tc.expectedResult, result)
		})
	}
}
