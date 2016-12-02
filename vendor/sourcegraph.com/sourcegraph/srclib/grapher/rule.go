package grapher

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"sourcegraph.com/sourcegraph/makex"
	"sourcegraph.com/sourcegraph/srclib"
	"sourcegraph.com/sourcegraph/srclib/buildstore"
	"sourcegraph.com/sourcegraph/srclib/config"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/plan"
	"sourcegraph.com/sourcegraph/srclib/toolchain"
	"sourcegraph.com/sourcegraph/srclib/unit"
	"sourcegraph.com/sourcegraph/srclib/util"
)

const graphOp = "graph"
const graphAllOp = "graph-all"

func init() {
	plan.RegisterRuleMaker(graphOp, makeGraphRules)
	plan.RegisterRuleMaker(graphAllOp, makeGraphAllRules)
	buildstore.RegisterDataType("graph", &graph.Output{})
}

func makeGraphRules(c *config.Tree, dataDir string, existing []makex.Rule) ([]makex.Rule, error) {
	var rules []makex.Rule
	for _, u := range c.SourceUnits {
		// HACK: ensure backward compatibility with old behavior where
		// we assume we should `graph` if no `graph` op explicitly specified
		if _, hasGraphAll := u.Ops[graphAllOp]; hasGraphAll {
			continue
		}
		toolRef, err := toolchain.ChooseTool(graphOp, u.Type)
		if err != nil {
			return nil, err
		}
		rules = append(rules, &GraphUnitRule{dataDir, u, toolRef})
	}
	return rules, nil
}

func makeGraphAllRules(c *config.Tree, dataDir string, existing []makex.Rule) ([]makex.Rule, error) {
	// Group all graph-all units by type.
	groupedUnits := make(map[string]unit.SourceUnits)
	for _, u := range c.SourceUnits {
		if _, ok := u.Ops[graphAllOp]; !ok {
			continue
		}

		groupedUnits[u.Type] = append(groupedUnits[u.Type], u)
	}

	// Make a GraphMultiUnitsRule for each group of source units
	var rules []makex.Rule
	for unitType, units := range groupedUnits {
		toolRef, err := toolchain.ChooseTool(graphOp, unitType)
		if err != nil {
			return nil, err
		}
		rules = append(rules, &GraphMultiUnitsRule{dataDir, units, unitType, toolRef})
	}
	return rules, nil
}

type GraphUnitRule struct {
	dataDir string
	Unit    *unit.SourceUnit
	Tool    *srclib.ToolRef
}

func (r *GraphUnitRule) Target() string {
	return filepath.ToSlash(filepath.Join(r.dataDir, plan.SourceUnitDataFilename(&graph.Output{}, r.Unit)))
}

func (r *GraphUnitRule) Prereqs() []string {
	ps := []string{filepath.ToSlash(filepath.Join(r.dataDir, plan.SourceUnitDataFilename(unit.SourceUnit{}, r.Unit)))}
	for _, file := range r.Unit.Files {
		if _, err := os.Stat(file); err != nil && os.IsNotExist(err) {
			// skip not-existent files listed in source unit
			continue
		}
		ps = append(ps, file)
	}
	return ps
}

func (r *GraphUnitRule) Recipes() []string {
	if r.Tool == nil {
		return nil
	}
	safeCommand := util.SafeCommandName(srclib.CommandName)
	return []string{
		fmt.Sprintf("%s tool %q %q < $< | %s internal normalize-graph-data --unit-type %q --dir . 1> $@", safeCommand, r.Tool.Toolchain, r.Tool.Subcmd, safeCommand, r.Unit.Type),
	}
}

type GraphMultiUnitsRule struct {
	dataDir   string
	Units     unit.SourceUnits
	UnitsType string
	Tool      *srclib.ToolRef
}

func (r *GraphMultiUnitsRule) Target() string {
	// This is a dummy target, which is only used for ensuring a stable ordering of
	// the makefile rules (see plan/util.go). Both import command and coverage command
	// call the Targets() method to get the *.graph.json filepaths for all units graphed
	// by this rule.
	return filepath.ToSlash(filepath.Join(r.dataDir, plan.SourceUnitDataFilename(&graph.Output{}, &unit.SourceUnit{Key: unit.Key{Type: r.UnitsType}})))
}

func (r *GraphMultiUnitsRule) Targets() map[string]*unit.SourceUnit {
	targets := make(map[string]*unit.SourceUnit)
	for _, u := range r.Units {
		targets[filepath.ToSlash(filepath.Join(r.dataDir, plan.SourceUnitDataFilename(&graph.Output{}, u)))] = u
	}
	return targets
}

func (r *GraphMultiUnitsRule) Prereqs() []string {
	ps := []string{}
	for _, u := range r.Units {
		ps = append(ps, filepath.ToSlash(filepath.Join(r.dataDir, plan.SourceUnitDataFilename(unit.SourceUnit{}, u))))
		for _, file := range u.Files {
			if _, err := os.Stat(file); err != nil && os.IsNotExist(err) {
				// skip not-existent files listed in source unit
				continue
			}
			ps = append(ps, file)
		}
	}
	return ps
}

func (r *GraphMultiUnitsRule) Recipes() []string {
	if r.Tool == nil {
		return nil
	}
	safeCommand := util.SafeCommandName(srclib.CommandName)
	unitFiles := []string{}
	for _, u := range r.Units {
		unitFiles = append(unitFiles, filepath.ToSlash(filepath.Join(r.dataDir, plan.SourceUnitDataFilename(unit.SourceUnit{}, u))))
	}

	// Use `find` command + `xargs` because otherwise the arguments list can become too long.
	var findCmd = "find -L"
	if runtime.GOOS == "windows" {
		findCmd = "/usr/bin/find"
	}
	return []string{
		fmt.Sprintf(`%s %s -name "*%s.unit.json" | xargs %s internal emit-unit-data  | %s tool %q %q | %s internal normalize-graph-data --unit-type %q --dir . --multi --data-dir %s`, findCmd, filepath.ToSlash(r.dataDir), r.UnitsType, safeCommand, safeCommand, r.Tool.Toolchain, r.Tool.Subcmd, safeCommand, r.UnitsType, filepath.ToSlash(r.dataDir)),
	}
}
