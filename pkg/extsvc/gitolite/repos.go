//go:generate mockgen -destination mock_gitolite/mocks.go github.com/sourcegraph/sourcegraph/pkg/extsvc/gitolite Command
package gitolite

import (
	"context"
	"os/exec"
	"strings"

	log15 "gopkg.in/inconshreveable/log15.v2"
)

// Repo is the repository metadata returned by the Gitolite API.
type Repo struct {
	// Name is the name of the repository as it is returned by `ssh git@GITOLITE_HOST info`
	Name string

	// URL is the clone URL of the repository.
	URL string
}

// Client is a client for the Gitolite API.
//
// IMPORTANT: in order to authenticate to the Gitolite API, the client must be invoked from a
// service in an environment that contains a Gitolite-authorized SSH key. As of writing, only
// gitserver meets this criterion. (I.e., only invoke this from gitserver.)
type Client struct {
	Host string

	command Command
}

func NewClient(host string) *Client {
	return &Client{
		Host:    host,
		command: command{},
	}
}

func (c *Client) ListRepos(ctx context.Context) ([]*Repo, error) {
	out, err := c.command.Output(ctx, "ssh", c.Host, "info")
	if err != nil {
		log15.Error("listing gitolite failed", "error", err, "out", string(out))
		return nil, err
	}

	lines := strings.Split(string(out), "\n")
	var repos []*Repo
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] != "R" {
			continue
		}
		name := fields[len(fields)-1]
		if len(fields) >= 2 && fields[0] == "R" {
			repos = append(repos, &Repo{
				Name: name,
				URL:  c.Host + ":" + name,
			})
		}
	}

	return repos, nil
}

// Command is an interface for invoking a shell command.
type Command interface {
	Output(ctx context.Context, name string, arg ...string) ([]byte, error)
}

// command is the default implementation of Command, which uses `exec.CommandContext`.
type command struct{}

func (c command) Output(ctx context.Context, name string, arg ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, arg...).Output()
}
