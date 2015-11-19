package sourcecode

import (
	"strings"
	"testing"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"src.sourcegraph.com/sourcegraph/util/testutil/srclibtest"
)

func Test_HtmlEscapeStringWithCodeBreaks(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{{
		"foo.bar:baz", "foo.<wbr>bar:<wbr>baz",
	}, {
		"foo", "foo",
	}, {
		"foo:bar.baz(=>Int):blah.blah.Whatever[A]", "foo:<wbr>bar.<wbr>baz(=&gt;Int):<wbr>blah.<wbr>blah.<wbr>Whatever[A]",
	}, {
		"foo<bar>", "foo&lt;bar&gt;",
	}}

	for _, test := range tests {
		out := htmlEscapeStringWithCodeBreaks(test.in)
		if test.out != out {
			t.Errorf(`
Exp "%s",
got "%s" for input "%s"`, test.out, out, test.in)
		}
	}
}

// Test that DefQualifiedName wraps the def's name in a <span
// class="name"/> tag.
func TestDefQualifiedName_nameSpan(t *testing.T) {
	graph.RegisterMakeDefFormatter("test", func(*graph.Def) graph.DefFormatter { return srclibtest.Formatter{} })

	def := &sourcegraph.Def{Def: graph.Def{
		DefKey: graph.DefKey{Repo: "x.com/r", UnitType: "test", Unit: "u", Path: "p"},
		Name:   "name",
	}}

	for _, qual := range graph.QualLevels {
		h := DefQualifiedName(def, string(qual))
		nameSpan := `<span class="name">` + def.Name + `</span>`
		if !strings.Contains(string(h), nameSpan) {
			t.Errorf("DefQualifiedName qual=%q: got %q, want it to contain %q", qual, h, nameSpan)
		}
	}
}
