package graphqlbackend

import (
	"context"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

type packageResolver struct {
	pkg *sourcegraph.PackageInfo
}

type packageMetadata struct {
	id      *string
	typ     *string
	name    *string
	commit  *string
	baseDir *string
	repoURL *string
	version *string
	packag  *string
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

func (r *packageResolver) Repo(ctx context.Context) (*repositoryResolver, error) {
	repo, err := localstore.Repos.Get(ctx, r.pkg.RepoID)
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

func (r packageMetadata) toPkgQuery() map[string]interface{} {
	pkgQuery := make(map[string]interface{})
	if r.id != nil {
		pkgQuery["id"] = *r.id
	}
	if r.typ != nil {
		pkgQuery["type"] = *r.typ
	}
	if r.name != nil {
		pkgQuery["name"] = *r.name
	}
	if r.commit != nil {
		pkgQuery["commit"] = *r.commit
	}
	if r.baseDir != nil {
		pkgQuery["baseDir"] = *r.baseDir
	}
	if r.repoURL != nil {
		pkgQuery["repoURL"] = *r.repoURL
	}
	if r.version != nil {
		pkgQuery["version"] = *r.version
	}
	if r.packag != nil {
		pkgQuery["package"] = *r.packag
	}
	return pkgQuery
}
