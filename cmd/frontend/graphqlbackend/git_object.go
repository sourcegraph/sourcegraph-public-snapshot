package graphqlbackend

import (
	"context"
	"fmt"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitObjectType string

func (GitObjectType) ImplementsGraphQLType(name string) bool { return name == "GitObjectType" }

const (
	GitObjectTypeCommit  GitObjectType = "GIT_COMMIT"
	GitObjectTypeTag     GitObjectType = "GIT_TAG"
	GitObjectTypeTree    GitObjectType = "GIT_TREE"
	GitObjectTypeBlob    GitObjectType = "GIT_BLOB"
	GitObjectTypeUnknown GitObjectType = "GIT_UNKNOWN"
)

func toGitObjectType(t gitdomain.ObjectType) GitObjectType {
	switch t {
	case gitdomain.ObjectTypeCommit:
		return GitObjectTypeCommit
	case gitdomain.ObjectTypeTag:
		return GitObjectTypeTag
	case gitdomain.ObjectTypeTree:
		return GitObjectTypeTree
	case gitdomain.ObjectTypeBlob:
		return GitObjectTypeBlob
	}
	return GitObjectTypeUnknown
}

type GitObjectID string

func (GitObjectID) ImplementsGraphQLType(name string) bool {
	return name == "GitObjectID"
}

func (id *GitObjectID) UnmarshalGraphQL(input any) error {
	if input, ok := input.(string); ok && gitserver.IsAbsoluteRevision(input) {
		*id = GitObjectID(input)
		return nil
	}
	return errors.New("GitObjectID: expected 40-character string (SHA-1 hash)")
}

type gitObject struct {
	repo *RepositoryResolver
	oid  GitObjectID
	typ  GitObjectType
}

func (o *gitObject) OID(ctx context.Context) (GitObjectID, error) { return o.oid, nil }
func (o *gitObject) AbbreviatedOID(ctx context.Context) (string, error) {
	return string(o.oid[:7]), nil
}

func (o *gitObject) Commit(ctx context.Context) (*GitCommitResolver, error) {
	return o.repo.Commit(ctx, &RepositoryCommitArgs{Rev: string(o.oid)})
}

// TODO: nsc this seems to be used for commits only?
func (o *gitObject) Tag(context.Context) (*GitTagResolver, error) {
	return nil, nil
}

func (o *gitObject) Type(context.Context) (GitObjectType, error) { return o.typ, nil }

type gitObjectResolver struct {
	repo    *RepositoryResolver
	revspec string

	once sync.Once
	oid  GitObjectID
	typ  GitObjectType
	err  error
}

func (o *gitObjectResolver) resolve(ctx context.Context) (GitObjectID, GitObjectType, error) {
	fmt.Printf("RESOLVING OBJECT %q\n", o.revspec)
	o.once.Do(func() {
		obj, err := gitserver.NewClient(o.repo.db).GetObject(ctx, o.repo.RepoName(), o.revspec)
		if err != nil {
			o.err = err
			return
		}
		o.oid = GitObjectID(obj.ID.String())
		o.typ = toGitObjectType(obj.Type)
		fmt.Printf("RESOLVED OBJECT %q %q %q\n", o.revspec, obj.ID.String(), obj.Type)
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

func (o *gitObjectResolver) Commit(ctx context.Context) (*GitCommitResolver, error) {
	fmt.Printf("SEARCHING FOR COMMIT %q\n", o.revspec)
	oid, typ, err := o.resolve(ctx)
	fmt.Printf("SEARCHED FOR COMMIT %q %q %q\n", o.revspec, typ, GitObjectTypeCommit)
	if err != nil || typ != GitObjectTypeCommit {
		return nil, err
	}
	return o.repo.Commit(ctx, &RepositoryCommitArgs{Rev: string(oid)})
}

func (o *gitObjectResolver) Tag(ctx context.Context) (*GitTagResolver, error) {
	fmt.Printf("SEARCHING FOR TAG %q\n", o.revspec)
	oid, typ, err := o.resolve(ctx)
	fmt.Printf("SEARCHED FOR TAG %q %q %q %q\n", oid, o.revspec, typ, GitObjectTypeTag)
	if err != nil || typ != GitObjectTypeTag {
		return nil, err
	}
	return o.repo.Tag(ctx, string(oid))
}

func (o *gitObjectResolver) Type(ctx context.Context) (GitObjectType, error) {
	_, typ, err := o.resolve(ctx)
	return typ, err
}
