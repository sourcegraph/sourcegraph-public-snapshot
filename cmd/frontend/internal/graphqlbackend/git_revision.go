package graphqlbackend

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
)

type gitRevSpecExpr struct {
	expr string
	repo *repositoryResolver
}

func (r *gitRevSpecExpr) Expr() string { return r.expr }

func (r *gitRevSpecExpr) Object(ctx context.Context) (*gitObject, error) {
	vcsrepo := backend.Repos.CachedVCS(r.repo.repo)
	oid, err := vcsrepo.ResolveRevision(ctx, r.expr, nil)
	if err != nil {
		return nil, err
	}
	return &gitObject{
		oid:  gitObjectID(oid),
		repo: r.repo,
	}, nil
}

type gitRevSpec struct {
	ref    *gitRefResolver
	expr   *gitRevSpecExpr
	object *gitObject
}

func (r *gitRevSpec) ToGitRef() (*gitRefResolver, bool)         { return r.ref, r.ref != nil }
func (r *gitRevSpec) ToGitRevSpecExpr() (*gitRevSpecExpr, bool) { return r.expr, r.expr != nil }
func (r *gitRevSpec) ToGitObject() (*gitObject, bool)           { return r.object, r.object != nil }

type gitRevisionRange struct {
	expr       string
	base, head *gitRevSpec
	mergeBase  *gitObject
}

func (r *gitRevisionRange) Expr() string      { return r.expr }
func (r *gitRevisionRange) Base() *gitRevSpec { return r.base }
func (r *gitRevisionRange) BaseRevSpec() *gitRevSpecExpr {
	expr, _ := r.base.ToGitRevSpecExpr()
	return expr
}
func (r *gitRevisionRange) Head() *gitRevSpec { return r.head }
func (r *gitRevisionRange) HeadRevSpec() *gitRevSpecExpr {
	expr, _ := r.head.ToGitRevSpecExpr()
	return expr
}
func (r *gitRevisionRange) MergeBase() *gitObject { return r.mergeBase }
