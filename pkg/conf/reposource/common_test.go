package reposource

import (
	"encoding/json"
	"net/url"
	"reflect"
	"testing"
)

// urlURI represents a cloneURL and expected corresponding repoURI
type urlURI struct {
	cloneURL string
	repoURI  string
}

func TestParseCloneURL(t *testing.T) {
	tests := []struct {
		input  string
		output *url.URL
	}{
		{
			input: "git@github.com/gorilla/mux.git",
			output: &url.URL{
				Scheme: "",
				User:   url.User("git"),
				Host:   "github.com",
				Path:   "/gorilla/mux.git",
			},
		}, {
			input: "https://github.com/gorilla/mux.git",
			output: &url.URL{
				Scheme: "https",
				Host:   "github.com",
				Path:   "/gorilla/mux.git",
			},
		}, {
			input: "https://github.com/gorilla/mux",
			output: &url.URL{
				Scheme: "https",
				Host:   "github.com",
				Path:   "/gorilla/mux",
			},
		}, {
			input: "ssh://git@github.com/gorilla/mux",
			output: &url.URL{
				Scheme: "ssh",
				User:   url.User("git"),
				Host:   "github.com",
				Path:   "/gorilla/mux",
			},
		}, {
			input: "ssh://github.com/gorilla/mux.git",
			output: &url.URL{
				Scheme: "ssh",
				Host:   "github.com",
				Path:   "/gorilla/mux.git",
			},
		},
	}
	for _, test := range tests {
		out, err := parseCloneURL(test.input)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(test.output, out) {
			got, _ := json.MarshalIndent(out, "", "  ")
			exp, _ := json.MarshalIndent(test.output, "", "  ")
			t.Errorf("for input %s, expected %s, but got %s", test.input, string(exp), string(got))
		}
	}
}
