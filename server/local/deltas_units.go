package local

import (
	"reflect"
	"sort"
	"strings"

	"github.com/rogpeppe/rog-go/parallel"
	"golang.org/x/net/context"
	srcstore "sourcegraph.com/sourcegraph/srclib/store"
	"sourcegraph.com/sourcegraph/srclib/unit"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

func (s *deltas) ListUnits(ctx context.Context, op *sourcegraph.DeltasListUnitsOp) (*sourcegraph.UnitDeltaList, error) {
	ds := op.Ds
	opt := op.Opt

	if opt == nil {
		opt = &sourcegraph.DeltaListUnitsOptions{}
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

	baseFilters := []srcstore.UnitFilter{srcstore.ByRepos(ds.Base.RepoSpec.URI), srcstore.ByCommitIDs(ds.Base.CommitID)}
	headFilters := []srcstore.UnitFilter{srcstore.ByRepos(ds.Head.RepoSpec.URI), srcstore.ByCommitIDs(ds.Head.CommitID)}

	par := parallel.NewRun(2)
	var baseUnits, headUnits []*unit.SourceUnit
	par.Do(func() (err error) {
		baseUnits, err = store.GraphFromContext(ctx).Units(baseFilters...)
		sort.Sort(unit.SourceUnits(baseUnits))
		return
	})
	par.Do(func() (err error) {
		headUnits, err = store.GraphFromContext(ctx).Units(headFilters...)
		sort.Sort(unit.SourceUnits(headUnits))
		return
	})
	if err := par.Wait(); err != nil {
		return nil, err
	}

	// HACK ensure there are units in both base and head. If not, we're
	// going to get erroneously HUGE diffs that will take forever to
	// load.
	if len(baseUnits) == 0 || len(headUnits) == 0 {
		return &sourcegraph.UnitDeltaList{}, nil
	}

	// Get the list of VCS-modified files (which is another way source
	// units can be modified).
	//
	// TODO(sqs): make it possible to only get the diff stat (i.e.,
	// list of changed files) from vcsstore, for better perf.
	fdiff, _, err := s.diff(ctx, ds)
	if err != nil {
		return nil, err
	}
	var baseModFiles, headModFiles []string
	for _, f := range fdiff {
		baseModFiles = append(baseModFiles, f.OrigName)
		headModFiles = append(headModFiles, f.NewName)
	}

	deltaUnits, err := diffUnits(baseUnits, headUnits, baseModFiles, headModFiles)
	if err != nil {
		return nil, err
	}
	sort.Sort(sourcegraph.UnitDeltas(deltaUnits))

	return &sourcegraph.UnitDeltaList{UnitDeltas: deltaUnits}, nil
}

// unitDeltaID is a unit.ID2 with the source unit names trimmed of
// repo prefixes (to make them not dependent on the repo URI, in the
// case of, e.g., Go package import paths).
type unitDeltaID struct{ unit.ID2 }

// makeUnitDeltaID returns a unitDeltaID that uniquely identifies a
// source unit across forks and commits. It includes special handling
// for Go packages (see makeDeltaUnitsTestData_golangCrossRepo).
func makeUnitDeltaID(u *unit.SourceUnit) unitDeltaID {
	files := u.Files
	sort.Strings(files)
	id2 := u.ID2()
	id2.Name = strings.TrimPrefix(id2.Name, u.Repo)
	return unitDeltaID{id2}
}

func unitChanged(base, head *unit.SourceUnit, baseModFiles, headModFiles map[string]struct{}) bool {
	sort.Strings(base.Files)
	sort.Strings(head.Files)
	for _, f := range base.Files {
		if _, mod := baseModFiles[f]; mod {
			return true
		}
	}
	for _, f := range head.Files {
		if _, mod := headModFiles[f]; mod {
			return true
		}
	}
	return !reflect.DeepEqual(base.Files, head.Files)
}

// diffUnits returns a list of UnitDeltas representing the changes
// between base and head's list of source units. The baseModFiles and
// headModFiles are lists of files that were added/changed/removed in
// the base and head VCS commits, respectively (baseModFiles uses
// filenames in the base commit, and headModFiles uses filenames from
// the head commit).
func diffUnits(base, head []*unit.SourceUnit, baseModFiles, headModFiles []string) ([]*sourcegraph.UnitDelta, error) {
	var du []*sourcegraph.UnitDelta

	makeFileMap := func(files []string) map[string]struct{} {
		m := make(map[string]struct{}, len(files))
		for _, f := range files {
			m[f] = struct{}{}
		}
		return m
	}
	baseFileMap, headFileMap := makeFileMap(baseModFiles), makeFileMap(headModFiles)

	baseSet := make(map[unitDeltaID]*unit.SourceUnit)
	for _, baseUnit := range base {
		baseSet[makeUnitDeltaID(baseUnit)] = baseUnit
	}
	for _, headUnit := range head {
		if baseUnit, inBase := baseSet[makeUnitDeltaID(headUnit)]; inBase {
			if unitChanged(baseUnit, headUnit, baseFileMap, headFileMap) {
				udelt, err := newUnitDelta(baseUnit, headUnit)
				if err != nil {
					return nil, err
				}
				du = append(du, udelt)
			}
			baseSet[makeUnitDeltaID(headUnit)] = nil
		} else {
			udelt, err := newUnitDelta(nil, headUnit)
			if err != nil {
				return nil, err
			}
			du = append(du, udelt)
		}
	}
	for _, baseUnit := range baseSet {
		if baseUnit != nil {
			udelt, err := newUnitDelta(baseUnit, nil)
			if err != nil {
				return nil, err
			}
			du = append(du, udelt)
		}
	}

	return du, nil
}

func newUnitDelta(base, head *unit.SourceUnit) (*sourcegraph.UnitDelta, error) {
	var ud sourcegraph.UnitDelta
	var err error
	if base != nil {
		ud.Base, err = unit.NewRepoSourceUnit(base)
		if err != nil {
			return nil, err
		}
	}
	if head != nil {
		ud.Head, err = unit.NewRepoSourceUnit(head)
		if err != nil {
			return nil, err
		}
	}
	return &ud, nil
}
