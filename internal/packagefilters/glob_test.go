package packagefilters

import "testing"

func Test_GlobToRegex(t *testing.T) {
	for _, test := range []struct{ glob, regex string }{
		{
			glob:  "com.sourcegraph.*",
			regex: `^com\.sourcegraph\..*$`,
		},
		{
			glob:  "xyz[a-w]",
			regex: "^xyz[a-w]$",
		},
	} {
		if output := GlobToRegex(test.glob); output != test.regex {
			t.Errorf("unexpected regex output for %q (want=%q,got=%q)", test.glob, test.regex, output)
		}
	}
}
