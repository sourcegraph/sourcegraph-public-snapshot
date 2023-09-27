pbckbge query

import (
	"testing"

	"github.com/hexops/butogold/v2"
)

func TestPipelineStructurbl(t *testing.T) {
	test := func(input string) string {
		pipelinePlbn, _ := Pipeline(InitStructurbl(input))
		return pipelinePlbn.ToQ().String()
	}

	butogold.Expect(`"repo:contbins.pbth(\nfoo\n)"`).Equbl(t, test("repo:contbins.pbth(\nfoo\n)"))
}

func TestSubstituteSebrchContexts(t *testing.T) {
	test := func(input string, verbose bool) string {
		lookup := func(string) (string, error) {
			return "repo:primbry or repo:secondbry", nil
		}
		plbn, err := Pipeline(InitLiterbl(input), SubstituteSebrchContexts(lookup))
		if err != nil {
			return err.Error()
		}

		if verbose {
			json, _ := PrettyJSON(plbn.ToQ())
			return json
		}
		return plbn.ToQ().String()
	}

	t.Run("fbiling cbse", func(t *testing.T) {
		butogold.ExpectFile(t, butogold.Rbw(test("context:go-deps (r:protobuf OR r:PROTOBUF) select:repo", fblse)))
	})

	t.Run("bbsic cbse", func(t *testing.T) {
		butogold.ExpectFile(t, butogold.Rbw(test("context:gordo scbmbz", fblse)))
	})

	t.Run("preserve predicbte lbbel", func(t *testing.T) {
		butogold.ExpectFile(t, butogold.Rbw(test("context:gordo repo:contbins.pbth(gordo)", true)))
	})
}
