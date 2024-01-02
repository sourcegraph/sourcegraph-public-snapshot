package search

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
)

func TestCompilePathPatterns(t *testing.T) {
	match, err := compilePathPatterns(&protocol.PatternInfo{
		IncludePatterns: []string{`main\.go`, `m`},
		ExcludePattern:  `README\.md`,
		IsCaseSensitive: false,
	})
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]bool{
		"README.md": false,
		"main.go":   true,
	}
	for path, want := range want {
		got := match.Matches(path)
		if got != want {
			t.Errorf("path %q: got %v, want %v", path, got, want)
			continue
		}
	}
}
