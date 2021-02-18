package server

import (
	"fmt"
	"net/url"
	"os/exec"
)

// GitCredentials are used to authenticate a git command by setting
// the GIT_ASKPASS environment variable to a short shell script that
// implements the read-only custom git credentials helper protocol.
// See https://git-scm.com/docs/gitcredentials#_custom_helpers
type GitCredentials struct {
	protocol string
	host     string
	username string
	password string
}

// Authenticate sets the GIT_ASKPASS environment variable that authenticates
// the git command with credentials.
func (c *GitCredentials) Authenticate(cmd *exec.Cmd) {
	cmd.Env = append(cmd.Env, "GIT_ASKPASS="+c.Command())
}

// Command returns the git credentials helper command that will echo back the credentials
// for a given repository's remote URL.
func (c *GitCredentials) Command() string {
	return fmt.Sprintf(
		`!f() { test "$1" = get && printf "protocol=%s\nhost=%s\nusername=%s\npassword=%s"; }; f`,
		c.protocol,
		c.host,
		c.username,
		c.password,
	)
}

// ParseGitCredentials instantiates a new GitCredentials struct from
// an authenticated remote URL, returning those credentials as well as
// the bare redacted cloneURL to be used instead of the original one.
func ParseGitCredentials(remoteURL *url.URL) (*GitCredentials, *url.URL) {
	creds := &GitCredentials{
		protocol: remoteURL.Scheme,
		host:     remoteURL.Host,
	}

	if remoteURL.User != nil {
		creds.username = remoteURL.User.Username()
		creds.password, _ = remoteURL.User.Password()
	}

	redacted := *remoteURL
	redacted.User = nil

	return creds, &redacted
}
