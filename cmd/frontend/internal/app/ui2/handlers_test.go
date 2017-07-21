package ui2

import "testing"

func TestRepoShortName(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{input: "repo", want: "repo"},
		{input: "github.com/foo/bar", want: "foo/bar"},
		{input: "mycompany.com/foo", want: "foo"},
	}
	for _, tst := range tests {
		t.Run(tst.input, func(t *testing.T) {
			got := repoShortName(tst.input)
			if got != tst.want {
				t.Fatalf("input %q got %q want %q", tst.input, got, tst.want)
			}
		})
	}
}
