package server

import (
	"context"
	"io"
	"os/exec"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server/common"
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
	// The revspec parameter is optional and specifies that the client is specifically
	// interested in fetching the provided revspec (example "v2.3.4^0").
	// For package hosts (vcsPackagesSyncer, npm/pypi/crates.io), the revspec is used
	// to lazily fetch package versions. More details at
	// https://github.com/sourcegraph/sourcegraph/issues/37921#issuecomment-1184301885
	// Beware that the revspec parameter can be any random user-provided string.
	Fetch(ctx context.Context, remoteURL *vcs.URL, dir common.GitDir, revspec string) ([]byte, error)
	// RemoteShowCommand returns the command to be executed for showing remote.
	RemoteShowCommand(ctx context.Context, remoteURL *vcs.URL) (cmd *exec.Cmd, err error)
}

// Cloner is an interface that optionally implements how we want all Syncers
// to work for Cloning. It is temporary until all Syncers implement it.
type Cloner interface {
	// Clone will clone remoteURL to tmpPath. The output of the clone commands
	// will be written to output which will be reported back to the user
	// (summarized to the last output line). The caller of Clone is
	// responsible for redaction of remoteURL secrets in output. (and in
	// err??)
	//
	// TODO maybe it should return output that has gone through a
	// progressWriter? Alternatively we ensure call sites to Clone use a
	// progressWriter on output. That might be the better approach since only
	// CloneCommand currently sets a non-nil output on runRemoteGitCommand
	Clone(ctx context.Context, remoteURL *vcs.URL, tmpPath string, output io.Writer) (err error)
}

type notFoundError struct{ error }

func (e notFoundError) NotFound() bool { return true }
