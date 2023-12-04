package search

import "testing"

func TestCompilePathPatterns(t *testing.T) {
	match, err := compilePathPatterns([]string{`main\.go`, `m`}, `README\.md`, false)
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]bool{
		"README.md": false,
		"main.go":   true,
	}
	for path, want := range want {
		got := match.MatchPath(path)
		if got != want {
			t.Errorf("path %q: got %v, want %v", path, got, want)
			continue
		}
	}
}
