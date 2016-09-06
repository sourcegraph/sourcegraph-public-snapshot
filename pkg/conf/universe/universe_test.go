package universe

import "testing"

func TestRepoChecker(t *testing.T) {
	cases := []struct {
		Enabled string
		Repo    string
		Want    bool
	}{
		// Test that sourcegraph is default
		{"", "foo", false},
		{"", "github.com/sourcegraph/sourcegraph", true},

		// Test all
		{"all", "foo", true},
		{"all", "github.com/sourcegraph/sourcegraph", true},

		// Test specified
		{"foo", "foo", true},
		{"foo", "github.com/sourcegraph/sourcegraph", false},
	}
	for _, c := range cases {
		if repoChecker(false, c.Enabled, c.Repo) {
			t.Errorf("repoChecker(false, %v, %v) should be false", c.Enabled, c.Repo)
		}
		if repoChecker(true, c.Enabled, c.Repo) != c.Want {
			t.Errorf("repoChecker(true, %v, %v) = %v, want %v", c.Enabled, c.Repo, !c.Want, c.Want)
		}
	}
}
