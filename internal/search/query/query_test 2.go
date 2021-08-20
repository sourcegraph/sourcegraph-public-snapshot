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

func TestPipeline_equivalence(t *testing.T) {
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

func TestPipeline(t *testing.T) {
	test := func(input string) string {
		pipelinePlan, _ := Pipeline(InitStructural(input))
		return planToString(Dnf(pipelinePlan.ToParseTree()))
	}

	autogold.Want("contains(...) spans newlines", `"repo:contains.file(\nfoo\n)"`).Equal(t, test("repo:contains.file(\nfoo\n)"))
}
