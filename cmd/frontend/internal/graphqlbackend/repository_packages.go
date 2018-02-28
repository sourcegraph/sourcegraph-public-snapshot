package graphqlbackend

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
)

func (r *repositoryResolver) Packages(ctx context.Context) ([]*packageResolver, error) {
	pkgs, err := backend.Pkgs.ListPackages(ctx, &api.ListPackagesOp{RepoID: r.repo.ID})
	if err != nil {
		return nil, err
	}
	resolvers := make([]*packageResolver, len(pkgs))
	for i, pkg := range pkgs {
		resolvers[i] = &packageResolver{&pkg}
	}
	return resolvers, nil
}

type packageResolver struct {
	pkg *api.PackageInfo
}

func (r *packageResolver) Lang() string { return r.pkg.Lang }
func (r *packageResolver) ID() *string {
	return r.pkgStringField("id")
}
func (r *packageResolver) Type() *string {
	return r.pkgStringField("typ")
}
func (r *packageResolver) Name() *string {
	return r.pkgStringField("name")
}
func (r *packageResolver) Commit() *string {
	return r.pkgStringField("commit")
}
func (r *packageResolver) BaseDir() *string {
	return r.pkgStringField("baseDir")
}
func (r *packageResolver) RepoURL() *string {
	return r.pkgStringField("repoURL")
}
func (r *packageResolver) Version() *string {
	return r.pkgStringField("version")
}

func (r *packageResolver) pkgStringField(name string) *string {
	if value, isStr := r.pkg.Pkg[name].(string); isStr {
		return &value
	}
	return nil
}

func (r *packageResolver) Repository(ctx context.Context) (*repositoryResolver, error) {
	repo, err := db.Repos.Get(ctx, r.pkg.RepoID)
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if err := refreshRepo(ctx, repo); err != nil {
		return nil, err
	}
	return &repositoryResolver{repo: repo}, nil
}
