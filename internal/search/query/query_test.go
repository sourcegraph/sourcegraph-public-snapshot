package query

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hexops/autogold"
)

func TestPipeline(t *testing.T) {
	planToString := func(nodes [][]Node) string {
		var plan Plan
		for _, node := range nodes {
			plan = append(plan, Q(node))
		}

		var result []string
		for _, q := range plan {
			result = append(result, toString(q))
		}
		return strings.Join(result, " ")
	}

	// The Pipeline must produce a value that is equivalent under DNF to parsing the query and processing it with DNF.
	test := func(input string) string {
		pipelinePlan, _ := Pipeline(InitLiteral(input))
		nodes, _ := Run(InitLiteral(input))
		if diff := cmp.Diff(
			planToString(Dnf(pipelinePlan.ToParseTree())),
			planToString(Dnf(nodes)),
		); diff != "" {
			return diff
		}
		return "equivalent"
	}

	autogold.Want("equivalent or-expression", "equivalent").Equal(t, test("(repo:bob or repo:jim) ((rev:olga or rev:ham) demo123232)"))
}
