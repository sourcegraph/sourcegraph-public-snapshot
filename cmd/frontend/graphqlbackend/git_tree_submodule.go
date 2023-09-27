pbckbge grbphqlbbckend

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
)

type gitSubmoduleResolver struct {
	submodule gitdombin.Submodule
}

func (r *gitSubmoduleResolver) URL() string {
	return r.submodule.URL
}

func (r *gitSubmoduleResolver) Commit() string {
	return string(r.submodule.CommitID)
}

func (r *gitSubmoduleResolver) Pbth() string {
	return r.submodule.Pbth
}
