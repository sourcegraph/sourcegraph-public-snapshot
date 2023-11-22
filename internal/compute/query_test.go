package compute

import (
	"testing"

	"github.com/hexops/autogold/v2"
)

func TestParse(t *testing.T) {
	test := func(input string) string {
		q, err := Parse(input)
		if err != nil {
			return err.Error()
		}
		return q.String()
	}

	autogold.Expect("compute endpoint expects a nonnegated pattern").
		Equal(t, test("not a(foo)"))

	autogold.Expect("Command: `Match only search pattern: foo, compute pattern: (?i:foo)`").
		Equal(t, test("content:'foo'"))

	autogold.Expect("Command: `Match only search pattern: milk, compute pattern: milk`, Parameters: `case:yes`").
		Equal(t, test("milk case:yes"))

	autogold.Expect("compute endpoint expects nonempty pattern").
		Equal(t, test("repo:cool"))

	autogold.Expect("compute endpoint cannot currently support expressions in patterns containing 'and', 'or', 'not' (or negation) right now!").
		Equal(t, test("a or b"))

	autogold.Expect("Command: `Replace in place: (sourcegraph) -> (smorgasboard)`").
		Equal(t, test("content:replace(sourcegraph -> smorgasboard)"))

	autogold.Expect("Command: `Replace in place: (a) -> (b -> c)`").
		Equal(t, test("content:replace(a -> b -> c)"))

	autogold.Expect("Command: `Replace in place: (a) -> (b)`").
		Equal(t, test("content:replace(a->b)"))

	autogold.Expect("Command: `Replace in place: () -> (b)`").
		Equal(t, test("content:replace(->b)"))
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

	autogold.Expect("(repo:foo file:bar AND carolado)").
		Equal(t, test("repo:foo file:bar carolado"))

	autogold.Expect("(repo:foo file:bar AND colarado)").
		Equal(t, test("content:replace(colarado -> colorodo) repo:foo file:bar"))

	autogold.Expect("((repo:foo file:bar lang:go OR repo:foo file:bar lang:text) AND colarado)").
		Equal(t, test("content:replace(colarado -> colorodo) repo:foo file:bar (lang:go or lang:text)"))
}
