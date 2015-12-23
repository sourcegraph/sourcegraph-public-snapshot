package changesets

import (
	"net/url"
	"os/exec"
	"regexp"
	"strings"

	"src.sourcegraph.com/sourcegraph/auth/userauth"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
)

// guessRepo uses the git remote origin url to guess the repo URI
func guessRepo() (string, error) {
	origin, err := gitRemoteOrigin()
	if err != nil {
		return "", err
	}
	if origin == nil {
		return "", nil
	}
	userAuth, err := userauth.Read(cli.Credentials.AuthFile)
	if err != nil {
		return "", err
	}
	return findEndpoint(origin, userAuth), nil
}

// findEndpoint takes a git remote url and saved sourcegraph endpoints to guess
// the repo URI for an endpoint
func findEndpoint(origin *url.URL, userAuth userauth.UserAuth) string {
	if origin.Host == "github.com" {
		// If the user's default upstream is github, they are likely
		// using changesets on a sourcegraph mirror.
		repo := "github.com/" + strings.TrimLeft(strings.TrimSuffix(origin.RequestURI(), ".git"), "/")
		return repo
	}
	// If we have an endpoint which matches, the repo is the path of the URI
	for endpoint := range userAuth {
		e, err := url.Parse(endpoint)
		if err != nil {
			continue
		}
		if e.Host == origin.Host {
			return strings.TrimLeft(origin.RequestURI(), "/")
		}
	}
	return ""
}

// scpSyntaxRe matches the SCP-like addresses used by Git to access
// repositories by SSH. Copied from `src/cmd/go/vcs.go`
var scpSyntaxRe = regexp.MustCompile(`^([a-zA-Z0-9_]+)@([a-zA-Z0-9._-]+):(.*)$`)

// gitRemoteOrigin returns a URL for the origin remote in the CWD
func gitRemoteOrigin() (*url.URL, error) {
	// based on code from src/cmd/go/vcs.go
	outb, err := exec.Command("git", "config", "remote.origin.url").Output()
	if err != nil {
		if outb != nil && len(outb) == 0 {
			return nil, nil
		}
		return nil, err
	}
	out := strings.TrimSpace(string(outb))

	var repoURL *url.URL
	if m := scpSyntaxRe.FindStringSubmatch(out); m != nil {
		// Match SCP-like syntax and convert it to a URL.
		// Eg, "git@github.com:user/repo" becomes
		// "ssh://git@github.com/user/repo".
		repoURL = &url.URL{
			Scheme: "ssh",
			User:   url.User(m[1]),
			Host:   m[2],
			Path:   "/" + m[3],
		}
	} else {
		repoURL, err = url.Parse(out)
		if err != nil {
			return nil, err
		}
	}
	return repoURL, nil
}
