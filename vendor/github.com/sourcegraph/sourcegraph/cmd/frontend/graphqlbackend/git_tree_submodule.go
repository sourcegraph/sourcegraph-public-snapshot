package graphqlbackend

import "github.com/sourcegraph/sourcegraph/pkg/vcs/git"

type gitSubmoduleResolver struct {
	submodule git.Submodule
}

func (r *gitSubmoduleResolver) URL() string {
	return r.submodule.URL
}

func (r *gitSubmoduleResolver) Commit() string {
	return string(r.submodule.CommitID)
}

func (r *gitSubmoduleResolver) Path() string {
	return r.submodule.Path
}
