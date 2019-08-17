package graphqlbackend

import (
	"context"
	"errors"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

type gitObjectType string

func (gitObjectType) ImplementsGraphQLType(name string) bool { return name == "GitObjectType" }

const (
	gitObjectTypeCommit  gitObjectType = "GIT_COMMIT"
	gitObjectTypeTag     gitObjectType = "GIT_TAG"
	gitObjectTypeTree    gitObjectType = "GIT_TREE"
	gitObjectTypeBlob    gitObjectType = "GIT_BLOB"
	gitObjectTypeUnknown gitObjectType = "GIT_UNKNOWN"
)

func toGitObjectType(t git.ObjectType) gitObjectType {
	switch t {
	case git.ObjectTypeCommit:
		return gitObjectTypeCommit
	case git.ObjectTypeTag:
		return gitObjectTypeTag
	case git.ObjectTypeTree:
		return gitObjectTypeTree
	case git.ObjectTypeBlob:
		return gitObjectTypeBlob
	}
	return gitObjectTypeUnknown
}

type GitObjectID string

func (GitObjectID) ImplementsGraphQLType(name string) bool {
	return name == "GitObjectID"
}

func (id *GitObjectID) UnmarshalGraphQL(input interface{}) error {
	if input, ok := input.(string); ok && git.IsAbsoluteRevision(input) {
		*id = GitObjectID(input)
		return nil
	}
	return errors.New("GitObjectID: expected 40-character string (SHA-1 hash)")
}

type gitObject struct {
	repo *repositoryResolver
	oid  GitObjectID
	typ  gitObjectType
}

func (o *gitObject) OID(ctx context.Context) (GitObjectID, error) { return o.oid, nil }
func (o *gitObject) AbbreviatedOID(ctx context.Context) (string, error) {
	return string(o.oid[:7]), nil
}

func (o *gitObject) Commit(ctx context.Context) (*gitCommitResolver, error) {
	return o.repo.Commit(ctx, &repositoryCommitArgs{Rev: string(o.oid)})
}
func (o *gitObject) Type(context.Context) (gitObjectType, error) { return o.typ, nil }

type gitObjectResolver struct {
	repo    *repositoryResolver
	revspec string

	once sync.Once
	oid  GitObjectID
	typ  gitObjectType
	err  error
}

func (o *gitObjectResolver) resolve(ctx context.Context) (GitObjectID, gitObjectType, error) {
	o.once.Do(func() {
		cachedRepo, err := backend.CachedGitRepo(ctx, o.repo.repo)
		if err != nil {
			o.err = err
			return
		}
		oid, objectType, err := git.GetObject(ctx, *cachedRepo, o.revspec)
		if err != nil {
			o.err = err
			return
		}
		o.oid = GitObjectID(oid.String())
		o.typ = toGitObjectType(objectType)
	})
	return o.oid, o.typ, o.err
}

func (o *gitObjectResolver) OID(ctx context.Context) (GitObjectID, error) {
	oid, _, err := o.resolve(ctx)
	return oid, err
}

func (o *gitObjectResolver) AbbreviatedOID(ctx context.Context) (string, error) {
	oid, _, err := o.resolve(ctx)
	if err != nil {
		return "", err
	}
	return string(oid[:7]), nil
}

func (o *gitObjectResolver) Commit(ctx context.Context) (*gitCommitResolver, error) {
	oid, _, err := o.resolve(ctx)
	if err != nil {
		return nil, err
	}
	return o.repo.Commit(ctx, &repositoryCommitArgs{Rev: string(oid)})
}

func (o *gitObjectResolver) Type(ctx context.Context) (gitObjectType, error) {
	_, typ, err := o.resolve(ctx)
	return typ, err
}
