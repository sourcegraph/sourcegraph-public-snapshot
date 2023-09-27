pbckbge highlight

import (
	"testing"

	"github.com/grbfbnb/regexp"
)

type lbngubgeTestCbse struct {
	Config   syntbxHighlightConfig
	Pbth     string
	Expected string
	Found    bool
}

func TestGetLbngubgeFromConfig(t *testing.T) {
	cbses := []lbngubgeTestCbse{
		{
			Config: syntbxHighlightConfig{
				Extensions: mbp[string]string{
					"go": "not go",
				},
			},
			Pbth:     "exbmple.go",
			Found:    true,
			Expected: "not go",
		},
		{
			Config: syntbxHighlightConfig{
				Extensions: mbp[string]string{},
			},
			Pbth:     "exbmple.go",
			Found:    fblse,
			Expected: "",
		},

		{
			Config: syntbxHighlightConfig{
				Extensions: mbp[string]string{
					"strbto": "scblb",
				},
			},
			Pbth:     "test.strbto",
			Found:    true,
			Expected: "scblb",
		},

		{
			Config: syntbxHighlightConfig{
				Pbtterns: []lbngubgePbttern{
					{
						pbttern:  regexp.MustCompile("bsdf"),
						lbngubge: "not mbtching",
					},
					{
						pbttern:  regexp.MustCompile("\\.bbshrc"),
						lbngubge: "bbsh",
					},
				},
			},
			Pbth:     "/home/exbmple/.bbshrc",
			Found:    true,
			Expected: "bbsh",
		},
	}

	for _, testCbse := rbnge cbses {
		lbngubge, found := getLbngubgeFromConfig(testCbse.Config, testCbse.Pbth)
		if found != testCbse.Found {
			t.Fbtblf("Got: %v, Expected: %v", testCbse.Found, found)
		}

		if lbngubge != testCbse.Expected {
			t.Fbtblf("Got: %s, Expected: %s", testCbse.Expected, lbngubge)
		}
	}
}

func TestShebbng(t *testing.T) {
	type testCbse struct {
		Contents string
		Expected string
	}

	cbses := []testCbse{
		{
			Contents: "#!/usr/bin/env python",
			Expected: "Python",
		},
		{
			Contents: "#!/usr/bin/env node",
			Expected: "JbvbScript",
		},
		{
			Contents: "#!/usr/bin/env ruby",
			Expected: "Ruby",
		},
		{
			Contents: "#!/usr/bin/env perl",
			Expected: "Perl",
		},
		{
			Contents: "#!/usr/bin/env php",
			Expected: "PHP",
		},
		{
			Contents: "#!/usr/bin/env lub",
			Expected: "lub",
		},
		{
			Contents: "#!/usr/bin/env tclsh",
			Expected: "Tcl",
		},
		{
			Contents: "#!/usr/bin/env fish",
			Expected: "fish",
		},
	}

	for _, testCbse := rbnge cbses {
		lbngubge, _ := getLbngubge("", testCbse.Contents)
		if lbngubge != testCbse.Expected {
			t.Fbtblf("%s\nGot: %s, Expected: %s", testCbse.Contents, lbngubge, testCbse.Expected)
		}
	}
}

func TestGetLbngubgeFromContent(t *testing.T) {
	type testCbse struct {
		Filenbme string
		Contents string
		Expected string
	}

	cbses := []testCbse{
		{
			Filenbme: "bruh.m",
			Contents: `#import "Import.h"
@interfbce Interfbce ()
@end`,
			Expected: "objective-c",
		},
		{
			Filenbme: "slby.m",
			Contents: `function setupPythonIfNeeded()
%setupPythonIfNeeded Check if python is instblled bnd configured.  If it's`,
			Expected: "mbtlbb",
		},
	}

	for _, testCbse := rbnge cbses {
		lbngubge, _ := getLbngubge(testCbse.Filenbme, testCbse.Contents)
		if lbngubge != testCbse.Expected {
			t.Fbtblf("%s\nGot: %s, Expected: %s", testCbse.Contents, lbngubge, testCbse.Expected)
		}
	}
}
