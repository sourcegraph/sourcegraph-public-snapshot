package server

import (
	"context"
	"os/exec"

	"github.com/sourcegraph/sourcegraph/internal/vcs"
)

// VCSSyncer describes whether and how to sync content from a VCS remote to
// local disk.
type VCSSyncer interface {
	// Type returns the type of the syncer.
	Type() string
	// IsCloneable checks to see if the VCS remote URL is cloneable. Any non-nil
	// error indicates there is a problem.
	IsCloneable(ctx context.Context, remoteURL *vcs.URL) error
	// CloneCommand returns the command to be executed for cloning from remote.
	CloneCommand(ctx context.Context, remoteURL *vcs.URL, tmpPath string) (cmd *exec.Cmd, err error)
	// Fetch tries to fetch updates from the remote to given directory.
	Fetch(ctx context.Context, remoteURL *vcs.URL, dir GitDir) error
	// RemoteShowCommand returns the command to be executed for showing remote.
	RemoteShowCommand(ctx context.Context, remoteURL *vcs.URL) (cmd *exec.Cmd, err error)
}
