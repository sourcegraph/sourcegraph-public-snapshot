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
	if name, isStr := r.dep.DepData["name"].(string); isStr {
		return &name
	}
	return nil
}
func (r *dependencyResolver) RepoURL() *string {
	if repoURL, isStr := r.dep.DepData["repoURL"].(string); isStr {
		return &repoURL
	}
	return nil
}
func (r *dependencyResolver) Depth() *int32 {
	if depth, isInt := r.dep.DepData["depth"].(int32); isInt {
		return &depth
	}
	return nil
}
func (r *dependencyResolver) Vendor() *bool {
	if vendor, isBool := r.dep.DepData["vendor"].(bool); isBool {
		return &vendor
	}
	return nil
}
func (r *dependencyResolver) Package() *string {
	if pkg, isStr := r.dep.DepData["package"].(string); isStr {
		return &pkg
	}
	return nil
}
func (r *dependencyResolver) Absolute() *string {
	if absolute, isStr := r.dep.DepData["absolute"].(string); isStr {
		return &absolute
	}
	return nil
}
func (r *dependencyResolver) Type() *string {
	if typ, isStr := r.dep.DepData["type"].(string); isStr {
		return &typ
	}
	return nil
}
func (r *dependencyResolver) Commit() *string {
	if commit, isStr := r.dep.DepData["commit"].(string); isStr {
		return &commit
	}
	return nil
}
func (r *dependencyResolver) Version() *string {
	if version, isStr := r.dep.DepData["version"].(string); isStr {
		return &version
	}
	return nil
}
func (r *dependencyResolver) ID() *string {
	if id, isStr := r.dep.DepData["id"].(string); isStr {
		return &id
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
