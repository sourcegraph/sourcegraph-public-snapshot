package query

import (
	"testing"

	"github.com/hexops/autogold/v2"
)

func TestPipelineStructural(t *testing.T) {
	test := func(input string) string {
		pipelinePlan, _ := Pipeline(InitStructural(input))
		return pipelinePlan.ToQ().String()
	}

	autogold.Expect(`"repo:contains.path(\nfoo\n)"`).Equal(t, test("repo:contains.path(\nfoo\n)"))
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
			json, _ := PrettyJSON(plan.ToQ())
			return json
		}
		return plan.ToQ().String()
	}

	t.Run("failing case", func(t *testing.T) {
		autogold.ExpectFile(t, autogold.Raw(test("context:go-deps (r:protobuf OR r:PROTOBUF) select:repo", false)))
	})

	t.Run("basic case", func(t *testing.T) {
		autogold.ExpectFile(t, autogold.Raw(test("context:gordo scamaz", false)))
	})

	t.Run("preserve predicate label", func(t *testing.T) {
		autogold.ExpectFile(t, autogold.Raw(test("context:gordo repo:contains.path(gordo)", true)))
	})
}
