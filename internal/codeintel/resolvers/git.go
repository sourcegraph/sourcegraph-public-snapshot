package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitTreeEntryResolver interface {
	Path() string
	Name() string
	ToGitTree() (GitTreeEntryResolver, bool)
	ToGitBlob() (GitTreeEntryResolver, bool)
	ByteSize(ctx context.Context) (int32, error)
	Content(ctx context.Context, args *GitTreeContentPageArgs) (string, error)
	Commit() GitCommitResolver
	Repository() RepositoryResolver
	CanonicalURL() string
	IsRoot() bool
	IsDirectory() bool
	URL(ctx context.Context) (string, error)
	Submodule() GitSubmoduleResolver
}

type GitTreeContentPageArgs struct {
	StartLine *int32
	EndLine   *int32
}

type RepositoryResolver interface {
	ID() graphql.ID
	Name() string
	Type(ctx context.Context) (*types.Repo, error)
	CommitFromID(ctx context.Context, args *RepositoryCommitArgs, commitID api.CommitID) (GitCommitResolver, error)
	URL() string
	URI(ctx context.Context) (string, error)
	ExternalRepository() ExternalRepositoryResolver
}

type RepositoryCommitArgs struct {
	Rev          string
	InputRevspec *string
}

type GitCommitResolver interface {
	ID() graphql.ID
	Repository() RepositoryResolver
	OID() GitObjectID
	AbbreviatedOID() string
	URL() string
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

type ExternalRepositoryResolver interface {
	ServiceType() string
	ServiceID() string
}

type GitSubmoduleResolver interface {
	URL() string
	Commit() string
	Path() string
}
