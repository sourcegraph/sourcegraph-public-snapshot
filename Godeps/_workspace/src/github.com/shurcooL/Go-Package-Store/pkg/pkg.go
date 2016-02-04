package pkg

import (
	"github.com/shurcooL/vcsstate"
	"golang.org/x/tools/go/vcs"
)

type Repo struct {
	// Path is the local filesystem path to the repository.
	Path string

	// Root is the import path corresponding to the root of the repository.
	Root string

	// RemoteURL is the remote URL, including scheme.
	// TODO: Consider moving/renaming to Remote.RepoURL.
	RemoteURL string

	// TODO: Consider.
	Cmd *vcs.Cmd

	// VCS allows getting the state of the VCS.
	VCS vcsstate.VCS

	// RemoteVCS allows getting the remote state of the VCS.
	RemoteVCS vcsstate.RemoteVCS

	Local struct {
		Revision string
	}
	Remote struct {
		Revision string
	}
}

// ImportPathPattern returns an import path pattern that matches all of the Go packages in this repo.
// E.g.,
//
// 	"github.com/owner/repo/..."
func (r Repo) ImportPathPattern() string {
	return r.Root + "/..."
}
