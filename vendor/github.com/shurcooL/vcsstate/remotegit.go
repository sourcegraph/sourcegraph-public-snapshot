package vcsstate

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/shurcooL/go/osutil"
)

type remoteGit struct{}

func (v remoteGit) RemoteRevision(remoteURL string) (string, error) {
	// true here is not a boolean value, but a command /bin/true that will make git think it asked for a password,
	// and prevent potential interactive password prompts (opting to return failure exit code instead).
	cmd := exec.Command("git", "-c", "core.askpass=true", "ls-remote", "--heads", remoteURL, v.defaultBranch())
	env := osutil.Environ(os.Environ())
	env.Set("GIT_SSH_COMMAND", "ssh -o StrictHostKeyChecking=yes") // Default for StrictHostKeyChecking is "ask", which we don't want since this is non-interactive and we prefer to fail than block asking for user input.
	cmd.Env = env

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	if len(out) < gitRevisionLength {
		return "", fmt.Errorf("output length %v is shorter than %v", len(out), gitRevisionLength)
	}
	return string(out[:gitRevisionLength]), nil
}

func (remoteGit) defaultBranch() string {
	return "master"
}
