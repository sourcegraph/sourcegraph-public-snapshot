package server

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

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

// PerforceDepotSyncer is a syncer for Perforce depots.
type PerforceDepotSyncer struct{}

// TODO(jchen)
func (s *PerforceDepotSyncer) IsCloneable(ctx context.Context, url string) error {
	return errors.New("not yet implemented") // p4 ping
}

// TODO(jchen)
func (s *PerforceDepotSyncer) CloneCommand(ctx context.Context, url, tmpPath string) (cmd *exec.Cmd, err error) {
	return nil, errors.New("not yet implemented") // git p4 clone
}

// TODO(jchen)
func (s *PerforceDepotSyncer) FetchCommand(ctx context.Context, url string) (cmd *exec.Cmd, configRemoteOpts bool, err error) {
	return nil, false, errors.New("not yet implemented") // git p4 sync
}
