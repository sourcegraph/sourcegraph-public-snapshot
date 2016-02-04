package vcsstate

import (
	"os/exec"
	"strings"

	"github.com/shurcooL/go/trim"
)

type remoteHg struct{}

func (v remoteHg) RemoteRevision(remoteURL string) (string, error) {
	cmd := exec.Command("hg", "--debug", "identify", "-i", "--rev", v.defaultBranch(), remoteURL)

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	// Get the last line of output.
	lines := strings.Split(trim.LastNewline(string(out)), "\n") // lines will always contain at least one element.
	return lines[len(lines)-1], nil
}

func (remoteHg) defaultBranch() string {
	return "default"
}
