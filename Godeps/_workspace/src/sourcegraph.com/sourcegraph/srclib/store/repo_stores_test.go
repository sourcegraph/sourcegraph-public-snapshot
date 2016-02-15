package store

import (
	"fmt"
	"reflect"
	"testing"

	"sort"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

// mockNeverCalledRepoStore calls t.Error if any of its methods are
// called.
func mockNeverCalledRepoStore(t *testing.T) RepoStore {
	return MockRepoStore{
		Versions_: func(f ...VersionFilter) ([]*Version, error) {
			t.Fatalf("(RepoStore).Versions called, but wanted it not to be called (arg f was %v)", f)
			return nil, nil
		},
		MockTreeStore: mockNeverCalledTreeStore(t),
	}
}

type emptyRepoStore struct{ emptyTreeStore }

func (m emptyRepoStore) Versions(f ...VersionFilter) ([]*Version, error) {
	return []*Version{}, nil
}

type mapRepoStoreOpener map[string]RepoStore

func (m mapRepoStoreOpener) openRepoStore(repo string) RepoStore {
	return m[repo]
}
func (m mapRepoStoreOpener) openAllRepoStores() (map[string]RepoStore, error) { return m, nil }

type recordingRepoStoreOpener struct {
	opened    map[string]int // how many times openRepoStore was called for each repo
	openedAll int            // how many times openAllRepoStores was called
	repoStoreOpener
}

func (m *recordingRepoStoreOpener) openRepoStore(repo string) RepoStore {
	if m.opened == nil {
		m.opened = map[string]int{}
	}
	m.opened[repo]++
	return m.repoStoreOpener.openRepoStore(repo)
}
func (m *recordingRepoStoreOpener) openAllRepoStores() (map[string]RepoStore, error) {
	m.openedAll++
	return m.repoStoreOpener.openAllRepoStores()
}
func (m *recordingRepoStoreOpener) reset() { m.opened = map[string]int{}; m.openedAll = 0 }

func TestRepoStores_filterByRepos(t *testing.T) {
	// Test that filters by a specific repo cause repo stores for
	// other repos to not be called.

	o := &recordingRepoStoreOpener{repoStoreOpener: mapRepoStoreOpener{
		"r":  emptyRepoStore{},
		"r2": mockNeverCalledRepoStore(t),
	}}
	rss := repoStores{opener: o}

	if defs, err := rss.Defs(ByDefKey(graph.DefKey{Repo: "r", CommitID: "c", UnitType: "t", Unit: "u", Path: "p"})); err != nil {
		t.Fatal(err)
	} else if len(defs) != 0 {
		t.Errorf("got defs %v, want none", defs)
	}
	if want := map[string]int{"r": 1}; !reflect.DeepEqual(o.opened, want) {
		t.Errorf("got opened %v, want %v", o.opened, want)
	}
	o.reset()

	if defs, err := rss.Defs(ByRepos("r")); err != nil {
		t.Error(err)
	} else if len(defs) > 0 {
		t.Errorf("got defs %v, want none", defs)
	}

	if refs, err := rss.Refs(ByRepos("r")); err != nil {
		t.Error(err)
	} else if len(refs) > 0 {
		t.Errorf("got refs %v, want none", refs)
	}

	if units, err := rss.Units(ByUnitKey(unit.Key{Repo: "r", CommitID: "c", UnitType: "t", Unit: "u"})); err != nil {
		t.Fatal(err)
	} else if len(units) != 0 {
		t.Errorf("got units %v, want none", units)
	}

	if units, err := rss.Units(ByRepos("r")); err != nil {
		t.Error(err)
	} else if len(units) > 0 {
		t.Errorf("got units %v, want none", units)
	}

	if versions, err := rss.Versions(ByRepos("r"), ByCommitIDs("c")); err != nil {
		t.Fatal(err)
	} else if len(versions) != 0 {
		t.Errorf("got versions %v, want none", versions)
	}

	if versions, err := rss.Versions(ByRepos("r")); err != nil {
		t.Error(err)
	} else if len(versions) > 0 {
		t.Errorf("got versions %v, want none", versions)
	}
}

func TestScopeRepos(t *testing.T) {
	tests := []struct {
		filters []interface{}
		want    []string
	}{
		{
			filters: nil,
			want:    nil,
		},
		{
			filters: []interface{}{ByRepos("r")},
			want:    []string{"r"},
		},
		{
			filters: []interface{}{nil, ByRepos("r")},
			want:    []string{"r"},
		},
		{
			filters: []interface{}{ByRepos("r"), nil},
			want:    []string{"r"},
		},
		{
			filters: []interface{}{nil, ByRepos("r"), nil},
			want:    []string{"r"},
		},
		{
			filters: []interface{}{ByRepos("r"), ByRepos("r")},
			want:    []string{"r"},
		},
		{
			filters: []interface{}{ByRepos("r1"), ByRepos("r2")},
			want:    []string{},
		},
		{
			filters: []interface{}{ByRepos("r1", "r2"), ByRepos("r2", "r1")},
			want:    []string{"r1", "r2"},
		},
		{
			filters: []interface{}{ByRepos("r1", "r2"), ByRepos("r2")},
			want:    []string{"r2"},
		},
		{
			filters: []interface{}{ByRepos("r1"), ByRepos("r2"), ByRepos("r1")},
			want:    []string{},
		},
		{
			filters: []interface{}{ByUnitKey(unit.Key{Repo: "r", CommitID: "c", UnitType: "t", Unit: "u"})},
			want:    []string{"r"},
		},
		{
			filters: []interface{}{
				ByUnitKey(unit.Key{Repo: "r", CommitID: "c", UnitType: "t", Unit: "u"}),
				ByUnitKey(unit.Key{Repo: "r", CommitID: "c2", UnitType: "t2", Unit: "u2"}),
			},
			want: []string{"r"},
		},
		{
			filters: []interface{}{
				ByUnitKey(unit.Key{Repo: "r", CommitID: "c", UnitType: "t", Unit: "u"}),
				ByUnitKey(unit.Key{Repo: "r2", CommitID: "c", UnitType: "t", Unit: "u"}),
			},
			want: []string{},
		},
		{
			filters: []interface{}{ByDefKey(graph.DefKey{Repo: "r", CommitID: "c", UnitType: "t", Unit: "u", Path: "p"})},
			want:    []string{"r"},
		},
		{
			filters: []interface{}{
				ByDefKey(graph.DefKey{Repo: "r", CommitID: "c", UnitType: "t", Unit: "u", Path: "p"}),
				ByDefKey(graph.DefKey{Repo: "r", CommitID: "c2", UnitType: "t2", Unit: "u2", Path: "p2"}),
			},
			want: []string{"r"},
		},
		{
			filters: []interface{}{
				ByDefKey(graph.DefKey{Repo: "r", CommitID: "c", UnitType: "t", Unit: "u", Path: "p"}),
				ByDefKey(graph.DefKey{Repo: "r2", CommitID: "c", UnitType: "t", Unit: "u", Path: "p"}),
			},
			want: []string{},
		},
		{
			filters: []interface{}{RepoFilterFunc(func(string) bool { return false })},
			want:    nil,
		},
		{
			filters: []interface{}{ByUnits(unit.ID2{Type: "t", Name: "u"})},
			want:    nil,
		},
	}
	for _, test := range tests {
		repos, err := scopeRepos(test.filters)
		if err != nil {
			t.Errorf("%+v: %v", test.filters, err)
			continue
		}
		sort.Strings(repos)
		sort.Strings(test.want)
		if !reflect.DeepEqual(repos, test.want) {
			t.Errorf("%+v: got repos %v, want %v", test.filters, repos, test.want)
		}
	}
}

func TestFiltersForRepo(t *testing.T) {
	tests := []struct {
		filters    interface{}
		wantByRepo map[string]interface{}
	}{
		{
			filters:    nil,
			wantByRepo: nil,
		},
		{
			filters:    []DefFilter{ByRepos("r")},
			wantByRepo: map[string]interface{}{"r": []DefFilter{}},
		},
		{
			filters:    []DefFilter{ByDefQuery("q1"), ByRepos("r"), ByDefQuery("q2")},
			wantByRepo: map[string]interface{}{"r": []DefFilter{ByDefQuery("q1"), ByDefQuery("q2")}},
		},
		{
			filters:    []DefFilter{ByRepos("r"), ByDefQuery("q1"), ByRepos("r"), ByDefQuery("q2"), ByRepos("r")},
			wantByRepo: map[string]interface{}{"r": []DefFilter{ByDefQuery("q1"), ByDefQuery("q2")}},
		},
		{
			filters:    []DefFilter{ByRepos("r"), ByUnits(unit.ID2{Type: "t", Name: "u"})},
			wantByRepo: map[string]interface{}{"r": []DefFilter{ByUnits(unit.ID2{Type: "t", Name: "u"})}},
		},
		{
			filters:    []DefFilter{ByRepoCommitIDs(Version{Repo: "r", CommitID: "c"})},
			wantByRepo: map[string]interface{}{"r": []DefFilter{ByCommitIDs("c")}},
		},
		{
			filters: []DefFilter{ByRepoCommitIDs(Version{Repo: "r1", CommitID: "c1"}, Version{Repo: "r2", CommitID: "c2"})},
			wantByRepo: map[string]interface{}{
				"r1": []DefFilter{ByCommitIDs("c1")},
				"r2": []DefFilter{ByCommitIDs("c2")},
			},
		},
	}
	for _, test := range tests {
		for repo, want := range test.wantByRepo {
			pre := fmt.Sprintf("%+v", test.filters)
			repoFilters := filtersForRepo(repo, test.filters)
			if !reflect.DeepEqual(repoFilters, want) {
				t.Errorf("%+v: repo %q: got repo filters %v, want %v", test.filters, repo, repoFilters, want)
			}
			if post := fmt.Sprintf("%+v", test.filters); pre != post {
				t.Errorf("%+v: filters modified: post filtersToRepo, filters == %v", test.filters, post)
			}
		}
	}
}
