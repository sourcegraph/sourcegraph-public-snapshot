package uri

import (
	"net/url"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := map[string]url.URL{}
	for uriStr, want := range tests {
		t.Run(strings.Replace(uriStr, "/", "-", -1), func(t *testing.T) {
			uri, err := Parse(uriStr)
			if err != nil {
				t.Fatal(err)
			}
			if uri.URL != want {
				t.Errorf("got %+v, want %+v", uri.URL, want)
			}
		})
	}
}

func TestParse_error(t *testing.T) {
	tests := map[string]string{
		"github.com/foo/bar":   "must be absolute",
		"/github.com/foo/bar":  "must be absolute",
		"//github.com/foo/bar": "must be absolute",
		"%": "invalid",
	}
	for uriStr, want := range tests {
		t.Run(strings.Replace(uriStr, "/", "-", -1), func(t *testing.T) {
			uri, err := Parse(uriStr)
			if err == nil {
				t.Fatalf("got nil error, want %q", want)
			}
			if uri != nil {
				t.Error("got non-nil URL, want nil")
			}
			if !strings.Contains(err.Error(), want) {
				t.Errorf("got %q, want it to contain %q", err, want)
			}
		})
	}
}

func TestURI_CloneURL(t *testing.T) {
	want := "https://github.com/foo/bar"
	uriStrs := []string{
		"https://github.com/foo/bar",
		"https://github.com/foo/bar?v",
		"https://github.com/foo/bar?v#f",
		"https://github.com/foo/bar#f",
	}
	for _, uriStr := range uriStrs {
		t.Run(strings.Replace(uriStr, "/", "-", -1), func(t *testing.T) {
			uri, err := Parse(uriStr)
			if err != nil {
				t.Fatal(err)
			}
			if uri.CloneURL().String() != want {
				t.Errorf("got %s, want %s", uri.CloneURL(), want)
			}
		})
	}
}

func TestURI_Rev(t *testing.T) {
	tests := map[string]string{
		"https://github.com/foo/bar":     "",
		"https://github.com/foo/bar?v":   "v",
		"https://github.com/foo/bar?v#":  "v",
		"https://github.com/foo/bar?v#f": "v",
	}
	for uriStr, want := range tests {
		t.Run(strings.Replace(uriStr, "/", "-", -1), func(t *testing.T) {
			uri, err := Parse(uriStr)
			if err != nil {
				t.Fatal(err)
			}
			if uri.Rev() != want {
				t.Errorf("got %s, want %s", uri.Rev(), want)
			}
		})
	}
}

func TestURI_FilePath(t *testing.T) {
	tests := map[string]string{
		"https://github.com/foo/bar":          "",
		"https://github.com/foo/bar?v":        "",
		"https://github.com/foo/bar?v#":       "",
		"https://github.com/foo/bar?v#.":      "",
		"https://github.com/foo/bar?v#f":      "f",
		"https://github.com/foo/bar?v#/f":     "f",
		"https://github.com/foo/bar?v#f/d":    "f/d",
		"https://github.com/foo/bar?v#f/..":   "",
		"https://github.com/foo/bar?v#f/d/..": "f",
		"https://github.com/foo/bar?v#//":     "",
		"https://github.com/foo/bar?v#d%2Ff":  "d/f",
	}
	for uriStr, want := range tests {
		t.Run(strings.Replace(uriStr, "/", "-", -1), func(t *testing.T) {
			uri, err := Parse(uriStr)
			if err != nil {
				t.Fatal(err)
			}
			if uri.FilePath() != want {
				t.Errorf("got %s, want %s", uri.FilePath(), want)
			}
		})
	}
}

// func TestURI_HasPrefix(t *testing.T) {
// 	tests := map[string]map[string]bool{
// 		"s://h/h?v#a/b": {
// 			"s://h/h?v":       true,
// 			"s://h/h?v#":      true,
// 			"s://h/h?v#a":     true,
// 			"s://h/h?v#a/b":   true,
// 			"s://h/h?v#a/b/c": false,
// 			"s://h/h?v#a/d":   false,
// 			"s://h/h?v#d":     false,
// 		},
// 	}
// 	for a, subtest := range tests {
// 		for b, want := range subtest {
// 			t.Run(a+" and "+b, func(t *testing.T) {
// 				a, err := Parse(a)
// 				if err != nil {
// 					t.Fatal(err)
// 				}
// 				b, err := Parse(b)
// 				if err != nil {
// 					t.Fatal(err)
// 				}
// 				if a.HasPrefix(b) != want {
// 					t.Errorf("got HasPrefix %v, want %v", !want, want)
// 				}
// 			})
// 		}
// 	}
// }
