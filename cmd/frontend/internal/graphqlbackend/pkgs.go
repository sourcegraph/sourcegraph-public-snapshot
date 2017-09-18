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
}

func (r *packageResolver) Lang() string { return r.pkg.Lang }
func (r *packageResolver) ID() *string {
	if id, isStr := r.pkg.Pkg["id"].(string); isStr {
		return &id
	}
	return nil
}
func (r *packageResolver) Type() *string {
	if typ, isStr := r.pkg.Pkg["typ"].(string); isStr {
		return &typ
	}
	return nil
}
func (r *packageResolver) Name() *string {
	if name, isStr := r.pkg.Pkg["name"].(string); isStr {
		return &name
	}
	return nil
}
func (r *packageResolver) Commit() *string {
	if commit, isStr := r.pkg.Pkg["commit"].(string); isStr {
		return &commit
	}
	return nil
}
func (r *packageResolver) BaseDir() *string {
	if baseDir, isStr := r.pkg.Pkg["baseDir"].(string); isStr {
		return &baseDir
	}
	return nil
}
func (r *packageResolver) RepoURL() *string {
	if repoURL, isStr := r.pkg.Pkg["repoURL"].(string); isStr {
		return &repoURL
	}
	return nil
}
func (r *packageResolver) Version() *string {
	if version, isStr := r.pkg.Pkg["version"].(string); isStr {
		return &version
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
	return pkgQuery
}
