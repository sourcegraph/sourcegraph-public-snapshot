package git

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseGitStatus(t *testing.T) {
	const input = `M  README.md
M  another_file.go
A  new_file.txt
A  barfoo/new_file.txt
D  to_be_deleted.txt
R  README.md -> README.markdown
`
	parsed, err := ParseGitStatus([]byte(input))
	if err != nil {
		t.Fatal(err)
	}

	want := Changes{
		Modified: []string{"README.md", "another_file.go"},
		Added:    []string{"new_file.txt", "barfoo/new_file.txt"},
		Deleted:  []string{"to_be_deleted.txt"},
		Renamed:  []string{"README.markdown"},
	}

	if !cmp.Equal(want, parsed) {
		t.Fatalf("wrong output:\n%s", cmp.Diff(want, parsed))
	}
}
