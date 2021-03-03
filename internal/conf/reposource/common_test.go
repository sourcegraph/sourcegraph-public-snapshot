package reposource

import (
	"testing"
)

// urlToRepoName represents a cloneURL and expected corresponding repo name
type urlToRepoName struct {
	cloneURL string
	repoName string
}

// urlToRepoNameErr is similar to urlToRepoName, but with an expected error value
type urlToRepoNameErr struct {
	cloneURL string
	repoName string
	err      error
}

func TestNameTransformations(t *testing.T) {
	opts := []NameTransformationOptions{
		{
			Regex:       `\.d/`,
			Replacement: "/",
		},
		{
			Regex:       "-git$",
			Replacement: "",
		},
	}

	nts := make([]NameTransformation, len(opts))
	for i, opt := range opts {
		nt, err := NewNameTransformation(opt)
		if err != nil {
			t.Fatalf("NewNameTransformation: %v", err)
		}
		nts[i] = nt
	}

	tests := []struct {
		input  string
		output string
	}{
		{"path/to.d/repo-git", "path/to/repo"},
		{"path/to.d/repo-git.git", "path/to/repo-git.git"},
		{"path/to.de/repo-git.git", "path/to.de/repo-git.git"},
	}
	for _, test := range tests {
		got := NameTransformations(nts).Transform(test.input)
		if test.output != got {
			t.Errorf("for input %s, expected %s, but got %s", test.input, test.output, got)
		}
	}
}
