package graphqlbackend

import (
	"context"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

type gitRevSpecExpr struct {
	expr string
	oid  GitObjectID
	repo *RepositoryResolver
}

func (r *gitRevSpecExpr) Expr() string { return r.expr }

func (r *gitRevSpecExpr) Object(ctx context.Context) (*gitObject, error) {
	if r.oid != "" {
		// Precomputed.
		return &gitObject{oid: r.oid, repo: r.repo}, nil
	}

	cachedRepo, err := backend.CachedGitRepo(ctx, r.repo.repo)
	if err != nil {
		return nil, err
	}
	oid, err := git.ResolveRevision(ctx, *cachedRepo, nil, r.expr, nil)
	if err != nil {
		return nil, err
	}
	return &gitObject{
		oid:  GitObjectID(oid),
		repo: r.repo,
	}, nil
}

type gitRevSpec struct {
	ref    *GitRefResolver
	expr   *gitRevSpecExpr
	object *gitObject
}

func (r *gitRevSpec) ToGitRef() (*GitRefResolver, bool)         { return r.ref, r.ref != nil }
func (r *gitRevSpec) ToGitRevSpecExpr() (*gitRevSpecExpr, bool) { return r.expr, r.expr != nil }
func (r *gitRevSpec) ToGitObject() (*gitObject, bool)           { return r.object, r.object != nil }

// GitRevisionRange implements the GitRevisionRange GraphQL type.
type GitRevisionRange interface {
	Expr() string
	Base() *gitRevSpec
	BaseRevSpec() *gitRevSpecExpr
	Head() *gitRevSpec
	HeadRevSpec() *gitRevSpecExpr
	MergeBase() *gitObject
}

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

// escapeRevspecForURL escapes revspec for use in a Sourcegraph URL. For niceness/readability, we do
// NOT escape slashes but we do escape other characters like '#' that are necessary for correctness.
func escapeRevspecForURL(revspec string) string {
	return strings.Replace(url.PathEscape(revspec), "%2F", "/", -1)
}
