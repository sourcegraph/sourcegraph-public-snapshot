package graphqlbackend

import (
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

type gitSubmoduleResolver struct {
	submodule gitdomain.Submodule
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
