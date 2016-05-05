package worker

import (
	"testing"

	"github.com/whilp/git-urls"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dockerutil"
)

func TestGitURLsParse(t *testing.T) {
	tests := map[string]string{
		"http://example.com":                  "http://@example.com",
		"git+ssh://example.com":               "git+ssh://@example.com",
		"git+ssh://u@example.com/my/repo.git": "git+ssh://u@example.com/my/repo.git",

		"u@example.com:/my/repo.git": "ssh://u@example.com/my/repo.git",
		"u@example.com:my/repo.git":  "ssh://u@example.com/my/repo.git",
		"example.com:my/repo.git":    "ssh://@example.com/my/repo.git",
	}
	for urlStr, want := range tests {
		u, err := giturls.Parse(urlStr)
		if err != nil {
			t.Errorf("%q: %s", urlStr, err)
			continue
		}
		if u.String() != want {
			t.Errorf("%q: got %q, want %q", urlStr, u, want)
			continue
		}
	}
}

func TestDroneRepoLink(t *testing.T) {
	tests := map[string]string{
		"http://example.com/my/repo":             "example.com/my/repo",
		"http://example.com/my/repo.git":         "example.com/my/repo",
		"http://u:p@example.com/my/repo":         "example.com/my/repo",
		"http://example.com:1234/my/repo":        "my/repo",
		"http://localhost/my/repo":               "my/repo",
		"http://127.0.0.1/my/repo":               "my/repo",
		"http://127.0.0.1:1234/my/repo":          "my/repo",
		"http://1.2.3.4/my/repo":                 "my/repo",
		"http://1.2.3.4:1234/my/repo":            "my/repo",
		"u@example.com:my/repo.git":              "example.com/my/repo",
		"u@example.com:/my/repo.git":             "example.com/my/repo",
		"example.com:/my/repo.git":               "example.com/my/repo",
		"git+ssh://user@example.com/my/repo.git": "example.com/my/repo",
	}
	for urlStr, want := range tests {
		u, err := giturls.Parse(urlStr)
		link, err := droneRepoLink(*u)
		if err != nil {
			t.Errorf("%q: %s", urlStr, err)
			continue
		}
		if link != want {
			t.Errorf("%q: got %q, want %q", urlStr, link, want)
			continue
		}
	}
}

func TestContainerAddrForHost(t *testing.T) {
	dockerutil.ContainerHost = func() (string, error) { return "0.0.0.0", nil } // mock, since CI may not have Docker

	containerHostname, err := dockerutil.ContainerHost()
	if err != nil {
		t.Fatal(err)
	}

	tests := map[string]string{
		"http://example.com/my/repo":             "http://@example.com/my/repo",
		"user@example.com:my/repo.git":           "ssh://user@example.com/my/repo.git",
		"example.com:/my/repo.git":               "ssh://@example.com/my/repo.git",
		"git+ssh://user@example.com/my/repo.git": "git+ssh://user@example.com/my/repo.git",

		"http://localhost/my/repo":           "http://@" + containerHostname + "/my/repo",
		"http://localhost:1234/my/repo":      "http://@" + containerHostname + ":1234/my/repo",
		"http://localhost:http/my/repo":      "http://@" + containerHostname + ":http/my/repo",
		"http://localhost/my/localhost/repo": "http://@" + containerHostname + "/my/localhost/repo",
		"u@localhost:my/repo.git":            "ssh://u@" + containerHostname + "/my/repo.git",
		"localhost:my/repo.git":              "ssh://@" + containerHostname + "/my/repo.git",
	}
	for urlStr, want := range tests {
		u, err := giturls.Parse(urlStr)
		if err != nil {
			t.Errorf("urlStr: %s, err: %s", urlStr, err)
			continue
		}
		_, containerCloneURL, err := containerAddrForHost(*u)
		if err != nil {
			t.Errorf("%q: %s", urlStr, err)
			continue
		}
		if containerCloneURL.String() != want {
			t.Errorf("%q: got %q, want %q", urlStr, containerCloneURL, want)
			continue
		}
	}
}
