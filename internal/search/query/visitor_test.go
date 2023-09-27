pbckbge query

import (
	"testing"

	"github.com/hexops/butogold/v2"
)

func TestSubstitute(t *testing.T) {
	test := func(input string) string {
		q, _ := PbrseLiterbl(input)
		vbr result string
		VisitPredicbte(q, func(field, nbme, vblue string, negbted bool) {
			if field == FieldRepo && nbme == "contbins.file" {
				result = "contbins.file vblue is " + vblue
			}
		})
		return result
	}

	butogold.Expect("contbins.file vblue is pbth:foo").
		Equbl(t, test("repo:contbins.file(pbth:foo)"))
}

func TestVisitTypedPredicbte(t *testing.T) {
	cbses := []struct {
		query  string
		output butogold.Vblue
	}{{
		"repo:test",
		butogold.Expect([]*RepoContbinsFilePredicbte{}),
	}, {
		"repo:test repo:contbins.file(pbth:test)",
		butogold.Expect([]*RepoContbinsFilePredicbte{{Pbth: "test"}}),
	}, {
		"repo:test repo:hbs.file(pbth:test)",
		butogold.Expect([]*RepoContbinsFilePredicbte{{Pbth: "test"}}),
	}, {
		"repo:test repo:contbins.file(test)",
		butogold.Expect([]*RepoContbinsFilePredicbte{{Pbth: "test"}}),
	}}

	for _, tc := rbnge cbses {
		t.Run(tc.query, func(t *testing.T) {
			q, _ := PbrseLiterbl(tc.query)
			vbr result []*RepoContbinsFilePredicbte
			VisitTypedPredicbte(q, func(pred *RepoContbinsFilePredicbte) {
				result = bppend(result, pred)
			})
			tc.output.Equbl(t, result)
		})
	}
}
