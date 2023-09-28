package internal

import (
	"context"
	"os/exec"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
)

// VCSSyncer describes whether and how to sync content from a VCS remote to
// local disk.
type VCSSyncer interface {
	// Type returns the type of the syncer.
	Type() string
	// IsCloneable checks to see if the VCS remote URL is cloneable. Any non-nil
	// error indicates there is a problem.
	IsCloneable(ctx context.Context, repoName api.RepoName, remoteURL *vcs.URL) error
	// CloneCommand returns the command to be executed for cloning from remote.
	CloneCommand(ctx context.Context, remoteURL *vcs.URL, tmpPath string) (cmd *exec.Cmd, err error)
	// Fetch tries to fetch updates from the remote to given directory.
	// The revspec parameter is optional and specifies that the client is specifically
	// interested in fetching the provided revspec (example "v2.3.4^0").
	// For package hosts (vcsPackagesSyncer, npm/pypi/crates.io), the revspec is used
	// to lazily fetch package versions. More details at
	// https://github.com/sourcegraph/sourcegraph/issues/37921#issuecomment-1184301885
	// Beware that the revspec parameter can be any random user-provided string.
	Fetch(ctx context.Context, remoteURL *vcs.URL, repoName api.RepoName, dir common.GitDir, revspec string) ([]byte, error)
	// RemoteShowCommand returns the command to be executed for showing remote.
	RemoteShowCommand(ctx context.Context, remoteURL *vcs.URL) (cmd *exec.Cmd, err error)
}

type notFoundError struct{ error }

func (e notFoundError) NotFound() bool { return true }
