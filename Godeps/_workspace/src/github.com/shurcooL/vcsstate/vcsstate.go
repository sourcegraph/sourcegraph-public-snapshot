// Package vcsstate allows getting the state of version control system repositories.
package vcsstate

import (
	"fmt"

	"golang.org/x/tools/go/vcs"
)

// VCS describes how to use a version control system to get the status of a repository
// rooted at dir.
type VCS interface {
	// DefaultBranch returns default branch name for this VCS type.
	DefaultBranch() string

	// Status returns the status of working directory.
	// It returns empty string if no outstanding status.
	Status(dir string) (string, error)

	// Branch returns the name of the locally checked out branch.
	Branch(dir string) (string, error)

	// LocalRevision returns current local revision of default branch.
	LocalRevision(dir string) (string, error)

	// Stash returns a non-empty string if the repository has a stash.
	Stash(dir string) (string, error)

	// Contains reports if the local default branch contains the commit specified by revision.
	Contains(dir string, revision string) (bool, error)

	// RemoteURL returns primary remote URL, as set in the local repository.
	RemoteURL(dir string) (string, error)

	// RemoteRevision returns latest remote revision of default branch.
	RemoteRevision(dir string) (string, error)
}

// NewVCS creates a VCS with same type as vcs.
func NewVCS(vcs *vcs.Cmd) (VCS, error) {
	switch vcs.Cmd {
	case "git":
		return git{}, nil
	case "hg":
		return hg{}, nil
	default:
		return nil, fmt.Errorf("unsupported vcs.Cmd: %v", vcs.Cmd)
	}
}

// RemoteVCS describes how to use a version control system to get the remote status of a repository
// with remoteURL.
type RemoteVCS interface {
	// RemoteRevision returns latest remote revision of default branch.
	RemoteRevision(remoteURL string) (string, error)
}

// NewRemoteVCS creates a RemoteVCS with same type as vcs.
func NewRemoteVCS(vcs *vcs.Cmd) (RemoteVCS, error) {
	switch vcs.Cmd {
	case "git":
		return remoteGit{}, nil
	case "hg":
		return remoteHg{}, nil
	default:
		return nil, fmt.Errorf("unsupported vcs.Cmd: %v", vcs.Cmd)
	}
}
