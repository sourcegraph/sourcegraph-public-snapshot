package local

import (
	"errors"
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"strings"

	"sourcegraph.com/sourcegraph/go-diff/diff"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	srcstore "sourcegraph.com/sourcegraph/srclib/store"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

// Test that DeltasService.Get returns partial info even if a call
// fails (e.g., it returns the base repo even if the head repo doesn't
// exist).
func TestDeltasService_Get_returnsPartialInfo(t *testing.T) {
	var s deltas
	ctx, mock := testContext()

	wantErr := errors.New("foo")

	var calledGet int
	mock.servers.Repos.MockGetCommit_ByID_NoCheck(t, "c")
	mock.servers.Builds.MockGetRepoBuildInfo(t, &sourcegraph.RepoBuildInfo{})
	mock.servers.Repos.Get_ = func(ctx context.Context, repoSpec *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
		calledGet++
		if calledGet == 2 {
			return nil, wantErr
		}
		return &sourcegraph.Repo{URI: "x", DefaultBranch: "b"}, nil
	}

	delta, err := s.Get(ctx, &sourcegraph.DeltaSpec{})
	if err != wantErr {
		t.Errorf("got error %v, want %v", err, wantErr)
	}
	if delta == nil || delta.BaseRepo == nil {
		t.Errorf("delta.BaseRepo==nil, want non-nil (partial result despite error)")
	}
	if want := 2; calledGet != want {
		t.Errorf("called get %d times, want %d times", calledGet, want)
	}
}

func TestDeltasService_ListUnits(t *testing.T) {
	tests := []struct {
		description string
		baseRevSpec sourcegraph.RepoRevSpec
		headRevSpec sourcegraph.RepoRevSpec
	}{{
		description: "same repo",
		baseRevSpec: sourcegraph.RepoRevSpec{
			RepoSpec: sourcegraph.RepoSpec{URI: "github.com/owner/foo"},
			Rev:      "baserev", CommitID: "basecommit",
		},
		headRevSpec: sourcegraph.RepoRevSpec{
			RepoSpec: sourcegraph.RepoSpec{URI: "github.com/owner/foo"},
			Rev:      "headrev", CommitID: "headcommit",
		},
	}, {
		description: "different repo",
		baseRevSpec: sourcegraph.RepoRevSpec{
			RepoSpec: sourcegraph.RepoSpec{URI: "github.com/owner/foo"},
			Rev:      "baserev", CommitID: "basecommit",
		},
		headRevSpec: sourcegraph.RepoRevSpec{
			RepoSpec: sourcegraph.RepoSpec{URI: "github.com/owner2/foob"},
			Rev:      "headrev", CommitID: "headcommit",
		},
	}}

	for _, test := range tests {
		func() {
			var s deltas
			ctx, mock := testContext()

			ds := sourcegraph.DeltaSpec{
				Base: test.baseRevSpec,
				Head: test.headRevSpec,
			}
			data := makeDeltaUnitsTestData(ds)
			mock.stores.Graph = newMockMultiRepoStoreWithUnits(data)

			s.mockDiffFunc = func(ctx context.Context, ds sourcegraph.DeltaSpec) ([]*diff.FileDiff, *sourcegraph.Delta, error) {
				return []*diff.FileDiff{
					{OrigName: "c", NewName: "c"},
				}, nil, nil
			}

			du, err := s.ListUnits(ctx, &sourcegraph.DeltasListUnitsOp{Ds: ds})
			if err != nil {
				t.Fatal(err)
			}

			checkDeltaUnits(t, du, deltaCheck{Added: true, Changed: true, VCSChanged: true, Deleted: true})
		}()
	}
}

func TestDeltasService_ListUnits_golangCrossRepo(t *testing.T) {
	baseRevSpec := sourcegraph.RepoRevSpec{
		RepoSpec: sourcegraph.RepoSpec{URI: "github.com/baseowner/foo"},
		Rev:      "baserev", CommitID: "basecommit",
	}
	headRevSpec := sourcegraph.RepoRevSpec{
		RepoSpec: sourcegraph.RepoSpec{URI: "github.com/headowner/foo"},
		Rev:      "headrev", CommitID: "headcommit",
	}

	var s deltas
	ctx, mock := testContext()

	ds := sourcegraph.DeltaSpec{
		Base: baseRevSpec,
		Head: headRevSpec,
	}
	defs := makeDeltaUnitsTestData_golangCrossRepo(ds)
	mock.stores.Graph = newMockMultiRepoStoreWithUnits(defs)

	s.mockDiffFunc = func(ctx context.Context, ds sourcegraph.DeltaSpec) ([]*diff.FileDiff, *sourcegraph.Delta, error) {
		return nil, nil, nil
	}

	du, err := s.ListUnits(ctx, &sourcegraph.DeltasListUnitsOp{Ds: ds})
	if err != nil {
		t.Fatal(err)
	}

	checkDeltaUnits(t, du, deltaCheck{Added: true, Deleted: true})
}

func newMockMultiRepoStoreWithUnits(units []*unit.SourceUnit) srcstore.MockMultiRepoStore {
	mockGraph := srcstore.MockMultiRepoStore{}
	mockGraph.Units_ = func(filters ...srcstore.UnitFilter) ([]*unit.SourceUnit, error) {
		var selectedUnits []*unit.SourceUnit
		for _, unit := range units {
			selected := true
			for _, filter := range filters {
				selected = selected && filter.SelectUnit(unit)
			}
			if selected {
				selectedUnits = append(selectedUnits, unit)
			}
		}
		return selectedUnits, nil
	}
	return mockGraph
}

func makeDeltaUnitsTestData_golangCrossRepo(ds sourcegraph.DeltaSpec) []*unit.SourceUnit {
	// Cross-repo comparisons for Go require special handling because
	// the non-vendored pkg unit names (Go import paths) of Go units
	// in 2 repos are always different, all units from head are shown
	// as added and all units from base are shown as deleted.
	//
	// But we also need to make sure it handles vendored pkg names
	// correctly. For example, packages in a repo's Godeps/_workspace
	// GOPATH do NOT have their import paths (unit names) prefixed
	// with the current repo.
	//
	// This test tests that package names are only considered different if
	// they differ after the repository URI.

	// The following test data has some packages that are the "same"
	// logically but just have different Go import paths.
	return []*unit.SourceUnit{
		&unit.SourceUnit{
			Repo:     ds.Base.RepoSpec.URI,
			CommitID: ds.Base.CommitID,
			Type:     "GoPackage",
			Name:     ds.Base.RepoSpec.URI + "/unchanged-unit",
		},
		&unit.SourceUnit{
			Repo:     ds.Head.RepoSpec.URI,
			CommitID: ds.Head.CommitID,
			Type:     "GoPackage",
			Name:     ds.Head.RepoSpec.URI + "/unchanged-unit",
		},
		&unit.SourceUnit{
			Repo:     ds.Base.RepoSpec.URI,
			CommitID: ds.Base.CommitID,
			Type:     "GoPackage",
			Name:     ds.Base.RepoSpec.URI + "/deleted-unit",
		},
		&unit.SourceUnit{
			Repo:     ds.Head.RepoSpec.URI,
			CommitID: ds.Head.CommitID,
			Type:     "GoPackage",
			Name:     ds.Head.RepoSpec.URI + "/added-unit",
		},

		// units from vendored packages
		&unit.SourceUnit{
			Repo:     ds.Base.RepoSpec.URI,
			CommitID: ds.Base.CommitID,
			Type:     "GoPackage",
			Name:     "github.com/myvendored/pkg/vendored-unit",
		},
		&unit.SourceUnit{
			Repo:     ds.Head.RepoSpec.URI,
			CommitID: ds.Head.CommitID,
			Type:     "GoPackage",
			Name:     "github.com/myvendored/pkg/vendored-unit",
		},
	}
}

// makeDeltaUnitsTestData returns test data that includes added, changed, unchanged, and deleted units for the given
// DeltaSpec. Should be used in conjunction with checkDeltaUnits.
func makeDeltaUnitsTestData(ds sourcegraph.DeltaSpec) []*unit.SourceUnit {
	return []*unit.SourceUnit{
		&unit.SourceUnit{
			Repo:     ds.Base.RepoSpec.URI,
			CommitID: ds.Base.CommitID,
			Type:     "t",
			Name:     "unchanged-unit",
		},
		&unit.SourceUnit{
			Repo:     ds.Head.RepoSpec.URI,
			CommitID: ds.Head.CommitID,
			Type:     "t",
			Name:     "unchanged-unit",
		},
		&unit.SourceUnit{
			Repo:     ds.Base.RepoSpec.URI,
			CommitID: ds.Base.CommitID,
			Type:     "t",
			Name:     "changed-unit",
			Files:    []string{"a"},
		},
		&unit.SourceUnit{
			Repo:     ds.Head.RepoSpec.URI,
			CommitID: ds.Head.CommitID,
			Type:     "t",
			Name:     "changed-unit",
			Files:    []string{"a", "b"}, // added file from base repo
		},
		&unit.SourceUnit{
			Repo:     ds.Base.RepoSpec.URI,
			CommitID: ds.Base.CommitID,
			Type:     "t",
			Name:     "vcs-changed-unit",
			Files:    []string{"c"},
		},
		&unit.SourceUnit{
			Repo:     ds.Head.RepoSpec.URI,
			CommitID: ds.Head.CommitID,
			Type:     "t",
			Name:     "vcs-changed-unit",
			Files:    []string{"c"}, // file "c" was changed in the VCS
		},
		&unit.SourceUnit{
			Repo:     ds.Base.RepoSpec.URI,
			CommitID: ds.Base.CommitID,
			Type:     "t",
			Name:     "deleted-unit",
		},
		&unit.SourceUnit{
			Repo:     ds.Head.RepoSpec.URI,
			CommitID: ds.Head.CommitID,
			Type:     "t",
			Name:     "added-unit",
		},
	}
}

type deltaCheck struct{ Added, Changed, VCSChanged, Deleted bool }

// checkDeltaUnits checks that du's added units all have the path
// "added-unit", du's changed units all have the path "changed-unit",
// and du's deleted units all have the path "deleted-unit", and also
// that its diff stat matches wantDiffStat.
func checkDeltaUnits(t *testing.T, dus *sourcegraph.UnitDeltaList, want deltaCheck) {
	// Ensure that the only added and deleted units are added-unit and
	// deleted-unit, respectively. The unchanged-unit should not
	// appear in either the added or deleted list.
	var saw deltaCheck
	for _, du := range dus.UnitDeltas {
		if du.Added() {
			saw.Added = true
			if want := "added-unit"; !strings.HasSuffix(du.Head.Unit, want) {
				t.Errorf("got added unit %v, want only unit with Name suffix %q", du.Head.Unit, want)
			}
		}
		if du.Changed() {
			saw.Changed = true
			if want := "changed-unit"; !strings.HasSuffix(du.Head.Unit, want) {
				t.Errorf("got changed unit %v, want only unit with Name suffix %q", du.Head.Unit, want)
			}
			if strings.HasSuffix(du.Head.Unit, "vcs-changed-unit") {
				saw.VCSChanged = true
			}
		}
		if du.Deleted() {
			saw.Deleted = true
			if want := "deleted-unit"; !strings.HasSuffix(du.Base.Unit, want) {
				t.Errorf("got deleted unit %v, want only unit with Name suffix %q", du.Base.Unit, want)
			}
		}
	}
	if !reflect.DeepEqual(saw, want) {
		t.Errorf("saw %+v, want all true", saw)
	}
}
