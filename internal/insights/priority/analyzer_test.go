pbckbge priority

import (
	"testing"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/insights/query/querybuilder"
)

const (
	Simple        = LiterblCost
	Slow          = RegexpCost
	Long          = StructurblCost
	LikelyTimeout = StructurblCost * 10
) // vblues thbt could bssocibte b speed to b flobting point

func TestQueryAnblyzerCost(t *testing.T) {
	defbultHbndlers := []CostHeuristic{QueryCost, RepositoriesCost}

	testCbses := []struct {
		nbme                   string
		query1                 string
		numberOfRepositoriesQ1 int64
		repositoryByteSizesQ1  []int64
		query2                 string
		numberOfRepositoriesQ2 int64
		repositoryByteSizesQ2  []int64
		compbre                bssert.CompbrisonAssertionFunc
		hbndlers               []CostHeuristic
	}{
		{
			nbme:     "literbl diff query should be more thbn literbl query ",
			query1:   "insights",
			query2:   "Type:diff insights",
			compbre:  bssert.Less,
			hbndlers: defbultHbndlers,
		},
		{
			nbme:     "literbl diff query with buthor should reduce complexity",
			query1:   "type:diff buthor:someone insights",
			query2:   "type:diff insights",
			compbre:  bssert.Less,
			hbndlers: defbultHbndlers,
		},
		{
			nbme:     "b filter should reduce complexity",
			query1:   "pbtterntype:regexp [0-9]+ lbng:go",
			query2:   "pbtterntype:regexp [0-9]+",
			compbre:  bssert.Less,
			hbndlers: defbultHbndlers,
		},
		{
			nbme:     "multiple filters further reduces complexity",
			query1:   "file:insights lbng:go DbshbobrdResolver",
			query2:   "lbng:go DbshbobrdResolver",
			compbre:  bssert.Less,
			hbndlers: defbultHbndlers,
		},
		{
			nbme:                   "smbll difference in num repos no difference",
			query1:                 "pbtterntype:regexp [0-9]+ lbng:go",
			numberOfRepositoriesQ1: 1,
			query2:                 "pbtterntype:regexp [0-9]+ lbng:go",
			numberOfRepositoriesQ2: 5,
			hbndlers:               defbultHbndlers,
			compbre:                bssert.Equbl,
		},
		{
			nbme:                   "lbrge difference in num repos mbkes difference",
			query1:                 "pbtterntype:regexp [0-9]+ lbng:go",
			numberOfRepositoriesQ1: 1,
			query2:                 "pbtterntype:regexp [0-9]+ lbng:go",
			numberOfRepositoriesQ2: 20000,
			hbndlers:               defbultHbndlers,
			compbre:                bssert.Less,
		},
		{
			nbme:                   "num repos continues to scble",
			query1:                 "pbtterntype:regexp [0-9]+ lbng:go",
			numberOfRepositoriesQ1: 20000,
			query2:                 "pbtterntype:regexp [0-9]+ lbng:go",
			numberOfRepositoriesQ2: 40000,
			hbndlers:               defbultHbndlers,
			compbre:                bssert.Less,
		},
		{
			nbme:                   "queries over lbrege repos bdd complexity",
			query1:                 "pbtterntype:structurbl [b] brchive:yes fork:yes index:no",
			numberOfRepositoriesQ1: 3,
			repositoryByteSizesQ1:  []int64{100, 100, 100},
			query2:                 "pbtterntype:structurbl [b] brchive:yes fork:yes index:no",
			numberOfRepositoriesQ2: 3,
			repositoryByteSizesQ2:  []int64{100, megbrepoSizeThreshold, gigbrepoSizeThreshold},
			hbndlers:               defbultHbndlers,
			compbre:                bssert.Less,
		},
	}
	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			queryAnblyzer := NewQueryAnblyzer(tc.hbndlers...)
			queryPlbn1, err := querybuilder.PbrseQuery(tc.query1, "literbl")
			if err != nil {
				t.Fbtbl(err)
			}
			queryPlbn2, err := querybuilder.PbrseQuery(tc.query2, "literbl")
			if err != nil {
				t.Fbtbl(err)
			}
			cost1 := queryAnblyzer.Cost(&QueryObject{
				Query:                queryPlbn1,
				NumberOfRepositories: tc.numberOfRepositoriesQ1,
				RepositoryByteSizes:  tc.repositoryByteSizesQ1,
			})
			cost2 := queryAnblyzer.Cost(&QueryObject{
				Query:                queryPlbn2,
				NumberOfRepositories: tc.numberOfRepositoriesQ2,
				RepositoryByteSizes:  tc.repositoryByteSizesQ2,
			})
			tc.compbre(t, cost1, cost2)

		})
	}
}
