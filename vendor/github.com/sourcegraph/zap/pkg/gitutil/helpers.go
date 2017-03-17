package gitutil

import (
	"fmt"
	"net/url"
	"strings"
)

// CurrentBranch returns the name of the current branch (i.e., the
// value of the HEAD symbolic ref) in a Git repository.
func CurrentBranch(gitRepo interface {
	ReadSymbolicRef(string) (string, error)
}) (string, error) {
	v, err := gitRepo.ReadSymbolicRef("HEAD")
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(v, "refs/heads/") {
		return v, nil
	}
	return "", fmt.Errorf("invalid HEAD %q", v)
}

// DefaultRepoName derives and returns a default repository name for the given
// remote URL. For example:
//
// 	DefaultRepoName("git@github.com:gorilla/mux") == "github.com/gorilla/mux"
// 	DefaultRepoName("https://github.com/gorilla/mux") == "github.com/gorilla/mux"
//  ... etc ...
//
func DefaultRepoName(remoteURL string) (string, error) {
	switch {
	case strings.HasPrefix(remoteURL, "http") || strings.HasPrefix(remoteURL, "https"):
		u, err := url.Parse(remoteURL)
		if err != nil {
			return "", err
		}
		return u.Host + strings.TrimSuffix(u.Path, ".git"), nil

	case strings.HasPrefix(remoteURL, "git@") || strings.HasPrefix(remoteURL, "git://") || strings.HasPrefix(remoteURL, "ssh://"):
		if !strings.HasPrefix(remoteURL, "git://") && !strings.HasPrefix(remoteURL, "ssh://") {
			remoteURL = "git://" + remoteURL
		}
		u, err := url.Parse(remoteURL)
		if err != nil {
			return "", err
		}
		s := strings.Split(u.Host, ":")
		if len(s) != 2 {
			return "", fmt.Errorf("cannot derive repository name from git remote %q", remoteURL)
		}
		return s[0] + "/" + s[1] + strings.TrimSuffix(u.Path, ".git"), nil

	default:
		// Default to the remote clone URL itself, as e.g. `zap remote` does when the
		// repository name is unspecified.
		return remoteURL, nil
	}
}
