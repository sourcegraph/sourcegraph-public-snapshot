package graphqlbackend

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
)

// Threadlike is the shared interface among threads, issues, and changesets.
type Threadlike interface {
	PartialComment
	ID() graphql.ID
	DBID() int64
	Repository(context.Context) (*RepositoryResolver, error)
	Number() string
	Title() string
	updatable
	URL(context.Context) (string, error)
	CampaignNode
	TimelineItems(context.Context, *EventConnectionCommonArgs) (EventConnection, error)
}

type updateThreadlikeInput struct {
	ID    graphql.ID
	Title *string
	Body  *string
}

type createThreadlikeInput struct {
	Repository graphql.ID
	Title      string
	Body       *string
}
