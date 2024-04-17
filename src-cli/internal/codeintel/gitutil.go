package codeintel

import (
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// InferRepo gets a Sourcegraph-friendly repo name from the git clone enclosing the working dir.
func InferRepo() (string, error) {
	remoteURL, err := runGitCommand("remote", "get-url", "origin")
	if err != nil {
		return "", err
	}

	return parseRemote(remoteURL)
}

// parseRemote converts a git origin url into a Sourcegraph-friendly repo name.
func parseRemote(remoteURL string) (string, error) {
	// e.g., git@github.com:sourcegraph/src-cli.git
	if strings.HasPrefix(remoteURL, "git@") {
		if parts := strings.Split(remoteURL, ":"); len(parts) == 2 {
			return strings.Join([]string{
				strings.TrimPrefix(parts[0], "git@"),
				strings.TrimSuffix(parts[1], ".git"),
			}, "/"), nil
		}
	}

	// e.g., https://github.com/sourcegraph/src-cli.git
	if url, err := url.Parse(remoteURL); err == nil {
		return url.Hostname() + strings.TrimSuffix(url.Path, ".git"), nil
	}

	return "", errors.Newf("unrecognized remote URL: %s", remoteURL)
}

// InferCommit gets a 40-character rev hash from the git clone enclosing the working dir.
func InferCommit() (string, error) {
	return runGitCommand("rev-parse", "HEAD")
}

// InferRoot gets the path relative to the root of the git clone enclosing the given file path.
func InferRoot(file string) (string, error) {
	topLevel, err := runGitCommand("rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}

	absoluteFile, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}

	relative, err := filepath.Rel(topLevel, absoluteFile)
	if err != nil {
		return "", err
	}

	return filepath.Dir(relative), nil
}

// runGitCommand runs a git command and trims all leading/trailing whitespace from the output.
func runGitCommand(args ...string) (string, error) {
	output, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		return "", errors.Newf("failed to run git command: %s\n%s", err, output)
	}

	return strings.TrimSpace(string(output)), nil
}
