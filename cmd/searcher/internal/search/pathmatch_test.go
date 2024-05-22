package search

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/searcher/protocol"
)

func TestCompilePathPatterns(t *testing.T) {
	match, err := toPathMatcher(&protocol.PatternInfo{
		IncludePaths:    []string{`main\.go`, `m`},
		ExcludePaths:    `README\.md`,
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

func TestCompileLangPatterns(t *testing.T) {
	match, err := toPathMatcher(&protocol.PatternInfo{
		IncludeLangs: []string{"Go"},
		ExcludeLangs: []string{"Markdown"},
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
