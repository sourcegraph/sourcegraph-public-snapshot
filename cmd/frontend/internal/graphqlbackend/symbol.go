package graphqlbackend

import "sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"

type symbolResolver struct {
	path      string
	line      int32
	character int32
	repo      *sourcegraph.Repo
}

func (r *symbolResolver) Path() string {
	return r.path
}

func (r *symbolResolver) Line() int32 {
	return r.line
}

func (r *symbolResolver) Character() int32 {
	return r.character
}

func (r *symbolResolver) Repository() *repositoryResolver {
	return &repositoryResolver{repo: r.repo}
}
