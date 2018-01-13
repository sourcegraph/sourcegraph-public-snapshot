package graphqlbackend

import "sourcegraph.com/sourcegraph/sourcegraph/schema"

type repoListConfigResolver struct {
	*schema.Repository
}

func (r *repoListConfigResolver) URL() string {
	return r.Repository.Url
}

func (r *repoListConfigResolver) Path() string {
	return r.Repository.Path
}

func (r *repoListConfigResolver) ViewURL() *string {
	u := r.Repository.ViewURL
	if u == "" {
		return nil
	}
	return &u
}

func (r *repoListConfigResolver) CommitURL() *string {
	u := r.Repository.CommitURL
	if u == "" {
		return nil
	}
	return &u
}

func (r *repoListConfigResolver) BlobURL() *string {
	u := r.Repository.BlobURL
	if u == "" {
		return nil
	}
	return &u
}

func (r *repoListConfigResolver) TreeURL() *string {
	u := r.Repository.TreeURL
	if u == "" {
		return nil
	}
	return &u
}
