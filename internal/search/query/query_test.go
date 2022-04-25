package query

import (
	"encoding/json"
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

func TestPipelineStructural(t *testing.T) {
	test := func(input string) string {
		pipelinePlan, _ := Pipeline(InitStructural(input))
		return planToString(Dnf(pipelinePlan.ToParseTree()))
	}

	autogold.Want("contains(...) spans newlines", `"repo:contains.file(\nfoo\n)"`).Equal(t, test("repo:contains.file(\nfoo\n)"))
}

func jsonFormatted(nodes []Node) string {
	var jsons []interface{}
	for _, node := range nodes {
		jsons = append(jsons, toJSON(node))
	}
	json, err := json.MarshalIndent(jsons, "", "  ")
	if err != nil {
		return ""
	}
	return string(json)
}

func TestSubstituteSearchContexts(t *testing.T) {
	test := func(input string, verbose bool) string {
		lookup := func(string) (string, error) {
			return "repo:primary or repo:secondary", nil
		}
		plan, err := Pipeline(InitLiteral(input), SubstituteSearchContexts(lookup))
		if err != nil {
			return err.Error()
		}

		if verbose {
			return jsonFormatted(plan.ToParseTree())
		}
		return plan.ToParseTree().String()
	}

	autogold.Want("basic", `(or (and "repo:primary" "scamaz") (and "repo:secondary" "scamaz"))`).Equal(t, test("context:gordo scamaz", false))

	autogold.Want("preserve predicate label", `[
  {
    "or": [
      {
        "and": [
          {
            "field": "repo",
            "value": "primary",
            "negated": false,
            "labels": [
              "None"
            ]
          },
          {
            "field": "repo",
            "value": "contains.file(gordo)",
            "negated": false,
            "labels": [
              "IsPredicate"
            ]
          }
        ]
      },
      {
        "and": [
          {
            "field": "repo",
            "value": "secondary",
            "negated": false,
            "labels": [
              "None"
            ]
          },
          {
            "field": "repo",
            "value": "contains.file(gordo)",
            "negated": false,
            "labels": [
              "IsPredicate"
            ]
          }
        ]
      }
    ]
  }
]`).
		Equal(t, test("context:gordo repo:contains.file(gordo)", true))
}
