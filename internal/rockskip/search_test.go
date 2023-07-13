package rockskip

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestIsFileExtensionMatch(t *testing.T) {
	tests := []struct {
		regex string
		want  []string
	}{
		{
			regex: "\\.(go)",
			want:  nil,
		},
		{
			regex: "(go)$",
			want:  nil,
		},
		{
			regex: "\\.(go)$",
			want:  []string{"go"},
		},
		{
			regex: "\\.(ts|tsx)$",
			want:  []string{"ts", "tsx"},
		},
	}
	for _, test := range tests {
		got := isFileExtensionMatch(test.regex)
		if diff := cmp.Diff(got, test.want); diff != "" {
			t.Fatalf("isFileExtensionMatch(%q) returned %v, want %v, diff: %s", test.regex, got, test.want, diff)
		}
	}
}

func TestIsLiteralPrefix(t *testing.T) {
	tests := []struct {
		expr   string
		prefix *string
	}{
		{``, nil},
		{`^`, pointers.Ptr(``)},
		{`^foo`, pointers.Ptr(`foo`)},
		{`^foo/bar\.go`, pointers.Ptr(`foo/bar.go`)},
		{`foo/bar\.go`, nil},
	}

	for _, test := range tests {
		prefix, isPrefix, err := isLiteralPrefix(test.expr)
		if err != nil {
			t.Fatal(err)
		}

		if test.prefix == nil {
			if isPrefix {
				t.Fatalf("expected isLiteralPrefix(%q) to return false", test.expr)
			}
			continue
		}

		if prefix != *test.prefix {
			t.Errorf("isLiteralPrefix(%q) = %v, want %v", test.expr, prefix, *test.prefix)
		}
	}
}
