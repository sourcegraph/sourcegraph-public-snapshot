package graphqlbackend

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
)

type gitObjectID string

func (gitObjectID) ImplementsGraphQLType(name string) bool {
	return name == "GitObjectID"
}

func (id *gitObjectID) UnmarshalGraphQL(input interface{}) error {
	if input, ok := input.(string); ok && isValidGitObjectID(input) {
		*id = gitObjectID(input)
		return nil
	}
	return errors.New("GitObjectID: expected 40-character string (SHA-1 hash)")
}

func isValidGitObjectID(s string) bool {
	if len(s) != 40 {
		return false
	}
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			return false
		}
	}
	return true
}

func gitRefPrefix(ref string) string {
	if strings.HasPrefix(ref, "refs/heads/") {
		return "refs/heads/"
	}
	if strings.HasPrefix(ref, "refs/tags/") {
		return "refs/tags/"
	}
	if strings.HasPrefix(ref, "refs/pull/") {
		return "refs/pull/"
	}
	if strings.HasPrefix(ref, "refs/") {
		return "refs/"
	}
	return ""
}

func gitRefDisplayName(ref string) string {
	prefix := gitRefPrefix(ref)

	if prefix == "refs/pull/" && (strings.HasSuffix(ref, "/head") || strings.HasSuffix(ref, "/merge")) {
		// Special-case GitHub pull requests for a nicer display name.
		numberStr := ref[len(prefix) : len(prefix)+strings.Index(ref[len(prefix):], "/")]
		number, err := strconv.Atoi(numberStr)
		if err == nil {
			return fmt.Sprintf("#%d", number)
		}
	}

	return strings.TrimPrefix(ref, prefix)
}

type gitRefResolver struct {
	repo *repositoryResolver
	name string
}

func (r *gitRefResolver) Name() string        { return r.name }
func (r *gitRefResolver) DisplayName() string { return gitRefDisplayName(r.name) }
func (r *gitRefResolver) Prefix() string      { return gitRefPrefix(r.name) }
func (r *gitRefResolver) Target() *gitObjectResolver {
	return &gitObjectResolver{repo: r.repo, revspec: r.name}
}
func (r *gitRefResolver) Repository() *repositoryResolver { return r.repo }

type gitObject struct {
	oid gitObjectID
}

func (o *gitObject) OID() gitObjectID { return o.oid }

type gitObjectResolver struct {
	repo    *repositoryResolver
	revspec string
}

func (o *gitObjectResolver) OID(ctx context.Context) (gitObjectID, error) {
	// 🚨 SECURITY: DO NOT REMOVE THIS CHECK! ResolveRev is responsible for ensuring 🚨
	// the user has permissions to access the repository.
	resolvedRev, err := backend.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: o.repo.repo.ID, Rev: o.revspec})
	if err != nil {
		return "", err
	}
	return gitObjectID(resolvedRev.CommitID), nil
}

type gitRevSpecExpr struct {
	expr string
}

func (e *gitRevSpecExpr) Expr() string { return e.expr }

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
