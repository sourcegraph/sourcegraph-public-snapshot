package server

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GitCommandError is an error of a failed Git command.
type GitCommandError struct {
	// Err is the original error produced by the git command that failed.
	Err error
	// Output is the std error output of the command that failed.
	Output string
}

func (e *GitCommandError) Error() string {
	return fmt.Sprintf("%s - output: %q", e.Err, e.Output)
}

func (e *GitCommandError) Unwrap() error {
	return e.Err
}

// GitRepoSyncer is a syncer for Git repositories.
type GitRepoSyncer struct{}

func (s *GitRepoSyncer) Type() string {
	return "git"
}

// IsCloneable checks to see if the Git remote URL is cloneable.
func (s *GitRepoSyncer) IsCloneable(ctx context.Context, remoteURL *vcs.URL) error {
	if isAlwaysCloningTestRemoteURL(remoteURL) {
		return nil
	}
	if testGitRepoExists != nil {
		return testGitRepoExists(ctx, remoteURL)
	}

	args := []string{"ls-remote", remoteURL.String(), "HEAD"}
	ctx, cancel := context.WithTimeout(ctx, shortGitCommandTimeout(args))
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	out, err := runWith(ctx, wrexec.Wrap(ctx, log.NoOp(), cmd), true, nil)
	if err != nil {
		if ctxerr := ctx.Err(); ctxerr != nil {
			err = ctxerr
		}
		if len(out) > 0 {
			err = &GitCommandError{Err: err, Output: string(out)}
		}
		return err
	}
	return nil
}

// CloneCommand returns the command to be executed for cloning a Git repository.
func (s *GitRepoSyncer) CloneCommand(ctx context.Context, remoteURL *vcs.URL, tmpPath string) (cmd *exec.Cmd, err error) {
	if err := os.MkdirAll(tmpPath, os.ModePerm); err != nil {
		return nil, errors.Wrapf(err, "clone failed to create tmp dir")
	}

	cmd = exec.CommandContext(ctx, "git", "init", "--bare", ".")
	cmd.Dir = tmpPath
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrapf(&GitCommandError{Err: err}, "clone setup failed")
	}

	cmd, _ = s.fetchCommand(ctx, remoteURL)
	cmd.Dir = tmpPath
	return cmd, nil
}

// Fetch tries to fetch updates of a Git repository.
func (s *GitRepoSyncer) Fetch(ctx context.Context, remoteURL *vcs.URL, dir GitDir, revspec string) error {
	cmd, configRemoteOpts := s.fetchCommand(ctx, remoteURL)
	dir.Set(cmd)
	if output, err := runWith(ctx, wrexec.Wrap(ctx, log.NoOp(), cmd), configRemoteOpts, nil); err != nil {
		return &GitCommandError{Err: err, Output: newURLRedactor(remoteURL).redact(string(output))}
	}
	return nil
}

// RemoteShowCommand returns the command to be executed for showing remote of a Git repository.
func (s *GitRepoSyncer) RemoteShowCommand(ctx context.Context, remoteURL *vcs.URL) (cmd *exec.Cmd, err error) {
	return exec.CommandContext(ctx, "git", "remote", "show", remoteURL.String()), nil
}

func (s *GitRepoSyncer) fetchCommand(ctx context.Context, remoteURL *vcs.URL) (cmd *exec.Cmd, configRemoteOpts bool) {
	configRemoteOpts = true
	if customCmd := customFetchCmd(ctx, remoteURL); customCmd != nil {
		cmd = customCmd
		configRemoteOpts = false
	} else if useRefspecOverrides() {
		cmd = refspecOverridesFetchCmd(ctx, remoteURL)
	} else {
		cmd = exec.CommandContext(ctx, "git", "fetch",
			"--progress", "--prune", remoteURL.String(),
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
	return cmd, configRemoteOpts
}
