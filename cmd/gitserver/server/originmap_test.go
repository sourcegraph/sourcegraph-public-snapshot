package server

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestMap(t *testing.T) {
	type testc struct {
		in       string
		exp      []prefixAndOrgin
		mappings [][2]string
	}
	tests := []testc{{
		in: "github.com/!https://github.com/%.git",
		exp: []prefixAndOrgin{{
			Prefix: "github.com/",
			Origin: "https://github.com/%.git",
		}},
		mappings: [][2]string{
			{"github.com/gorilla/mux", "https://github.com/gorilla/mux.git"},
			{"github.com/gorilla/pat", "https://github.com/gorilla/pat.git"},
		},
	}, {
		in: "local/!local/%",
		exp: []prefixAndOrgin{{
			Prefix: "local/",
			Origin: "local/%",
		}},
		mappings: [][2]string{
			{"local/foo", "local/foo"},
		},
	}, {
		in: "local/!local/% github.com/!https://github.com/%.git",
		exp: []prefixAndOrgin{{
			Prefix: "local/",
			Origin: "local/%",
		}, {
			Prefix: "github.com/",
			Origin: "https://github.com/%.git",
		}},
		mappings: [][2]string{
			{"github.com/gorilla/mux", "https://github.com/gorilla/mux.git"},
			{"github.com/gorilla/pat", "https://github.com/gorilla/pat.git"},
			{"local/foo", "local/foo"},
			{"nomatch", ""},
		},
	}}

	restoreOriginMap := originMap[:]
	defer func() {
		originMap = restoreOriginMap
	}()

	for _, test := range tests {
		actual, err := parse(test.in)
		if err != nil {
			t.Errorf("on input %q, unexpected err: %v", test.in, err)
			continue
		}
		if !reflect.DeepEqual(test.exp, actual) {
			t.Errorf("on input %q, expected %s, but got %s", test.in, spew.Sdump(test.exp), spew.Sdump(actual))
		}

		originMap = actual
		for _, mapping := range test.mappings {
			if gotURI := OriginMap(mapping[0]); gotURI != mapping[1] {
				t.Errorf("on input %q, input URI %q, got %q, but expected %q", test.in, mapping[0], gotURI, mapping[1])
			}
		}
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		name, in string
		mappings [][2]string
	}{
		{
			name: "github",
			in:   "github.com/!https://github.com/%.git",
			mappings: [][2]string{
				{"https://github.com/gorilla/mux.git", "github.com/gorilla/mux"},
				{"https://github.com/gorilla/pat.git", "github.com/gorilla/pat"},
			},
		},
		{
			name: "local",
			in:   "local/!local/%",
			mappings: [][2]string{
				{"local/foo", "local/foo"},
			},
		},
		{
			name: "local_and_github",
			in:   "local/!local/% github.com/!https://github.com/%.git",
			mappings: [][2]string{
				{"https://github.com/gorilla/mux.git", "github.com/gorilla/mux"},
				{"https://github.com/gorilla/pat.git", "github.com/gorilla/pat"},
				{"local/foo", "local/foo"},
				{"nomatch", ""},
			},
		},

		// Test inputs for those that are in addGitHubDefaults
		{
			name: "http_remote_url",
			in:   "github.com/!http://github.com/%.git",
			mappings: [][2]string{
				{"http://github.com/gorilla/mux.git", "github.com/gorilla/mux"},
				{"http://github.com/gorilla/pat.git", "github.com/gorilla/pat"},
				{"http://github.com/gorilla/mux", "github.com/gorilla/mux"},
				{"http://github.com/gorilla/pat", "github.com/gorilla/pat"},
			},
		},
		{
			name: "https_remote_url",
			in:   "github.com/!https://github.com/%.git",
			mappings: [][2]string{
				{"https://github.com/gorilla/mux.git", "github.com/gorilla/mux"},
				{"https://github.com/gorilla/pat.git", "github.com/gorilla/pat"},
				{"https://github.com/gorilla/mux", "github.com/gorilla/mux"},
				{"https://github.com/gorilla/pat", "github.com/gorilla/pat"},
			},
		},
		{
			name: "ssh_remote_url",
			in:   "github.com/!ssh://git@github.com:%.git",
			mappings: [][2]string{
				{"ssh://git@github.com:gorilla/mux.git", "github.com/gorilla/mux"},
				{"ssh://git@github.com:gorilla/pat.git", "github.com/gorilla/pat"},
				{"ssh://git@github.com:gorilla/mux", "github.com/gorilla/mux"},
				{"ssh://git@github.com:gorilla/pat", "github.com/gorilla/pat"},
			},
		},
		{
			name: "git_remote_url",
			in:   "github.com/!git://git@github.com:%.git",
			mappings: [][2]string{
				{"git://git@github.com:gorilla/mux.git", "github.com/gorilla/mux"},
				{"git://git@github.com:gorilla/pat.git", "github.com/gorilla/pat"},
				{"git://git@github.com:gorilla/mux", "github.com/gorilla/mux"},
				{"git://git@github.com:gorilla/pat", "github.com/gorilla/pat"},
			},
		},
		{
			name: "git_no_scheme_remote_url",
			in:   "github.com/!git@github.com:%.git",
			mappings: [][2]string{
				{"git@github.com:gorilla/mux.git", "github.com/gorilla/mux"},
				{"git@github.com:gorilla/pat.git", "github.com/gorilla/pat"},
				{"git@github.com:gorilla/mux", "github.com/gorilla/mux"},
				{"git@github.com:gorilla/pat", "github.com/gorilla/pat"},
			},
		},
	}

	restoreOriginMap := originMap[:]
	defer func() {
		originMap = restoreOriginMap
	}()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var err error
			originMap, err = parse(test.in)
			if err != nil {
				t.Fatalf("on input %q, unexpected err: %v", test.in, err)
			}
			for _, mapping := range test.mappings {
				if gotRepo := reverse(mapping[0]); gotRepo != mapping[1] {
					t.Errorf("on input %q, input clone URL %q, got %q, but expected %q", test.in, mapping[0], gotRepo, mapping[1])
				}
			}
		})
	}
}
