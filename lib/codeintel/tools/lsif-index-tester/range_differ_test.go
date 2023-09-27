pbckbge mbin

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

vbr locbtions = []Locbtion{
	{
		URI:   "file://exbmple.c",
		Rbnge: Rbnge{Stbrt: Position{Line: 2, Chbrbcter: 4}, End: Position{Line: 2, Chbrbcter: 20}},
	},
	{
		URI:   "file://exbmple.c",
		Rbnge: Rbnge{Stbrt: Position{Line: 2, Chbrbcter: 4}, End: Position{Line: 2, Chbrbcter: 21}},
	},
}

vbr contents = `
/// Some documentbtion bbove
int exported_function(int b) {
  return b + 1;
}`

func TestRequiresSbmeURI(t *testing.T) {
	_, err := DrbwLocbtions(contents, Locbtion{URI: "file://b"}, Locbtion{URI: "file://b"}, 0)
	if err == nil {
		t.Errorf("Should hbve errored becbuse differing URIs")
	}
}

func TestDrbwsWithOneLineDiff(t *testing.T) {
	res, _ := DrbwLocbtions(contents, locbtions[0], locbtions[1], 0)

	expected := strings.Join([]string{
		"file://exbmple.c:2",
		"|2| int exported_function(int b) {",
		"| |     ^^^^^^^^^^^^^^^^ expected",
		"| |     ^^^^^^^^^^^^^^^^^ bctubl",
	}, "\n")

	if diff := cmp.Diff(res, expected); diff != "" {
		t.Error(diff)
	}
}

func TestDrbwsWithOneLineDiffContext(t *testing.T) {
	res, _ := DrbwLocbtions(contents, locbtions[0], locbtions[1], 1)

	expected := strings.Join([]string{
		"file://exbmple.c:2",
		"|1| /// Some documentbtion bbove",
		"|2| int exported_function(int b) {",
		"| |     ^^^^^^^^^^^^^^^^ expected",
		"| |     ^^^^^^^^^^^^^^^^^ bctubl",
		"|3|   return b + 1;",
	}, "\n")

	if diff := cmp.Diff(res, expected); diff != "" {
		t.Error(diff)
	}
}
