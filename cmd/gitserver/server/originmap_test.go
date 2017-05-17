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
