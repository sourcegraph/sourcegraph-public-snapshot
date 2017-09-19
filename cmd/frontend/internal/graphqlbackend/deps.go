package graphqlbackend

import (
	"context"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

type dependencyResolver struct {
	dep *sourcegraph.DependencyReference
}

func (r *dependencyResolver) Name() *string {
	return r.depDataStringField("name")
}
func (r *dependencyResolver) RepoURL() *string {
	return r.depDataStringField("repoURL")
}
func (r *dependencyResolver) Depth() *int32 {
	return r.depDataIntField("depth")
}
func (r *dependencyResolver) Vendor() *bool {
	return r.depDataBoolField("vendor")
}
func (r *dependencyResolver) Package() *string {
	return r.depDataStringField("package")
}
func (r *dependencyResolver) Absolute() *string {
	return r.depDataStringField("absolute")
}
func (r *dependencyResolver) Type() *string {
	return r.depDataStringField("type")
}
func (r *dependencyResolver) Commit() *string {
	return r.depDataStringField("commit")
}
func (r *dependencyResolver) Version() *string {
	return r.depDataStringField("version")
}
func (r *dependencyResolver) ID() *string {
	return r.depDataStringField("id")
}

func (r *dependencyResolver) depDataStringField(name string) *string {
	if value, isStr := r.dep.DepData[name].(string); isStr {
		return &value
	}
	return nil
}

func (r *dependencyResolver) depDataBoolField(name string) *bool {
	if value, isStr := r.dep.DepData[name].(bool); isStr {
		return &value
	}
	return nil
}

func (r *dependencyResolver) depDataIntField(name string) *int32 {
	if value, isStr := r.dep.DepData[name].(int32); isStr {
		return &value
	}
	return nil
}

func (r *dependencyResolver) Repo(ctx context.Context) (*repositoryResolver, error) {
	repo, err := localstore.Repos.Get(ctx, r.dep.RepoID)
	if err != nil {
		if err, ok := err.(legacyerr.Error); ok && err.Code == legacyerr.NotFound {
			return nil, nil
		}
		return nil, err
	}

	if err := refreshRepo(ctx, repo); err != nil {
		return nil, err
	}

	return &repositoryResolver{repo: repo}, nil
}
