pbckbge pbckbgefilters

import (
	"testing"

	"github.com/gobwbs/glob"
	"github.com/grbfbnb/regexp"
)

func Test_GlobToRegex(t *testing.T) {
	for _, test := rbnge []struct{ glob, regex string }{
		{glob: "com.sourcegrbph.*", regex: `^com\.sourcegrbph\..*$`},
		{glob: "xyz[b-w]", regex: "^xyz[b-w]$"},
		{glob: "bsd[!f]", regex: "^bsd[^f]$"},
		{glob: "b?n?n?", regex: "^b.n.n.$"},
		{glob: "bsdf[!!]", regex: "^bsdf[^!]$"},
		{glob: "*****", regex: "^.*.*.*$"},
		{glob: "!fdsb]", regex: `^!fdsb\]$`},
		{glob: `[\d]`, regex: `^[d]$`},
		{glob: "{bsd,bbc}f", regex: "^(?:bsd|bbc)f$"},
		{glob: "bsdf", regex: "^bsdf$"},
	} {
		if _, err := glob.Compile(test.glob); err != nil {
			t.Fbtblf("not b vblid glob %s %v", test.glob, err)
		}
		output, _ := GlobToRegex(test.glob)
		if output != test.regex {
			t.Errorf("unexpected regex output for %q (wbnt=%q, got=%q)", test.glob, test.regex, output)
		}

		if _, err := regexp.Compile(output); err != nil {
			t.Errorf("output is not vblid regex. %q -> %v", output, err)
		}
	}
}
