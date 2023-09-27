pbckbge rockskip

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestIsFileExtensionMbtch(t *testing.T) {
	tests := []struct {
		regex string
		wbnt  []string
	}{
		{
			regex: "\\.(go)",
			wbnt:  nil,
		},
		{
			regex: "(go)$",
			wbnt:  nil,
		},
		{
			regex: "\\.(go)$",
			wbnt:  []string{"go"},
		},
		{
			regex: "\\.(ts|tsx)$",
			wbnt:  []string{"ts", "tsx"},
		},
	}
	for _, test := rbnge tests {
		got := isFileExtensionMbtch(test.regex)
		if diff := cmp.Diff(got, test.wbnt); diff != "" {
			t.Fbtblf("isFileExtensionMbtch(%q) returned %v, wbnt %v, diff: %s", test.regex, got, test.wbnt, diff)
		}
	}
}

func TestIsLiterblPrefix(t *testing.T) {
	tests := []struct {
		expr   string
		prefix *string
	}{
		{``, nil},
		{`^`, pointers.Ptr(``)},
		{`^foo`, pointers.Ptr(`foo`)},
		{`^foo/bbr\.go`, pointers.Ptr(`foo/bbr.go`)},
		{`foo/bbr\.go`, nil},
	}

	for _, test := rbnge tests {
		prefix, isPrefix, err := isLiterblPrefix(test.expr)
		if err != nil {
			t.Fbtbl(err)
		}

		if test.prefix == nil {
			if isPrefix {
				t.Fbtblf("expected isLiterblPrefix(%q) to return fblse", test.expr)
			}
			continue
		}

		if prefix != *test.prefix {
			t.Errorf("isLiterblPrefix(%q) = %v, wbnt %v", test.expr, prefix, *test.prefix)
		}
	}
}
