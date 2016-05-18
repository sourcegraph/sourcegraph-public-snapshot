package dep

import (
	"fmt"
	"path/filepath"

	"sourcegraph.com/sourcegraph/makex"
	"sourcegraph.com/sourcegraph/srclib"
	"sourcegraph.com/sourcegraph/srclib/buildstore"
	"sourcegraph.com/sourcegraph/srclib/config"
	"sourcegraph.com/sourcegraph/srclib/plan"
	"sourcegraph.com/sourcegraph/srclib/toolchain"
	"sourcegraph.com/sourcegraph/srclib/unit"
	"sourcegraph.com/sourcegraph/srclib/util"
)

const depresolveOp = "depresolve"

func init() {
	plan.RegisterRuleMaker(depresolveOp, makeDepRules)
	buildstore.RegisterDataType("depresolve", []*ResolvedDep{})
}

func makeDepRules(c *config.Tree, dataDir string, existing []makex.Rule) ([]makex.Rule, error) {
	const op = depresolveOp
	var rules []makex.Rule
	for _, u := range c.SourceUnits {
		toolRef, err := toolchain.ChooseTool(depresolveOp, u.Type)
		if err != nil {
			return nil, err
		}

		rules = append(rules, &ResolveDepsRule{dataDir, u, toolRef})
	}
	return rules, nil
}

type ResolveDepsRule struct {
	dataDir string
	Unit    *unit.SourceUnit
	Tool    *srclib.ToolRef
}

func (r *ResolveDepsRule) Target() string {
	return filepath.ToSlash(filepath.Join(r.dataDir, plan.SourceUnitDataFilename([]*ResolvedDep{}, r.Unit)))
}

func (r *ResolveDepsRule) Prereqs() []string {
	return []string{filepath.ToSlash(filepath.Join(r.dataDir, plan.SourceUnitDataFilename(unit.SourceUnit{}, r.Unit)))}
}

func (r *ResolveDepsRule) Recipes() []string {
	if r.Tool == nil {
		return nil
	}
	return []string{
		fmt.Sprintf("%s tool %q %q < $^ 1> $@", util.SafeCommandName(srclib.CommandName), r.Tool.Toolchain, r.Tool.Subcmd),
	}
}
