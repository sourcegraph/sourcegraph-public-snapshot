package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var locations = []Location{
	{
		URI:   "file://example.c",
		Range: Range{Start: Position{Line: 2, Character: 4}, End: Position{Line: 2, Character: 20}},
	},
	{
		URI:   "file://example.c",
		Range: Range{Start: Position{Line: 2, Character: 4}, End: Position{Line: 2, Character: 21}},
	},
}

var contents = `
/// Some documentation above
int exported_function(int a) {
  return a + 1;
}`

func TestRequiresSameURI(t *testing.T) {
	_, err := DrawLocations(contents, Location{URI: "file://a"}, Location{URI: "file://b"}, 0)
	if err == nil {
		t.Errorf("Should have errored because differing URIs")
	}
}

func TestDrawsWithOneLineDiff(t *testing.T) {
	res, _ := DrawLocations(contents, locations[0], locations[1], 0)

	expected := strings.Join([]string{
		"file://example.c:2",
		"|2| int exported_function(int a) {",
		"| |     ^^^^^^^^^^^^^^^^ expected",
		"| |     ^^^^^^^^^^^^^^^^^ actual",
	}, "\n")

	if diff := cmp.Diff(res, expected); diff != "" {
		t.Error(diff)
	}
}

func TestDrawsWithOneLineDiffContext(t *testing.T) {
	res, _ := DrawLocations(contents, locations[0], locations[1], 1)

	expected := strings.Join([]string{
		"file://example.c:2",
		"|1| /// Some documentation above",
		"|2| int exported_function(int a) {",
		"| |     ^^^^^^^^^^^^^^^^ expected",
		"| |     ^^^^^^^^^^^^^^^^^ actual",
		"|3|   return a + 1;",
	}, "\n")

	if diff := cmp.Diff(res, expected); diff != "" {
		t.Error(diff)
	}
}
