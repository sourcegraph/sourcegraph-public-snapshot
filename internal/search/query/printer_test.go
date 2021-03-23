package query

import (
	"testing"

	"github.com/hexops/autogold"
)

func TestStringHuman(t *testing.T) {
	test := func(input string) string {
		q, _ := ParseLiteral(input)
		return StringHuman(q)
	}

	autogold.Want("00", "a b c").Equal(t, test("a b c"))
	autogold.Want("01", "(this or that)").Equal(t, test("this or that"))
	autogold.Want("02", "(this or (that and here) or there)").Equal(t, test("this or that and here or there"))
	autogold.Want("03", "((not x) and y)").Equal(t, test("not x and y"))
	autogold.Want("04", "repo:foo (a and b)").Equal(t, test("repo:foo a and b"))
	autogold.Want("05", `repo:"quoted\""`).Equal(t, test(`repo:"quoted\""`))
	autogold.Want("06", `repo:"quoted"`).Equal(t, test(`repo:'quoted'`))
	autogold.Want("07", `"ab\"cd"`).Equal(t, test(`"ab\"cd"`))
	autogold.Want("08", "repo:foo file:bar (baz and qux)").Equal(t, test(`repo:foo file:bar baz and qux`))
	autogold.Want("09", "patterntype:regexp /abcd\\//").Equal(t, test(`/abcd\// patterntype:regexp`))
	autogold.Want("10", "repo:foo file:bar").Equal(t, test("repo:foo file:bar"))
	autogold.Want("11", "((repo:foo or repo:bar) or ((repo:baz or repo:qux) and (a or b)))").Equal(t, test("(repo:foo or repo:bar) or (repo:baz or repo:qux) (a or b)"))
	autogold.Want("12", "((repo:foo or repo:bar file:a) or ((repo:baz or repo:qux file:b) and a and b))").Equal(t, test("(repo:foo or repo:bar file:a) or (repo:baz or repo:qux and file:b) a and b"))
	autogold.Want("13", "repo:foo ((not b) and (not c) and a)").Equal(t, test("repo:foo a -content:b -content:c"))
	autogold.Want("14", "-repo:modspeed -file:pogspeed ((not Phoenicians) and Arizonan)").Equal(t, test("-repo:modspeed -file:pogspeed Arizonan -content:Phoenicians"))
}
