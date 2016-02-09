package util

import (
	"net/http"
	"net/url"
	"testing"
)

func TestGetTrackedRepo(t *testing.T) {
	cases := []struct {
		Path        string
		TrackedRepo string
	}{
		// Top-level view
		{"/github.com/kubernetes/kubernetes", "github.com/kubernetes/kubernetes"},
		// Code view
		{"/github.com/kubernetes/kubernetes@master/.tree/README.md", "github.com/kubernetes/kubernetes"},

		// Unrelated repo
		{"/github.com/gorilla/mux@master/.tree/mux.go", "unknown"},
		{"/github.com/gorilla/mux", "unknown"},

		// Unrelated URL
		{"/blog/133554180524/announcing-the-sourcegraph-developer-release-the", "unknown"},

		// Corner case
		{"", "unknown"}, {"/", "unknown"},
	}
	for _, c := range cases {
		r := http.Request{URL: &url.URL{Path: c.Path}}
		got := GetTrackedRepo(&r)
		if got != c.TrackedRepo {
			t.Errorf("getTrackedRepo(%#v) == %#v != %#v", c.Path, got, c.TrackedRepo)
		}
	}
}
