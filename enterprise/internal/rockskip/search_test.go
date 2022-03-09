package rockskip

import (
	"testing"

	"github.com/google/go-cmp/cmp"
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
