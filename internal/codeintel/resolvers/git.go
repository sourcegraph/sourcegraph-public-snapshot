// Package resolvers contains slimmed down copies of some of
// the resolvers in the jumbo graphqlbackend package.
//
// Ideally, we would not have these duplicate copies at all,
// so avoid adding new functionality here.
package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RepositoryResolver interface {
	RepoID() api.RepoID // exposed for internal caches
	ID() graphql.ID
	Name() string
	URL() string
	ExternalRepository() ExternalRepositoryResolver
}

type ExternalRepositoryResolver interface {
	ServiceType() string
	ServiceID() string
}

type GitCommitResolver interface {
	ID() graphql.ID
	Repository() RepositoryResolver
	OID() GitObjectID
	AbbreviatedOID() string
	URL() string
	URI() string                                // exposed for internal URL construction
	Tags(ctx context.Context) ([]string, error) // exposed for internal memoization of gitserver requests
}

type GitObjectID string

func (GitObjectID) ImplementsGraphQLType(name string) bool {
	return name == "GitObjectID"
}

func (id *GitObjectID) UnmarshalGraphQL(input any) error {
	if input, ok := input.(string); ok && gitdomain.IsAbsoluteRevision(input) {
		*id = GitObjectID(input)
		return nil
	}
	return errors.New("GitObjectID: expected 40-character string (SHA-1 hash)")
}

type GitTreeEntryResolver interface {
	Repository() RepositoryResolver
	Commit() GitCommitResolver
	Path() string
	Name() string
	URL() string
	Content(ctx context.Context, args *GitTreeContentPageArgs) (string, error)
}

// GitBlobResolver is a thin wrapper around GitTreeEntryResolver for readability.
//
// By embedding GitTreeEntryResolver, we enable seamless conversion between the interfaces.
type GitBlobResolver interface {
	GitTreeEntryResolver
}

// GitTreeResolver is a thin wrapper around GitTreeEntryResolver for readability.
//
// By embedding GitTreeEntryResolver, we enable seamless conversion between the interfaces.
type GitTreeResolver interface {
	GitTreeEntryResolver
}

type GitTreeContentPageArgs struct {
	StartLine *int32
	EndLine   *int32
}
