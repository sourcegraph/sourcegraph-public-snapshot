package query

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hexops/autogold"
)

func planToString(disjuncts [][]Node) string {
	var plan Plan
	for _, disjunct := range disjuncts {
		parameters, pattern, _ := PartitionSearchPattern(disjunct)
		plan = append(plan, Basic{Parameters: parameters, Pattern: pattern})
	}

	var result []string
	for _, basic := range plan {
		result = append(result, toString(basic.ToParseTree()))
	}
	return strings.Join(result, " ")
}

func TestPipeline(t *testing.T) {
	// The Pipeline must produce a value that is equivalent under DNF to parsing the query and processing it with DNF.
	test := func(input string) string {
		pipelinePlan, _ := Pipeline(InitLiteral(input))
		nodes, _ := Run(InitLiteral(input))
		disjuncts := Dnf(nodes)
		plan, _ := ToPlan(disjuncts)
		manualPlan := MapPlan(plan, ConcatRevFilters)
		if diff := cmp.Diff(
			planToString(Dnf(pipelinePlan.ToParseTree())),
			planToString(Dnf(manualPlan.ToParseTree())),
		); diff != "" {
			return diff
		}
		return "equivalent"
	}

	autogold.Want("equivalent or-expression", "equivalent").Equal(t, test("(repo:bob or repo:jim) ((rev:olga or rev:ham) demo123232)"))
}

func TestTypeRepoToFilter(t *testing.T) {
	test := func(input string, searchType SearchType) string {
		pipelinePlan, err := Pipeline(Init(input, searchType))
		if err != nil {
			return err.Error()
		}
		return planToString(Dnf(pipelinePlan.ToParseTree()))
	}

	autogold.Want("01", `"repo:typescript" "repo:sourcegraph" "repo:derp"`).Equal(t, test("type:repo typescript sourcegraph derp", SearchTypeLiteral))

	autogold.Want("02", `"repo:source\\.\\*graph"`).Equal(t, test("type:repo source.*graph", SearchTypeLiteral))

	autogold.Want("03", `"repo:source.*graph"`).Equal(t, test("type:repo source.*graph", SearchTypeRegex))

	autogold.Want("04", "this query expression is not compatible with `type:repo`. You might try adding an `and` keyword before or after any parenthesized expressions.").Equal(t, test("type:repo (e OR f) g h i", SearchTypeLiteral))
}
