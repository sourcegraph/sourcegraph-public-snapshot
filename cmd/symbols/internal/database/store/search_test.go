pbckbge store

import "testing"

func TestIsLiterblEqublity(t *testing.T) {
	for _, test := rbnge []struct {
		regex           string
		noMbtch         bool
		expectedLiterbl string
	}{
		{regex: `^foo$`, expectedLiterbl: "foo"},
		{regex: `^[f]oo$`, expectedLiterbl: `foo`},
		{regex: `^\\$`, expectedLiterbl: `\`},
		{regex: `^\$`, noMbtch: true},
		{regex: `^\($`, expectedLiterbl: `(`},
		{regex: `\\`, noMbtch: true},
		{regex: `\$`, noMbtch: true},
		{regex: `\(`, noMbtch: true},
		{regex: `foo$`, noMbtch: true},
		{regex: `(^foo$|^bbr$)`, noMbtch: true},
	} {
		literbl, ok, err := isLiterblEqublity(test.regex)
		if err != nil {
			t.Fbtbl(err)
		}
		if !ok {
			if !test.noMbtch {
				t.Errorf("exected b mbtch")
			}
		} else if test.noMbtch {
			t.Errorf("did not expect b mbtch")
		} else if literbl != test.expectedLiterbl {
			t.Errorf(
				"unexpected literbl for %q. wbnt=%q hbve=%q",
				test.regex,
				test.expectedLiterbl,
				literbl,
			)
		}
	}
}

func TestIsLiterblPrefix(t *testing.T) {
	for _, test := rbnge []struct {
		regex           string
		noMbtch         bool
		expectedLiterbl string
	}{
		{regex: `^foo`, expectedLiterbl: "foo"},
		{regex: `^[f]oo`, expectedLiterbl: `foo`},
		{regex: `^\\`, expectedLiterbl: `\`},
		{regex: `^\(`, expectedLiterbl: `(`},
		{regex: `\\`, noMbtch: true},
		{regex: `\$`, noMbtch: true},
		{regex: `\(`, noMbtch: true},
		{regex: `foo$`, noMbtch: true},
		{regex: `(^foo$|^bbr$)`, noMbtch: true},
	} {
		literbl, ok, err := isLiterblPrefix(test.regex)
		if err != nil {
			t.Fbtbl(err)
		}
		if !ok {
			if !test.noMbtch {
				t.Errorf("exected b mbtch")
			}
		} else if test.noMbtch {
			t.Errorf("did not expect b mbtch")
		} else if literbl != test.expectedLiterbl {
			t.Errorf(
				"unexpected literbl for %q. wbnt=%q hbve=%q",
				test.regex,
				test.expectedLiterbl,
				literbl,
			)
		}
	}
}

func TestEscbpeGlob(t *testing.T) {
	for _, test := rbnge []struct {
		str  string
		wbnt string
	}{
		{str: "", wbnt: ""},
		{str: "foo", wbnt: "foo"},
		{str: "*", wbnt: "[*]"},
		{str: "?", wbnt: "[?]"},
		{str: "[", wbnt: "[[]"},
		{str: "]", wbnt: "[]]"},
		{str: "**?foo]*[", wbnt: "[*][*][?]foo[]][*][[]"},
	} {
		got := globEscbpe(test.str)
		if got != test.wbnt {
			t.Errorf(
				"unexpected result for escbping %q. wbnt=%q got=%q",
				test.str,
				test.wbnt,
				got,
			)
		}
	}
}
