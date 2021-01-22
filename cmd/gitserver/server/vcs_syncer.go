package server

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

// VCSSyncer describes whether and how to sync content from a VCS remote to
// local disk.
type VCSSyncer interface {
	// IsCloneable checks to see if the VCS remote URL is cloneable. Any non-nil
	// error indicates there is a problem.
	IsCloneable(ctx context.Context, url string) error
	// CloneCommand returns the command to be executed for cloning from remote.
	CloneCommand(ctx context.Context, url, tmpPath string) (cmd *exec.Cmd, err error)
	// FetchCommand returns the command to be executed for fetching updates from remote.
	FetchCommand(ctx context.Context, url string) (cmd *exec.Cmd, configRemoteOpts bool, err error)
	// RemoteShowCommand returns the command to be executed for showing remote.
	RemoteShowCommand(ctx context.Context, url string) (cmd *exec.Cmd, err error)
}

// GitRepoSyncer is a syncer for Git repositories.
type GitRepoSyncer struct{}

// IsCloneable checks to see if the Git remote URL is cloneable.
func (s *GitRepoSyncer) IsCloneable(ctx context.Context, url string) error {
	if strings.ToLower(string(protocol.NormalizeRepo(api.RepoName(url)))) == "github.com/sourcegraphtest/alwayscloningtest" {
		return nil
	}
	if testGitRepoExists != nil {
		return testGitRepoExists(ctx, url)
	}

	args := []string{"ls-remote", url, "HEAD"}
	ctx, cancel := context.WithTimeout(ctx, shortGitCommandTimeout(args))
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	out, err := runWithRemoteOpts(ctx, cmd, nil)
	if err != nil {
		if ctxerr := ctx.Err(); ctxerr != nil {
			err = ctxerr
		}
		if len(out) > 0 {
			err = fmt.Errorf("%s (output follows)\n\n%s", err, out)
		}
		return err
	}
	return nil
}

// CloneCommand returns the command to be executed for cloning a Git repository.
func (s *GitRepoSyncer) CloneCommand(ctx context.Context, url, tmpPath string) (cmd *exec.Cmd, err error) {
	if useRefspecOverrides() {
		return refspecOverridesCloneCmd(ctx, url, tmpPath)
	}
	return exec.CommandContext(ctx, "git", "clone", "--mirror", "--progress", url, tmpPath), nil
}

// FetchCommand returns the command to be executed for fetching updates of a Git repository.
func (s *GitRepoSyncer) FetchCommand(ctx context.Context, url string) (cmd *exec.Cmd, configRemoteOpts bool, err error) {
	configRemoteOpts = true
	if customCmd := customFetchCmd(ctx, url); customCmd != nil {
		cmd = customCmd
		configRemoteOpts = false
	} else if useRefspecOverrides() {
		cmd = refspecOverridesFetchCmd(ctx, url)
	} else {
		cmd = exec.CommandContext(ctx, "git", "fetch", "--prune", url,
			// Normal git refs
			"+refs/heads/*:refs/heads/*", "+refs/tags/*:refs/tags/*",
			// GitHub pull requests
			"+refs/pull/*:refs/pull/*",
			// GitLab merge requests
			"+refs/merge-requests/*:refs/merge-requests/*",
			// Bitbucket pull requests
			"+refs/pull-requests/*:refs/pull-requests/*",
			// Gerrit changesets
			"+refs/changes/*:refs/changes/*",
			// Possibly deprecated refs for sourcegraph zap experiment?
			"+refs/sourcegraph/*:refs/sourcegraph/*")
	}
	return cmd, configRemoteOpts, nil
}

// RemoteShowCommand returns the command to be executed for showing remote of a Git repository.
func (s *GitRepoSyncer) RemoteShowCommand(ctx context.Context, url string) (cmd *exec.Cmd, err error) {
	return exec.CommandContext(ctx, "git", "remote", "show", url), nil
}

// PerforceDepotSyncer is a syncer for Perforce depots.
type PerforceDepotSyncer struct{}

// decomposePerforceCloneURL decomposes information back from a clone URL for a
// Perforce depot.
func decomposePerforceCloneURL(cloneURL string) (username, password, host, depot string, err error) {
	url, err := url.Parse(cloneURL)
	if err != nil {
		return "", "", "", "", err
	}

	if url.Scheme != "perforce" {
		return "", "", "", "", errors.New(`scheme is not "perforce"`)
	}

	password, _ = url.User.Password()
	return url.User.Username(), password, url.Host, url.Path, nil
}

// IsCloneable checks to see if the Perforce remote URL is cloneable.
func (s *PerforceDepotSyncer) IsCloneable(ctx context.Context, url string) error {
	username, password, host, _, err := decomposePerforceCloneURL(url)
	if err != nil {
		return errors.Wrap(err, "decompose")
	}
	_ = password // TODO

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// FIXME: Need to find a way to determine if depot exists instead of a general ping to the Perforce server.
	cmd := exec.CommandContext(ctx, "p4", "ping", "-c", "1")
	cmd.Env = append(os.Environ(),
		"P4PORT="+host,
		"P4USER="+username,
	)
	out, err := runWith(ctx, cmd, false, nil)
	fmt.Println("out:", string(out)) // TODO: Delete me
	if err != nil {
		if ctxerr := ctx.Err(); ctxerr != nil {
			err = ctxerr
		}
		if len(out) > 0 {
			err = fmt.Errorf("%s (output follows)\n\n%s", err, out)
		}
		return err
	}
	return nil
}

// CloneCommand returns the command to be executed for cloning a Perforce depot as a Git repository.
func (s *PerforceDepotSyncer) CloneCommand(ctx context.Context, url, tmpPath string) (*exec.Cmd, error) {
	username, password, host, depot, err := decomposePerforceCloneURL(url)
	if err != nil {
		return nil, errors.Wrap(err, "decompose")
	}
	_ = password // TODO

	cmd := exec.CommandContext(ctx, "git", "p4", "clone", "--bare", depot+"@all", tmpPath)
	cmd.Env = append(os.Environ(),
		"P4PORT="+host,
		"P4USER="+username,
	)

	return cmd, nil
}

// FetchCommand returns the command to be executed for fetching updates of a Perforce depot as a Git repository.
func (s *PerforceDepotSyncer) FetchCommand(ctx context.Context, url string) (cmd *exec.Cmd, configRemoteOpts bool, err error) {
	username, password, host, _, err := decomposePerforceCloneURL(url)
	if err != nil {
		return nil, false, errors.Wrap(err, "decompose")
	}
	_ = password // TODO

	cmd = exec.CommandContext(ctx, "git", "p4", "sync")
	cmd.Env = append(os.Environ(),
		"P4PORT="+host,
		"P4USER="+username,
	)

	return cmd, false, nil
}

// RemoteShowCommand returns the command to be executed for showing Git remote of a Perforce depot.
func (s *PerforceDepotSyncer) RemoteShowCommand(ctx context.Context, url string) (cmd *exec.Cmd, err error) {
	// Remote info is encoded as in the current repository
	return exec.CommandContext(ctx, "git", "remote", "show", "./"), nil
}
