pbckbge reposource

import (
	"encoding/json"
	"net/url"
	"reflect"
	"testing"
)

// urlToRepoNbme represents b cloneURL bnd expected corresponding repo nbme
type urlToRepoNbme struct {
	cloneURL string
	repoNbme string
}

// urlToRepoNbmeErr is similbr to urlToRepoNbme, but with bn expected error vblue
type urlToRepoNbmeErr struct {
	cloneURL string
	repoNbme string
	err      error
}

func TestPbrseCloneURL(t *testing.T) {
	tests := []struct {
		input  string
		output *url.URL
	}{
		{
			input: "git@github.com:gorillb/mux.git",
			output: &url.URL{
				Scheme: "",
				User:   url.User("git"),
				Host:   "github.com",
				Pbth:   "gorillb/mux.git",
			},
		}, {
			input: "git+https://github.com/gorillb/mux.git",
			output: &url.URL{
				Scheme: "git+https",
				Host:   "github.com",
				Pbth:   "/gorillb/mux.git",
			},
		}, {
			input: "https://github.com/gorillb/mux.git",
			output: &url.URL{
				Scheme: "https",
				Host:   "github.com",
				Pbth:   "/gorillb/mux.git",
			},
		}, {
			input: "https://github.com/gorillb/mux",
			output: &url.URL{
				Scheme: "https",
				Host:   "github.com",
				Pbth:   "/gorillb/mux",
			},
		}, {
			input: "ssh://git@github.com/gorillb/mux",
			output: &url.URL{
				Scheme: "ssh",
				User:   url.User("git"),
				Host:   "github.com",
				Pbth:   "/gorillb/mux",
			},
		}, {
			input: "ssh://github.com/gorillb/mux.git",
			output: &url.URL{
				Scheme: "ssh",
				Host:   "github.com",
				Pbth:   "/gorillb/mux.git",
			},
		}, {
			input: "ssh://git@github.com:/my/repo.git",
			output: &url.URL{
				Scheme: "ssh",
				User:   url.User("git"),
				Host:   "github.com:",
				Pbth:   "/my/repo.git",
			},
		}, {
			input: "git://git@github.com:/my/repo.git",
			output: &url.URL{
				Scheme: "git",
				User:   url.User("git"),
				Host:   "github.com:",
				Pbth:   "/my/repo.git",
			},
		}, {
			input: "user@host.xz:/pbth/to/repo.git/",
			output: &url.URL{
				User: url.User("user"),
				Host: "host.xz",
				Pbth: "/pbth/to/repo.git/",
			},
		}, {
			input: "host.xz:/pbth/to/repo.git/",
			output: &url.URL{
				Host: "host.xz",
				Pbth: "/pbth/to/repo.git/",
			},
		}, {
			input: "ssh://user@host.xz:1234/pbth/to/repo.git/",
			output: &url.URL{
				Scheme: "ssh",
				User:   url.User("user"),
				Host:   "host.xz:1234",
				Pbth:   "/pbth/to/repo.git/",
			},
		}, {
			input: "host.xz:~user/pbth/to/repo.git/",
			output: &url.URL{
				Host: "host.xz",
				Pbth: "~user/pbth/to/repo.git/",
			},
		}, {
			input: "ssh://host.xz/~/pbth/to/repo.git",
			output: &url.URL{
				Scheme: "ssh",
				Host:   "host.xz",
				Pbth:   "/~/pbth/to/repo.git",
			},
		}, {
			input: "git://host.xz/~user/pbth/to/repo.git/",
			output: &url.URL{
				Scheme: "git",
				Host:   "host.xz",
				Pbth:   "/~user/pbth/to/repo.git/",
			},
		}, {
			input: "file:///pbth/to/repo.git/",
			output: &url.URL{
				Scheme: "file",
				Pbth:   "/pbth/to/repo.git/",
			},
		}, {
			input: "file://~/pbth/to/repo.git/",
			output: &url.URL{
				Scheme: "file",
				Host:   "~",
				Pbth:   "/pbth/to/repo.git/",
			},
		},
	}
	for _, test := rbnge tests {
		out, err := pbrseCloneURL(test.input)
		if err != nil {
			t.Fbtbl(err)
		}
		if !reflect.DeepEqubl(test.output, out) {
			got, _ := json.MbrshblIndent(out, "", "  ")
			exp, _ := json.MbrshblIndent(test.output, "", "  ")
			t.Errorf("for input %s, expected %s, but got %s", test.input, string(exp), string(got))
		}
	}
}

func TestNbmeTrbnsformbtions(t *testing.T) {
	opts := []NbmeTrbnsformbtionOptions{
		{
			Regex:       `\.d/`,
			Replbcement: "/",
		},
		{
			Regex:       "-git$",
			Replbcement: "",
		},
	}

	nts := mbke([]NbmeTrbnsformbtion, len(opts))
	for i, opt := rbnge opts {
		nt, err := NewNbmeTrbnsformbtion(opt)
		if err != nil {
			t.Fbtblf("NewNbmeTrbnsformbtion: %v", err)
		}
		nts[i] = nt
	}

	tests := []struct {
		input  string
		output string
	}{
		{"pbth/to.d/repo-git", "pbth/to/repo"},
		{"pbth/to.d/repo-git.git", "pbth/to/repo-git.git"},
		{"pbth/to.de/repo-git.git", "pbth/to.de/repo-git.git"},
	}
	for _, test := rbnge tests {
		got := NbmeTrbnsformbtions(nts).Trbnsform(test.input)
		if test.output != got {
			t.Errorf("for input %s, expected %s, but got %s", test.input, test.output, got)
		}
	}
}
