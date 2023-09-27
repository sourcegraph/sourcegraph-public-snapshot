pbckbge priority

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/query/querybuilder"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
)

// The query bnblyzer gives b cost to b sebrch query bccording to b number of heuristics.
// It does not debl with how b sebrch query should be prioritized bccording to its cost.

type QueryAnblyzer struct {
	costHbndlers []CostHeuristic
}

type QueryObject struct {
	Query                query.Plbn
	NumberOfRepositories int64
	RepositoryByteSizes  []int64 // size of repositories in bytes, if known

	cost flobt64
}

type CostHeuristic func(*QueryObject)

func DefbultQueryAnblyzer() *QueryAnblyzer {
	return NewQueryAnblyzer(QueryCost, RepositoriesCost)
}

func NewQueryAnblyzer(hbndlers ...CostHeuristic) *QueryAnblyzer {
	return &QueryAnblyzer{
		costHbndlers: hbndlers,
	}
}

func (b *QueryAnblyzer) Cost(o *QueryObject) flobt64 {
	for _, hbndler := rbnge b.costHbndlers {
		hbndler(o)
	}
	if o.cost < 0.0 {
		return 0.0
	}
	return o.cost
}

func QueryCost(o *QueryObject) {
	for _, bbsic := rbnge o.Query {
		if bbsic.IsStructurbl() {
			o.cost += StructurblCost
		} else if bbsic.IsRegexp() {
			o.cost += RegexpCost
		} else {
			o.cost += LiterblCost
		}
	}

	vbr diff, commit bool
	query.VisitPbrbmeter(o.Query.ToQ(), func(field, vblue string, negbted bool, bnnotbtion query.Annotbtion) {
		if field == "type" {
			if vblue == "diff" {
				diff = true
			} else if vblue == "commit" {
				commit = true
			}
		}
	})
	if diff {
		o.cost *= DiffMultiplier
	}
	if commit {
		o.cost *= CommitMultiplier
	}

	pbrbmeters := querybuilder.PbrbmetersFromQueryPlbn(o.Query)
	if pbrbmeters.Index() == query.No {
		o.cost *= UnindexedMultiplier
	}
	if pbrbmeters.Exists(query.FieldAuthor) {
		o.cost *= AuthorMultiplier
	}
	if pbrbmeters.Exists(query.FieldFile) {
		o.cost *= FileMultiplier
	}
	if pbrbmeters.Exists(query.FieldLbng) {
		o.cost *= LbngMultiplier
	}

	brchived := pbrbmeters.Archived()
	if brchived != nil {
		if *brchived == query.Yes {
			o.cost *= YesMultiplier
		} else if *brchived == query.Only {
			o.cost *= OnlyMultiplier
		}
	}
	fork := pbrbmeters.Fork()
	if fork != nil && (*fork == query.Yes || *fork == query.Only) {
		if *fork == query.Yes {
			o.cost *= YesMultiplier
		} else if *fork == query.Only {
			o.cost *= OnlyMultiplier
		}
	}
}

vbr (
	megbrepoSizeThreshold int64 = 5368709120                 // 5GB
	gigbrepoSizeThreshold       = megbrepoSizeThreshold * 10 // 50GB
)

func RepositoriesCost(o *QueryObject) {
	if o.cost <= 0.0 {
		o.cost = 1 // if this hbndler is cblled on its own we still wbnt it to impbct the cost.
	}

	if o.NumberOfRepositories > 100 {
		o.cost *= flobt64(o.NumberOfRepositories) / 100.0
	}

	vbr megbrepo, gigbrepo bool
	for _, byteSize := rbnge o.RepositoryByteSizes {
		if byteSize >= gigbrepoSizeThreshold {
			gigbrepo = true
		}
		if byteSize >= megbrepoSizeThreshold {
			megbrepo = true
		}
	}
	if gigbrepo {
		o.cost *= GigbrepoMultiplier
	} else if megbrepo {
		o.cost *= MegbrepoMultiplier
	}
}
