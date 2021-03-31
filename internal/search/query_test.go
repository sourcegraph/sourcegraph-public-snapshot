package search

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hexops/autogold"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
)

func TestPipeline(t *testing.T) {
	planToString := func(disjuncts [][]query.Node) string {
		var plan query.Plan
		for _, disjunct := range disjuncts {
			parameters, pattern, _ := query.PartitionSearchPattern(disjunct)
			plan = append(plan, query.Basic{Parameters: parameters, Pattern: pattern})
		}

		var result []string
		for _, basic := range plan {
			result = append(result, query.ToString(basic.ToParseTree()))
		}
		return strings.Join(result, " ")
	}

	// The Pipeline must produce a value that is equivalent under DNF to parsing the query and processing it with DNF.
	test := func(input string) string {
		pipelinePlan, _ := Pipeline(query.InitLiteral(input))
		nodes, _ := query.Run(query.InitLiteral(input))
		disjuncts := query.Dnf(nodes)
		plan, _ := query.ToPlan(disjuncts)
		manualPlan := query.MapPlan(plan, query.ConcatRevFilters)
		if diff := cmp.Diff(
			planToString(query.Dnf(pipelinePlan.ToParseTree())),
			planToString(query.Dnf(manualPlan.ToParseTree())),
		); diff != "" {
			return diff
		}
		return "equivalent"
	}

	autogold.Want("equivalent or-expression", "equivalent").Equal(t, test("(repo:bob or repo:jim) ((rev:olga or rev:ham) demo123232)"))
}
