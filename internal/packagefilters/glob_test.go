package packagefilters

import (
	"testing"

	"github.com/gobwas/glob"
	"github.com/grafana/regexp"
)

func Test_GlobToRegex(t *testing.T) {
	for _, test := range []struct{ glob, regex string }{
		{glob: "com.sourcegraph.*", regex: `^com\.sourcegraph\..*$`},
		{glob: "xyz[a-w]", regex: "^xyz[a-w]$"},
		{glob: "asd[!f]", regex: "^asd[^f]$"},
		{glob: "b?n?n?", regex: "^b.n.n.$"},
		{glob: "asdf[!!]", regex: "^asdf[^!]$"},
		{glob: "*****", regex: "^.*.*.*$"},
		{glob: "!fdsa]", regex: `^!fdsa\]$`},
		{glob: `[\d]`, regex: `^[d]$`},
		{glob: "{asd,abc}f", regex: "^(?:asd|abc)f$"},
		{glob: "asdf", regex: "^asdf$"},
	} {
		if _, err := glob.Compile(test.glob); err != nil {
			t.Fatalf("not a valid glob %s %v", test.glob, err)
		}
		output, _ := GlobToRegex(test.glob)
		if output != test.regex {
			t.Errorf("unexpected regex output for %q (want=%q, got=%q)", test.glob, test.regex, output)
		}

		if _, err := regexp.Compile(output); err != nil {
			t.Errorf("output is not valid regex. %q -> %v", output, err)
		}
	}
}
