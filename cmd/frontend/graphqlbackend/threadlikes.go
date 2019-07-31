package graphqlbackend

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
)

// Threadlike is the shared interface among threads, issues, and changesets.
type Threadlike interface {
	Comment
	ID() graphql.ID
	DBID() int64
	Type() ThreadlikeType
	Repository(context.Context) (*RepositoryResolver, error)
	Number() string
	Title() string
	ExternalURL() *string
	URL(context.Context) (string, error)
}

type ThreadlikeType string

const (
	ThreadlikeTypeThread    ThreadlikeType = "THREAD"
	ThreadlikeTypeIssue                    = "ISSUE"
	ThreadlikeTypeChangeset                = "CHANGESET"
)

type updateThreadlikeInput struct {
	ID          graphql.ID
	Title       *string
	Body        *string
	ExternalURL *string
}

type createThreadlikeInput struct {
	Repository  graphql.ID
	Title       string
	Body        *string
	ExternalURL *string
}
