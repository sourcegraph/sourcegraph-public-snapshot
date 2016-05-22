package repotrackutil

import "testing"

func TestGetTrackedRepo(t *testing.T) {
	cases := []struct {
		Path        string
		TrackedRepo string
	}{
		// Top-level view
		{"/github.com/kubernetes/kubernetes", "github.com/kubernetes/kubernetes"},
		// Code view
		{"/github.com/kubernetes/kubernetes@master/-/tree/README.md", "github.com/kubernetes/kubernetes"},

		// Unrelated repo
		{"/github.com/gorilla/muxy@master/-/tree/mux.go", "unknown"},
		{"/github.com/gorilla/muxy", "unknown"},

		// Unrelated URL
		{"/blog/133554180524/announcing-the-sourcegraph-developer-release-the", "unknown"},

		// Corner case
		{"", "unknown"}, {"/", "unknown"},
	}
	for _, c := range cases {
		got := GetTrackedRepo(c.Path)
		if got != c.TrackedRepo {
			t.Errorf("getTrackedRepo(%#v) == %#v != %#v", c.Path, got, c.TrackedRepo)
		}
	}
	// a trackedRepo must always be tracked
	for _, r := range trackedRepo {
		if GetTrackedRepo(r) != r {
			t.Errorf("Repo should be tracked: %v", r)
		}
	}

}
