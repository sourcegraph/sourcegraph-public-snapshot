package universe

import "testing"

func TestRepoChecker(t *testing.T) {
	cases := []struct {
		Enabled string
		Repo    string
		Want    bool
	}{
		// Test that RxJava is default
		{"", "foo", false},
		{"", "github.com/slimsag/RxJava", true},

		// Test all
		{"all", "foo", true},
		{"all", "github.com/slimsag/RxJava", true},

		// Test specified
		{"foo", "foo", true},
		{"foo", "github.com/slimsag/RxJava", false},
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
