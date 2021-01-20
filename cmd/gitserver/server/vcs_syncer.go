package server

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

// TODO
type VCSSyncer interface {
	// IsCloneable checks to see if the VCS remote URL is cloneable.
	IsCloneable(ctx context.Context, url string) error
	// TODO
	CloneCommand(ctx context.Context, url, tmpPath string) (cmd *exec.Cmd, err error)
	// TODO
	FetchCommand(ctx context.Context, url string) (cmd *exec.Cmd, configRemoteOpts bool, err error)
}

// GitRepoSyncer is a syncer for Git repositories.
type GitRepoSyncer struct{}

// IsCloneable checks to see if the Git remote URL is cloneable.
func (s *GitRepoSyncer) IsCloneable(ctx context.Context, url string) error {
	args := []string{"ls-remote", url, "HEAD"}
	ctx, cancel := context.WithTimeout(ctx, shortGitCommandTimeout(args))
	defer cancel()

	if strings.ToLower(string(protocol.NormalizeRepo(api.RepoName(url)))) == "github.com/sourcegraphtest/alwayscloningtest" {
		return nil
	}
	if testGitRepoExists != nil {
		return testGitRepoExists(ctx, url)
	}

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

// TODO
func (s *GitRepoSyncer) CloneCommand(ctx context.Context, url, tmpPath string) (cmd *exec.Cmd, err error) {
	if useRefspecOverrides() {
		return refspecOverridesCloneCmd(ctx, url, tmpPath)
	}
	return exec.CommandContext(ctx, "git", "clone", "--mirror", "--progress", url, tmpPath), nil
}

// TODO
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

// TODO
func (s *PerforceDepotSyncer) IsCloneable(ctx context.Context, url string) error {
	panic("implement me")
}

// TODO
func (s *PerforceDepotSyncer) CloneCommand(ctx context.Context, url, tmpPath string) (cmd *exec.Cmd, err error) {
	panic("implement me")
}

// TODO
func (s *PerforceDepotSyncer) FetchCommand(ctx context.Context, url string) (cmd *exec.Cmd, configRemoteOpts bool, err error) {
	panic("implement me")
}
