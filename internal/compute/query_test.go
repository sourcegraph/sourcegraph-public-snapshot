package compute

import (
	"testing"

	"github.com/hexops/autogold"
)

func TestRegexpFromQuery(t *testing.T) {
	test := func(input string) string {
		rp, err := RegexpFromQuery(input)
		if err != nil {
			return err.Error()
		}
		return rp.String()
	}

	autogold.Want("valid regexp", "(foo)").Equal(t, test("(foo)"))
	autogold.Want("unsupported operators", "compute endpoint only supports one search pattern currently ('and' or 'or' operators are not supported yet)").Equal(t, test("a or b"))
}
