package local

import (
	"reflect"
	"strings"
	"testing"

	"sourcegraph.com/sourcegraph/srclib/unit"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store/mockstore"
)

func TestUnitsService_Get(t *testing.T) {
	var s units
	ctx, mock := testContext()

	wantUnit := &unit.SourceUnit{Type: "t", Name: "u"}
	wantRSUnit := wrapUnits([]*unit.SourceUnit{wantUnit})[0]

	calledUnits := mockstore.GraphMockUnits(&mock.stores.Graph, wantUnit)

	rsUnit, err := s.Get(ctx, &sourcegraph.UnitSpec{
		RepoRevSpec: sourcegraph.RepoRevSpec{RepoSpec: sourcegraph.RepoSpec{URI: "r"}, Rev: "v", CommitID: "c"},
		Unit:        "u",
		UnitType:    "t",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rsUnit, wantRSUnit) {
		t.Errorf("got %+v, want %+v", rsUnit, wantRSUnit)
	}
	if !*calledUnits {
		t.Error("*calledUnits")
	}
}

func TestUnitsService_Get_notFound(t *testing.T) {
	var s units
	ctx, mock := testContext()

	calledUnits := mockstore.GraphMockUnits(&mock.stores.Graph)

	_, err := s.Get(ctx, &sourcegraph.UnitSpec{
		RepoRevSpec: sourcegraph.RepoRevSpec{RepoSpec: sourcegraph.RepoSpec{URI: "r"}, Rev: "v", CommitID: "c"},
		Unit:        "u",
		UnitType:    "DOESNTEXIST",
	})
	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("got error %v, want '...does not exist'", err)
	}
	if !*calledUnits {
		t.Error("*calledUnits")
	}
}

func TestUnitsService_List(t *testing.T) {
	var s units
	ctx, mock := testContext()

	commitID := strings.Repeat("c", 40)

	wantUnits := []*unit.SourceUnit{
		{Repo: "r", CommitID: commitID, Type: "t", Name: "u1"},
		{Repo: "r", CommitID: commitID, Type: "t", Name: "u2"},
	}
	wantRSUnits := wrapUnits(wantUnits)

	calledUnits := mockstore.GraphMockUnits(&mock.stores.Graph, wantUnits...)

	rsUnits, err := s.List(ctx, &sourcegraph.UnitListOptions{
		RepoRevs: []string{"r@" + commitID},
		UnitType: "t",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rsUnits.Units, wantRSUnits) {
		t.Errorf("got %+v, want %+v", rsUnits.Units, wantRSUnits)
	}
	if !*calledUnits {
		t.Error("*calledUnits")
	}
}

func TestUnitsService_List_Pagination(t *testing.T) {
	var s units
	ctx, mock := testContext()

	commitID := strings.Repeat("c", 40)

	inputUnits := []*unit.SourceUnit{
		{Repo: "r", CommitID: commitID, Type: "t", Name: "u1"},
		{Repo: "r", CommitID: commitID, Type: "t", Name: "u2"},
	}
	wantUnits := []*unit.SourceUnit{{Repo: "r", CommitID: commitID, Type: "t", Name: "u1"}}
	wantRSUnits := wrapUnits(wantUnits)

	mockstore.GraphMockUnits(&mock.stores.Graph, inputUnits...)

	rsUnits, err := s.List(ctx, &sourcegraph.UnitListOptions{
		RepoRevs:    []string{"r@" + commitID},
		UnitType:    "t",
		ListOptions: sourcegraph.ListOptions{PerPage: 1, Page: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rsUnits.Units, wantRSUnits) {
		t.Errorf("got %+v, want %+v", rsUnits.Units, wantRSUnits)
	}
}

func wrapUnits(units []*unit.SourceUnit) []*unit.RepoSourceUnit {
	rsUnits := make([]*unit.RepoSourceUnit, len(units))
	for i, u := range units {
		var err error
		rsUnits[i], err = unit.NewRepoSourceUnit(u)
		if err != nil {
			panic(err)
		}
	}
	return rsUnits
}

func unwrapUnits(rsUnits []*unit.RepoSourceUnit) []*unit.SourceUnit {
	units := make([]*unit.SourceUnit, len(rsUnits))
	for i, rsUnit := range rsUnits {
		var err error
		units[i], err = rsUnit.SourceUnit()
		if err != nil {
			panic(err)
		}
	}
	return units
}
