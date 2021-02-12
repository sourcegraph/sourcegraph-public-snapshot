package search

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

func TestRepoStatusMap(t *testing.T) {
	aM := map[api.RepoID]RepoStatus{
		1: RepoStatusSearched,
		2: RepoStatusCloning,
		3: RepoStatusSearched | RepoStatusIndexed,
		4: RepoStatusIndexed,
	}
	a := mkStatusMap(aM)
	b := mkStatusMap(map[api.RepoID]RepoStatus{
		2: RepoStatusCloning,
		4: RepoStatusSearched,
		5: RepoStatusMissing,
	})
	c := mkStatusMap(map[api.RepoID]RepoStatus{
		8: RepoStatusSearched | RepoStatusIndexed,
		9: RepoStatusSearched,
	})

	// Get
	if got, want := a.Get(10), RepoStatus(0); got != want {
		t.Errorf("a.Get(10) got %s want %s", got, want)
	}
	if got, want := a.Get(3), RepoStatusSearched|RepoStatusIndexed; got != want {
		t.Errorf("a.Get(3) got %s want %s", got, want)
	}

	// Any
	if !c.Any(RepoStatusIndexed) {
		t.Error("c.Any(RepoStatusIndexed) should be true")
	}
	if c.Any(RepoStatusCloning) {
		t.Error("c.Any(RepoStatusCloning) should be false")
	}

	// All
	if !c.All(RepoStatusSearched) {
		t.Error("c.All(RepoStatusSearched) should be true")
	}
	if c.All(RepoStatusIndexed) {
		t.Error("c.All(RepoStatusIndexed) should be false")
	}

	// Len
	if got, want := c.Len(), 2; got != want {
		t.Errorf("c.Len got %d want %d", got, want)
	}

	// Update
	c.Update(9, RepoStatusIndexed)
	if got, want := c.Get(9), RepoStatusSearched|RepoStatusIndexed; got != want {
		t.Errorf("c.Get(9) got %s want %s", got, want)
	}

	// Update with add
	c.Update(123, RepoStatusCloning)
	if got, want := c.Get(123), RepoStatusCloning; got != want {
		t.Errorf("c.Get(123) got %s want %s", got, want)
	}
	if got, want := c.Len(), 3; got != want {
		t.Errorf("c.Len after add got %d want %d", got, want)
	}

	// Iterate
	gotIterate := map[api.RepoID]RepoStatus{}
	a.Iterate(func(id api.RepoID, s RepoStatus) {
		gotIterate[id] = s
	})
	if d := cmp.Diff(aM, gotIterate); d != "" {
		t.Errorf("a.Iterate diff (-want, +got):\n%s", d)
	}

	// Filter
	assertAFilter := func(status RepoStatus, want []int) {
		t.Helper()
		var got []int
		a.Filter(status, func(id api.RepoID) {
			got = append(got, int(id))
		})
		sort.Ints(got)
		if d := cmp.Diff(want, got, cmpopts.EquateEmpty()); d != "" {
			t.Errorf("a.Filter(%s) diff (-want, +got):\n%s", status, d)
		}
	}
	assertAFilter(RepoStatusSearched, []int{1, 3})
	assertAFilter(RepoStatusTimedout, []int{})

	// Union
	t.Logf("%s", &a)
	t.Logf("%s", &b)
	b.Union(&a)
	t.Logf("%s", &b)
	abUnionWant := mkStatusMap(map[api.RepoID]RepoStatus{
		1: RepoStatusSearched,
		2: RepoStatusCloning,
		3: RepoStatusSearched | RepoStatusIndexed,
		4: RepoStatusSearched | RepoStatusIndexed,
		5: RepoStatusMissing,
	})
	assertReposStatusEqual(t, abUnionWant, b)

	// Union on uninitialized LHS
	var empty RepoStatusMap
	empty.Union(&a)
	assertReposStatusEqual(t, a, empty)
}

// Test we have reasonable behaviour on nil maps
func TestRepoStatusMap_nil(t *testing.T) {
	var x *RepoStatusMap
	t.Logf("%s", x)
	x.Iterate(func(api.RepoID, RepoStatus) {
		t.Error("Iterate should be empty")
	})
	x.Filter(RepoStatusSearched, func(api.RepoID) {
		t.Error("Filter should be empty")
	})
	if got, want := x.Get(10), RepoStatus(0); got != want {
		t.Errorf("Get got %s want %s", got, want)
	}
	if x.Any(RepoStatusSearched) {
		t.Error("Any should be false")
	}
	if x.All(RepoStatusSearched) {
		t.Error("All should be false")
	}
	if got, want := x.Len(), 0; got != want {
		t.Errorf("Len got %d want %d", got, want)
	}
}

func TestRepoStatusSingleton(t *testing.T) {
	x := RepoStatusSingleton(123, RepoStatusSearched|RepoStatusIndexed)
	want := mkStatusMap(map[api.RepoID]RepoStatus{
		123: RepoStatusSearched | RepoStatusIndexed,
	})
	assertReposStatusEqual(t, want, x)
}

func mkStatusMap(m map[api.RepoID]RepoStatus) RepoStatusMap {
	var rsm RepoStatusMap
	for id, status := range m {
		rsm.Update(id, status)
	}
	return rsm
}

func assertReposStatusEqual(t *testing.T, want, got RepoStatusMap) {
	t.Helper()

	wantm := map[api.RepoID]RepoStatus{}
	gotm := map[api.RepoID]RepoStatus{}

	want.Iterate(func(id api.RepoID, mask RepoStatus) {
		wantm[id] = mask
	})
	got.Iterate(func(id api.RepoID, mask RepoStatus) {
		gotm[id] = mask
	})
	if diff := cmp.Diff(wantm, gotm); diff != "" {
		t.Errorf("RepoStatusMap mismatch (-want +got):\n%s", diff)
	}
}
