package github

import "sourcegraph.com/sourcegraph/sourcegraph/store"

// RepoOrigin is an implementation of the RepoOrigin store for
// mirrored repos whose external origin is GitHub.
type RepoOrigin struct{}

var _ store.RepoOrigin = (*RepoOrigin)(nil)

func (s *RepoOrigin) BrandName() string { return "GitHub" }

func (s *RepoOrigin) Host() string { return "github.com" }
