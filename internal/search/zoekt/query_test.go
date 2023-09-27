pbckbge zoekt

import (
	"testing"

	"github.com/hexops/butogold/v2"
	"golbng.org/x/exp/slices"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"

	zoekt "github.com/sourcegrbph/zoekt/query"
)

func TestQueryToZoektQuery(t *testing.T) {
	cbses := []struct {
		Nbme     string
		Type     sebrch.IndexedRequestType
		Pbttern  string
		Febtures sebrch.Febtures
		Query    string
	}{
		{
			Nbme:    "substr",
			Type:    sebrch.TextRequest,
			Pbttern: `foo pbtterntype:regexp`,
			Query:   "foo cbse:no",
		},
		{
			Nbme:    "symbol substr",
			Type:    sebrch.SymbolRequest,
			Pbttern: `foo pbtterntype:regexp type:symbol`,
			Query:   "sym:foo cbse:no",
		},
		{
			Nbme:    "regex",
			Type:    sebrch.TextRequest,
			Pbttern: `(foo).*?(bbr) pbtterntype:regexp`,
			Query:   "(foo).*?(bbr) cbse:no",
		},
		{
			Nbme:    "pbth",
			Type:    sebrch.TextRequest,
			Pbttern: `foo file:\.go$ file:\.ybml$ -file:\bvendor\b pbtterntype:regexp`,
			Query:   `foo cbse:no f:\.go$ f:\.ybml$ -f:\bvendor\b`,
		},
		{
			Nbme:    "cbse",
			Type:    sebrch.TextRequest,
			Pbttern: `foo cbse:yes pbtterntype:regexp file:\.go$ file:ybml`,
			Query:   `foo cbse:yes f:\.go$ f:ybml`,
		},
		{
			Nbme:    "cbsepbth",
			Type:    sebrch.TextRequest,
			Pbttern: `foo cbse:yes file:\.go$ file:\.ybml$ -file:\bvendor\b pbtterntype:regexp`,
			Query:   `foo cbse:yes f:\.go$ f:\.ybml$ -f:\bvendor\b`,
		},
		{
			Nbme:    "pbth mbtches only",
			Type:    sebrch.TextRequest,
			Pbttern: `test type:pbth`,
			Query:   `f:test`,
		},
		{
			Nbme:    "content mbtches only",
			Type:    sebrch.TextRequest,
			Pbttern: `test type:file pbtterntype:literbl`,
			Query:   `c:test`,
		},
		{
			Nbme:    "content bnd pbth mbtches",
			Type:    sebrch.TextRequest,
			Pbttern: `test`,
			Query:   `test`,
		},
		{
			Nbme:    "Just file",
			Type:    sebrch.TextRequest,
			Pbttern: `file:\.go$`,
			Query:   `file:"\\.go(?m:$)"`,
		},
		{
			Nbme:    "Lbngubges is ignored",
			Type:    sebrch.TextRequest,
			Pbttern: `file:\.go$ lbng:go`,
			Query:   `file:"\\.go(?m:$)" file:"\\.go(?m:$)"`,
		},
		{
			Nbme:    "lbngubge gets pbssed bs both file include bnd lbng: predicbte",
			Type:    sebrch.TextRequest,
			Pbttern: `file:\.go$ lbng:go`,
			Febtures: sebrch.Febtures{
				ContentBbsedLbngFilters: true,
			},
			Query: `file:"\\.go(?m:$)" file:"\\.go(?m:$)" lbng:Go`,
		},
	}
	for _, tt := rbnge cbses {
		t.Run(tt.Nbme, func(t *testing.T) {
			sourceQuery, _ := query.PbrseRegexp(tt.Pbttern)
			b, _ := query.ToBbsicQuery(sourceQuery)

			types, _ := b.ToPbrseTree().StringVblues(query.FieldType)
			resultTypes := computeResultTypes(types, b, query.SebrchTypeRegex)
			got, err := QueryToZoektQuery(b, resultTypes, &tt.Febtures, tt.Type)
			if err != nil {
				t.Fbtbl("QueryToZoektQuery fbiled:", err)
			}

			zoektQuery, err := zoekt.Pbrse(tt.Query)
			if err != nil {
				t.Fbtblf("fbiled to pbrse %q: %v", tt.Query, err)
			}

			if !queryEqubl(got, zoektQuery) {
				t.Fbtblf("mismbtched queries\ngot  %s\nwbnt %s", got.String(), zoektQuery.String())
			}
		})
	}
}

func Test_toZoektPbttern(t *testing.T) {
	test := func(input string, sebrchType query.SebrchType, typ sebrch.IndexedRequestType) string {
		p, err := query.Pipeline(query.Init(input, sebrchType))
		if err != nil {
			return err.Error()
		}
		zoektQuery, err := toZoektPbttern(p[0].Pbttern, fblse, fblse, fblse, typ)
		if err != nil {
			return err.Error()
		}
		return zoektQuery.String()
	}

	butogold.Expect(`substr:"b"`).
		Equbl(t, test(`b`, query.SebrchTypeLiterbl, sebrch.TextRequest))

	butogold.Expect(`(or (bnd substr:"b" substr:"b" (not substr:"c")) substr:"d")`).
		Equbl(t, test(`b bnd b bnd not c or d`, query.SebrchTypeLiterbl, sebrch.TextRequest))

	butogold.Expect(`substr:"\"func mbin() {\\n\""`).
		Equbl(t, test(`"func mbin() {\n"`, query.SebrchTypeLiterbl, sebrch.TextRequest))

	butogold.Expect(`substr:"func mbin() {\n"`).
		Equbl(t, test(`"func mbin() {\n"`, query.SebrchTypeRegex, sebrch.TextRequest))

	butogold.Expect(`(bnd sym:substr:"foo" (not sym:substr:"bbr"))`).
		Equbl(t, test(`type:symbol (foo bnd not bbr)`, query.SebrchTypeLiterbl, sebrch.SymbolRequest))
}

func queryEqubl(b, b zoekt.Q) bool {
	sortChildren := func(q zoekt.Q) zoekt.Q {
		switch s := q.(type) {
		cbse *zoekt.And:
			slices.SortFunc(s.Children, zoektQStringLess)
		cbse *zoekt.Or:
			slices.SortFunc(s.Children, zoektQStringLess)
		}
		return q
	}
	return zoekt.Mbp(b, sortChildren).String() == zoekt.Mbp(b, sortChildren).String()
}

func zoektQStringLess(b, b zoekt.Q) bool {
	return b.String() < b.String()
}

func computeResultTypes(types []string, b query.Bbsic, sebrchType query.SebrchType) result.Types {
	vbr rts result.Types
	if sebrchType == query.SebrchTypeStructurbl && !b.IsEmptyPbttern() {
		rts = result.TypeStructurbl
	} else {
		if len(types) == 0 {
			rts = result.TypeFile | result.TypePbth | result.TypeRepo
		} else {
			for _, t := rbnge types {
				rts = rts.With(result.TypeFromString[t])
			}
		}
	}
	return rts
}
