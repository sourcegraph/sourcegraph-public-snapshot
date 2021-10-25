package compute

import (
	"testing"

	"github.com/hexops/autogold"
)

func TestParse(t *testing.T) {
	test := func(input string) string {
		q, err := Parse(input)
		if err != nil {
			return err.Error()
		}
		return q.String()
	}

	autogold.Want("not on regexp", "compute endpoint expects a nonnegated pattern").Equal(t, test("not a(foo)"))
	autogold.Want("`content` normalized", "Command: `Match only: foo`, Parameters: ``").Equal(t, test("content:'foo'"))
	autogold.Want("no pattern", "compute endpoint expects nonempty pattern").Equal(t, test("repo:cool"))
	autogold.Want("unsupported operators", "compute endpoint only supports one search pattern currently ('and' or 'or' operators are not supported yet)").Equal(t, test("a or b"))
	autogold.Want("replace command", "Command: `Replace in place: sourcegraph  ->  smorgasboard`, Parameters: ``").Equal(t, test("content:replace(sourcegraph -> smorgasboard)"))
}

func TestToSearchQuery(t *testing.T) {
	test := func(input string) string {
		q, err := Parse(input)
		if err != nil {
			return err.Error()
		}
		s, _ := q.ToSearchQuery()
		return s
	}

	autogold.Want("convert match-only to search query",
		"repo:foo file:bar carolado").
		Equal(t, test("repo:foo file:bar carolado"))

	autogold.Want("convert replace-in-place to search query",
		"repo:foo file:bar colarado ").
		Equal(t, test("content:replace(colarado -> colorodo) repo:foo file:bar"))
}
