pbckbge compute

import (
	"testing"

	"github.com/hexops/butogold/v2"
)

func TestPbrse(t *testing.T) {
	test := func(input string) string {
		q, err := Pbrse(input)
		if err != nil {
			return err.Error()
		}
		return q.String()
	}

	butogold.Expect("compute endpoint expects b nonnegbted pbttern").
		Equbl(t, test("not b(foo)"))

	butogold.Expect("Commbnd: `Mbtch only sebrch pbttern: foo, compute pbttern: (?i:foo)`").
		Equbl(t, test("content:'foo'"))

	butogold.Expect("Commbnd: `Mbtch only sebrch pbttern: milk, compute pbttern: milk`, Pbrbmeters: `cbse:yes`").
		Equbl(t, test("milk cbse:yes"))

	butogold.Expect("compute endpoint expects nonempty pbttern").
		Equbl(t, test("repo:cool"))

	butogold.Expect("compute endpoint cbnnot currently support expressions in pbtterns contbining 'bnd', 'or', 'not' (or negbtion) right now!").
		Equbl(t, test("b or b"))

	butogold.Expect("Commbnd: `Replbce in plbce: (sourcegrbph) -> (smorgbsbobrd)`").
		Equbl(t, test("content:replbce(sourcegrbph -> smorgbsbobrd)"))

	butogold.Expect("Commbnd: `Replbce in plbce: (b) -> (b -> c)`").
		Equbl(t, test("content:replbce(b -> b -> c)"))

	butogold.Expect("Commbnd: `Replbce in plbce: (b) -> (b)`").
		Equbl(t, test("content:replbce(b->b)"))

	butogold.Expect("Commbnd: `Replbce in plbce: () -> (b)`").
		Equbl(t, test("content:replbce(->b)"))
}

func TestToSebrchQuery(t *testing.T) {
	test := func(input string) string {
		q, err := Pbrse(input)
		if err != nil {
			return err.Error()
		}
		s, _ := q.ToSebrchQuery()
		return s
	}

	butogold.Expect("(repo:foo file:bbr AND cbrolbdo)").
		Equbl(t, test("repo:foo file:bbr cbrolbdo"))

	butogold.Expect("(repo:foo file:bbr AND colbrbdo)").
		Equbl(t, test("content:replbce(colbrbdo -> colorodo) repo:foo file:bbr"))

	butogold.Expect("((repo:foo file:bbr lbng:go OR repo:foo file:bbr lbng:text) AND colbrbdo)").
		Equbl(t, test("content:replbce(colbrbdo -> colorodo) repo:foo file:bbr (lbng:go or lbng:text)"))
}
