package gitdomain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateBranchName(t *testing.T) {
	for _, tc := range []struct {
		name   string
		branch string
		valid  bool
	}{
		{name: "Valid branch", branch: "valid-branch", valid: true},
		{name: "Valid branch with slash", branch: "rgs/valid-branch", valid: true},
		{name: "Valid branch with @", branch: "valid@branch", valid: true},
		{name: "Path component with .", branch: "valid-/.branch", valid: false},
		{name: "Double dot", branch: "valid..branch", valid: false},
		{name: "End with .lock", branch: "valid-branch.lock", valid: false},
		{name: "No space", branch: "valid branch", valid: false},
		{name: "No tilde", branch: "valid~branch", valid: false},
		{name: "No carat", branch: "valid^branch", valid: false},
		{name: "No colon", branch: "valid:branch", valid: false},
		{name: "No question mark", branch: "valid?branch", valid: false},
		{name: "No asterisk", branch: "valid*branch", valid: false},
		{name: "No open bracket", branch: "valid[branch", valid: false},
		{name: "No trailing slash", branch: "valid-branch/", valid: false},
		{name: "No beginning slash", branch: "/valid-branch", valid: false},
		{name: "No double slash", branch: "valid//branch", valid: false},
		{name: "No trailing dot", branch: "valid-branch.", valid: false},
		{name: "Cannot contain @{", branch: "valid@{branch", valid: false},
		{name: "Cannot be @", branch: "@", valid: false},
		{name: "Cannot contain backslash", branch: "valid\\branch", valid: false},
		{name: "head not allowed", branch: "head", valid: false},
		{name: "Head not allowed", branch: "Head", valid: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			valid := ValidateBranchName(tc.branch)
			assert.Equal(t, tc.valid, valid)
		})
	}
}

func TestRefGlobs(t *testing.T) {
	tests := map[string]struct {
		globs   []RefGlob
		match   []string
		noMatch []string
		want    []string
	}{
		"empty": {
			globs:   nil,
			noMatch: []string{"a"},
		},
		"globs": {
			globs:   []RefGlob{{Include: "refs/heads/"}},
			match:   []string{"refs/heads/a", "refs/heads/b/c"},
			noMatch: []string{"refs/tags/t"},
		},
		"excludes": {
			globs: []RefGlob{
				{Include: "refs/heads/"}, {Exclude: "refs/heads/x"},
			},
			match:   []string{"refs/heads/a", "refs/heads/b", "refs/heads/x/c"},
			noMatch: []string{"refs/tags/t", "refs/heads/x"},
		},
		"implicit leading refs/": {
			globs: []RefGlob{{Include: "heads/"}},
			match: []string{"refs/heads/a"},
		},
		"implicit trailing /*": {
			globs:   []RefGlob{{Include: "refs/heads/a"}},
			match:   []string{"refs/heads/a", "refs/heads/a/b"},
			noMatch: []string{"refs/heads/b"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			m, err := CompileRefGlobs(test.globs)
			if err != nil {
				t.Fatal(err)
			}
			for _, ref := range test.match {
				if !m.Match(ref) {
					t.Errorf("want match %q", ref)
				}
			}
			for _, ref := range test.noMatch {
				if m.Match(ref) {
					t.Errorf("want no match %q", ref)
				}
			}
		})
	}
}

func TestIsAbsoluteRevision(t *testing.T) {
	yes := []string{"8cb03d28ad1c6a875f357c5d862237577b06e57c", "20697a062454c29d84e3f006b22eb029d730cd00"}
	no := []string{"ref: refs/heads/appsinfra/SHEP-20-review", "master", "HEAD", "refs/heads/master", "20697a062454c29d84e3f006b22eb029d730cd0", "20697a062454c29d84e3f006b22eb029d730cd000", "  20697a062454c29d84e3f006b22eb029d730cd00  ", "20697a062454c29d84e3f006b22eb029d730cd0 "}
	for _, s := range yes {
		if !IsAbsoluteRevision(s) {
			t.Errorf("%q should be an absolute revision", s)
		}
	}
	for _, s := range no {
		if IsAbsoluteRevision(s) {
			t.Errorf("%q should not be an absolute revision", s)
		}
	}
}
