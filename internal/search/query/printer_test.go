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
	autogold.Want("01", "(this OR that)").Equal(t, test("this or that"))
	autogold.Want("02", "(this OR (that AND here) OR there)").Equal(t, test("this or that and here or there"))
	autogold.Want("03", "((NOT x) AND y)").Equal(t, test("not x and y"))
	autogold.Want("04", "repo:foo (a AND b)").Equal(t, test("repo:foo a and b"))
	autogold.Want("05", `repo:"quoted\""`).Equal(t, test(`repo:"quoted\""`))
	autogold.Want("06", `repo:"quoted"`).Equal(t, test(`repo:'quoted'`))
	autogold.Want("07", `"ab\"cd"`).Equal(t, test(`"ab\"cd"`))
	autogold.Want("08", "repo:foo file:bar (baz AND qux)").Equal(t, test(`repo:foo file:bar baz and qux`))
	autogold.Want("09", "patterntype:regexp /abcd\\//").Equal(t, test(`/abcd\// patterntype:regexp`))
	autogold.Want("10", "repo:foo file:bar").Equal(t, test("repo:foo file:bar"))
	autogold.Want("11", "((repo:foo OR repo:bar) OR ((repo:baz OR repo:qux) AND (a OR b)))").Equal(t, test("(repo:foo or repo:bar) or (repo:baz or repo:qux) (a or b)"))
	autogold.Want("12", "((repo:foo OR repo:bar file:a) OR ((repo:baz OR repo:qux file:b) AND a AND b))").Equal(t, test("(repo:foo or repo:bar file:a) or (repo:baz or repo:qux and file:b) a and b"))
	autogold.Want("13", "repo:foo ((NOT b) AND (NOT c) AND a)").Equal(t, test("repo:foo a -content:b -content:c"))
	autogold.Want("14", "-repo:modspeed -file:pogspeed ((NOT Phoenicians) AND Arizonan)").Equal(t, test("-repo:modspeed -file:pogspeed Arizonan -content:Phoenicians"))

	test = func(input string) string {
		q, _ := ParseRegexp(input)
		return StringHuman(q)
	}

	autogold.Want("15", "/bo/u\\gros/").Equal(t, test(`bo/u\gros`))
}
