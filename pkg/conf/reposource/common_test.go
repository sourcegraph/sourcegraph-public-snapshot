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
			input: "git@github.com:gorilla/mux.git",
			output: &url.URL{
				Scheme: "",
				User:   url.User("git"),
				Host:   "github.com",
				Path:   "gorilla/mux.git",
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
		}, {
			input: "ssh://git@github.com:/my/repo.git",
			output: &url.URL{
				Scheme: "ssh",
				User:   url.User("git"),
				Host:   "github.com:",
				Path:   "/my/repo.git",
			},
		}, {
			input: "git://git@github.com:/my/repo.git",
			output: &url.URL{
				Scheme: "git",
				User:   url.User("git"),
				Host:   "github.com:",
				Path:   "/my/repo.git",
			},
		}, {
			input: "user@host.xz:/path/to/repo.git/",
			output: &url.URL{
				User: url.User("user"),
				Host: "host.xz",
				Path: "/path/to/repo.git/",
			},
		}, {
			input: "host.xz:/path/to/repo.git/",
			output: &url.URL{
				Host: "host.xz",
				Path: "/path/to/repo.git/",
			},
		}, {
			input: "ssh://user@host.xz:port/path/to/repo.git/",
			output: &url.URL{
				Scheme: "ssh",
				User:   url.User("user"),
				Host:   "host.xz:port",
				Path:   "/path/to/repo.git/",
			},
		}, {
			input: "host.xz:~user/path/to/repo.git/",
			output: &url.URL{
				Host: "host.xz",
				Path: "~user/path/to/repo.git/",
			},
		}, {
			input: "ssh://host.xz/~/path/to/repo.git",
			output: &url.URL{
				Scheme: "ssh",
				Host:   "host.xz",
				Path:   "/~/path/to/repo.git",
			},
		}, {
			input: "git://host.xz/~user/path/to/repo.git/",
			output: &url.URL{
				Scheme: "git",
				Host:   "host.xz",
				Path:   "/~user/path/to/repo.git/",
			},
		}, {
			input: "file:///path/to/repo.git/",
			output: &url.URL{
				Scheme: "file",
				Path:   "/path/to/repo.git/",
			},
		}, {
			input: "file://~/path/to/repo.git/",
			output: &url.URL{
				Scheme: "file",
				Host:   "~",
				Path:   "/path/to/repo.git/",
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
