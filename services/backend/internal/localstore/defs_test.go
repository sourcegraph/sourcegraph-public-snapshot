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
			Path: "path_x/pathFooBarHelloWorldThisIsLong",
		},
		File: "file1/file2",
		Name: "name",
		Docs: []*graph.DefDoc{
			&graph.DefDoc{Data: "foo <b>bar</b>"},
			&graph.DefDoc{Data: "baz"},
		},
	})

	expectedAToks := []string{"pathFooBarHelloWorldThisIsLong", "pathFooBarHelloWorldThisIsLong", "pathFooBarHelloWorldThisIsLong", "pathFooBarHelloWorldThis", "FooBarHelloWorldThis", "pathBarHelloWorldThis", "BarHelloWorldThis", "pathFooHelloWorldThis", "FooHelloWorldThis", "pathHelloWorldThis", "HelloWorldThis", "pathFooBarWorldThis", "FooBarWorldThis", "pathBarWorldThis", "BarWorldThis", "pathFooWorldThis", "FooWorldThis", "pathWorldThis", "WorldThis", "pathFooBarHelloThis", "FooBarHelloThis", "pathBarHelloThis", "BarHelloThis", "pathFooHelloThis", "FooHelloThis", "pathHelloThis", "HelloThis", "pathFooBarThis", "FooBarThis", "pathBarThis", "BarThis", "pathFooThis", "FooThis", "pathThis", "This", "pathFooBarHelloWorld", "FooBarHelloWorld", "pathBarHelloWorld", "BarHelloWorld", "pathFooHelloWorld", "FooHelloWorld", "pathHelloWorld", "HelloWorld", "pathFooBarWorld", "FooBarWorld", "pathBarWorld", "BarWorld", "pathFooWorld", "FooWorld", "pathWorld", "World", "pathFooBarHello", "FooBarHello", "pathBarHello", "BarHello", "pathFooHello", "FooHello", "pathHello", "Hello", "pathFooBar", "FooBar", "pathBar", "Bar", "pathFoo", "Foo", "path", "Is", "Long", "name"}
	expectedBToks := []string{"repo1", "repo2", "repo2", "repo2", "unit1", "unit2", "unit2", "unit2", "path_x", "path_x", "pathFooBarHelloWorldThisIsLong", "pathFooBarHelloWorldThisIsLong"}
	expectedCToks := []string{"pathx", "x", "path", "pathFooBarHelloWorldThis", "FooBarHelloWorldThis", "pathBarHelloWorldThis", "BarHelloWorldThis", "pathFooHelloWorldThis", "FooHelloWorldThis", "pathHelloWorldThis", "HelloWorldThis", "pathFooBarWorldThis", "FooBarWorldThis", "pathBarWorldThis", "BarWorldThis", "pathFooWorldThis", "FooWorldThis", "pathWorldThis", "WorldThis", "pathFooBarHelloThis", "FooBarHelloThis", "pathBarHelloThis", "BarHelloThis", "pathFooHelloThis", "FooHelloThis", "pathHelloThis", "HelloThis", "pathFooBarThis", "FooBarThis", "pathBarThis", "BarThis", "pathFooThis", "FooThis", "pathThis", "This", "pathFooBarHelloWorld", "FooBarHelloWorld", "pathBarHelloWorld", "BarHelloWorld", "pathFooHelloWorld", "FooHelloWorld", "pathHelloWorld", "HelloWorld", "pathFooBarWorld", "FooBarWorld", "pathBarWorld", "BarWorld", "pathFooWorld", "FooWorld", "pathWorld", "World", "pathFooBarHello", "FooBarHello", "pathBarHello", "BarHello", "pathFooHello", "FooHello", "pathHello", "Hello", "pathFooBar", "FooBar", "pathBar", "Bar", "pathFoo", "Foo", "path", "Is", "Long", "file1", "file2", "file2", "file2"}
	expectedDToks := []string{"foo bar", "baz"}

	if !stringSliceEqual(aToks, expectedAToks) {
		t.Errorf("wrong aToks, expected %#v, got %#v", expectedAToks, aToks)
	}
	if !stringSliceEqual(bToks, expectedBToks) {
		t.Errorf("wrong bToks, expected %#v, got %#v", expectedBToks, bToks)
	}
	if !stringSliceEqual(cToks, expectedCToks) {
		t.Errorf("wrong cToks, expected %#v, got %#v", expectedCToks, cToks)
	}
	if !stringSliceEqual(dToks, expectedDToks) {
		t.Errorf("wrong dToks, expected %#v, got %#v", expectedDToks, dToks)
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
