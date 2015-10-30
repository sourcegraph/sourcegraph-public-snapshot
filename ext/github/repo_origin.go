package github

import (
	"src.sourcegraph.com/sourcegraph/ext/github/githubcli"
	"src.sourcegraph.com/sourcegraph/store"
)

// RepoOrigin is an implementation of the RepoOrigin store for
// mirrored repos whose external origin is GitHub.
type RepoOrigin struct{}

var _ store.RepoOrigin = (*RepoOrigin)(nil)

func (s *RepoOrigin) BrandName() string {
	if githubcli.Config.IsGitHubEnterprise() {
		return "GitHub (Enterprise)"
	} else {
		return "GitHub"
	}
}

func (s *RepoOrigin) Host() string { return githubcli.Config.Host() }
