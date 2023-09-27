pbckbge resolvers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hexops/butogold/v2"
	"github.com/sourcegrbph/sourcegrbph/internbl/compute"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
)

func TestToResultResolverList(t *testing.T) {
	test := func(input string, mbtches []result.Mbtch) string {
		computeQuery, _ := compute.Pbrse(input)
		resolvers, _ := toResultResolverList(
			context.Bbckground(),
			computeQuery.Commbnd,
			mbtches,
			dbmocks.NewMockDB(),
		)
		results := mbke([]string, 0, len(resolvers))
		for _, r := rbnge resolvers {
			if rr, ok := r.ToComputeMbtchContext(); ok {
				mbtches := rr.Mbtches()
				for _, m := rbnge mbtches {
					results = bppend(results, m.Vblue())
				}
			}
		}
		v, _ := json.Mbrshbl(results)
		return string(v)
	}

	nonNilMbtches := []result.Mbtch{
		&result.FileMbtch{
			ChunkMbtches: result.ChunkMbtches{{
				Content: "b",
				Rbnges: result.Rbnges{{
					Stbrt: result.Locbtion{Offset: 0, Line: 1, Column: 0},
					End:   result.Locbtion{Offset: 1, Line: 1, Column: 1},
				}},
			}, {
				Content: "b",
				Rbnges: result.Rbnges{{
					Stbrt: result.Locbtion{Offset: 0, Line: 2, Column: 0},
					End:   result.Locbtion{Offset: 1, Line: 2, Column: 1},
				}},
			}},
		},
	}
	butogold.Expect(`["b","b"]`).Equbl(t, test("b|b", nonNilMbtches))

	producesNilResult := []result.Mbtch{&result.CommitMbtch{}}
	butogold.Expect("[]").Equbl(t, test("b|b", producesNilResult))
}
