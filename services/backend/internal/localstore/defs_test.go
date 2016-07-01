package localstore

import (
	"testing"

	"sourcegraph.com/sourcegraph/srclib/graph"
)

func TestToTextSearchTokens(t *testing.T) {
	aToks, bToks, cToks, dToks := toTextSearchTokens(&graph.Def{
		DefKey: graph.DefKey{
			Repo: "repo1/repo2",
			Unit: "unit1/unit2",
			Path: "path_x/pathFooBar",
		},
		File: "file1/file2",
		Name: "name",
		Docs: []*graph.DefDoc{
			&graph.DefDoc{Data: "foo <b>bar</b>"},
			&graph.DefDoc{Data: "baz"},
		},
	})

	expectedAToks := []string{"pathFooBar", "pathFooBar", "pathFooBar", "pathFooBar", "FooBar", "pathBar", "Bar", "pathFoo", "Foo", "path", "name"}
	expectedBToks := []string{"repo1", "repo2", "repo2", "repo2", "unit1", "unit2", "unit2", "unit2", "path_x", "path_x", "pathFooBar", "pathFooBar"}
	expectedCToks := []string{"pathx", "x", "path", "pathFooBar", "FooBar", "pathBar", "Bar", "pathFoo", "Foo", "path", "file1", "file2", "file2", "file2"}
	expectedDToks := []string{"foo bar", "baz"}

	if !stringSliceEqual(aToks, expectedAToks) {
		t.Errorf("wrong aToks, expected %#v, got %#v", expectedAToks, aToks)
	}
	if !stringSliceEqual(bToks, expectedBToks) {
		t.Errorf("wrong aToks, expected %#v, got %#v", expectedBToks, bToks)
	}
	if !stringSliceEqual(cToks, expectedCToks) {
		t.Errorf("wrong aToks, expected %#v, got %#v", expectedCToks, cToks)
	}
	if !stringSliceEqual(dToks, expectedDToks) {
		t.Errorf("wrong aToks, expected %#v, got %#v", expectedDToks, dToks)
	}
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
