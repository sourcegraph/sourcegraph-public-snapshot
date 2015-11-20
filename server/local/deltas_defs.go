package local

import (
	"sort"
	"strings"

	"code.google.com/p/rog-go/parallel"
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/srclib/graph"
	srcstore "sourcegraph.com/sourcegraph/srclib/store"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

func (s *deltas) ListDefs(ctx context.Context, op *sourcegraph.DeltasListDefsOp) (*sourcegraph.DeltaDefs, error) {
	ds := op.Ds
	opt := op.Opt

	if opt == nil {
		opt = &sourcegraph.DeltaListDefsOptions{}
	}

	if ds.Base.URI == "" {
		panic("empty Base URI")
	}
	if ds.Base.CommitID == "" {
		panic("empty Base CommitID")
	}
	if ds.Head.URI == "" {
		panic("empty Head URI")
	}
	if ds.Head.CommitID == "" {
		panic("empty Head CommitID")
	}

	nonLocal := srcstore.DefFilterFunc(func(def *graph.Def) bool { return !def.Local })
	baseFilters := []srcstore.DefFilter{srcstore.ByRepos(ds.Base.RepoSpec.URI), srcstore.ByCommitIDs(ds.Base.CommitID), nonLocal}
	headFilters := []srcstore.DefFilter{srcstore.ByRepos(ds.Head.RepoSpec.URI), srcstore.ByCommitIDs(ds.Head.CommitID), nonLocal}

	// Support using different base/head source unit names for Go
	// packages, whose unit name is often prefixed with the repo.
	baseFilters = append(baseFilters, opt.DefFilters()...)
	if strings.HasPrefix(opt.Unit, ds.Base.RepoSpec.URI) && ds.Base.RepoSpec.URI != ds.Head.RepoSpec.URI {
		opt2 := *opt
		opt2.Unit = ds.Head.RepoSpec.URI + strings.TrimPrefix(opt.Unit, ds.Base.RepoSpec.URI)
		headFilters = append(headFilters, opt2.DefFilters()...)
	} else {
		headFilters = append(headFilters, opt.DefFilters()...)
	}

	par := parallel.NewRun(2)
	var baseDefs, headDefs []*graph.Def
	par.Do(func() (err error) {
		baseDefs, err = store.GraphFromContext(ctx).Defs(baseFilters...)
		sort.Sort(graph.Defs(baseDefs))
		return
	})
	par.Do(func() (err error) {
		headDefs, err = store.GraphFromContext(ctx).Defs(headFilters...)
		sort.Sort(graph.Defs(headDefs))
		return
	})
	if err := par.Wait(); err != nil {
		return nil, err
	}

	// HACK ensure there are defs in both base and head. If not, we're
	// going to get erroneously HUGE diffs that will take forever to
	// load.
	if len(baseDefs) == 0 || len(headDefs) == 0 {
		return &sourcegraph.DeltaDefs{}, nil
	}

	deltaDefs := diffDefs(baseDefs, headDefs)
	sort.Sort(deltaDefs)

	// Paginate
	if opt.Page == 0 {
		opt.Page = 1
	}
	if opt.PerPage == 0 {
		opt.PerPage = 100
	}
	lower, upper := (opt.PageOrDefault()-1)*opt.PerPageOrDefault(), opt.PageOrDefault()*opt.PerPageOrDefault()
	if lower > len(deltaDefs.Defs) {
		lower = len(deltaDefs.Defs)
	}
	if upper > len(deltaDefs.Defs) {
		upper = len(deltaDefs.Defs)
	}
	deltaDefs.Defs = deltaDefs.Defs[lower:upper]

	for _, dd := range deltaDefs.Defs {
		if dd.Base != nil {
			populateDefFormatStrings(dd.Base)
		}
		if dd.Head != nil {
			populateDefFormatStrings(dd.Head)
		}
	}

	return deltaDefs, nil
}

type defID struct {
	// UnitType == Def.DefKey.UnitType
	UnitType string

	// Unit == Def.DefKey.Unit *except* in the case when Def.DefKey.Repo is a prefix of Def.DefKey.Unit
	Unit string

	// Path == Def.DefKey.Path
	Path string
}

// makeDefID returns a defID that uniquely identifies a definition across forks and commits. It includes special
// handling for Go definitions (see makeDeltaDefsTestDefns_golangCrossRepo).
func makeDefID(d *graph.Def) defID {
	return defID{UnitType: d.UnitType, Unit: strings.TrimPrefix(d.Unit, d.Repo), Path: d.Path}
}

// defChanged compares 2 versions of the same "logical" def and returns whether or not that def has changed.
// It assumes that the caller has verified that the 2 defs are indeed versions of the same "logical" def.
// HACK: currently assumes the def has changed if and only if its length has changed.
func defChanged(d1, d2 *graph.Def) bool {
	return d1.DefEnd-d1.DefStart != d2.DefEnd-d2.DefStart
}

func diffDefs(base, head []*graph.Def) *sourcegraph.DeltaDefs {
	var delta sourcegraph.DeltaDefs

	baseSet := make(map[defID]*graph.Def)
	for _, baseDef := range base {
		baseSet[makeDefID(baseDef)] = baseDef
	}
	for _, headDef := range head {
		if baseDef, inBase := baseSet[makeDefID(headDef)]; inBase {
			if defChanged(baseDef, headDef) {
				delta.DiffStat.Changed++
				delta.Defs = append(delta.Defs, &sourcegraph.DefDelta{Base: &sourcegraph.Def{Def: *baseDef}, Head: &sourcegraph.Def{Def: *headDef}})
			}
			baseSet[makeDefID(headDef)] = nil
		} else {
			delta.DiffStat.Added++
			delta.Defs = append(delta.Defs, &sourcegraph.DefDelta{Head: &sourcegraph.Def{Def: *headDef}})
		}
	}
	for _, baseDef := range baseSet {
		if baseDef != nil {
			delta.DiffStat.Deleted++
			delta.Defs = append(delta.Defs, &sourcegraph.DefDelta{Base: &sourcegraph.Def{Def: *baseDef}})
		}
	}

	return &delta
}

// baseDefsChangedAndRemoved returns removed and changed defs from a
// DeltaDefs (i.e., it excludes added defs).
func baseDefsChangedAndRemoved(dd *sourcegraph.DeltaDefs) []*sourcegraph.Def {
	var defsChangedRemoved []*sourcegraph.Def
	for _, def := range dd.Defs {
		if def.Base != nil { // in base means it was either removed or changed
			// Use the defs on the base commit because those are more likely
			// to have refs (base is more commonly used than forks, in
			// general).
			defsChangedRemoved = append(defsChangedRemoved, def.Base)
		}
	}
	return defsChangedRemoved
}
