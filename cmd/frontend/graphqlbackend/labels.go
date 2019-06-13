package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// Labels is the implementation of the GraphQL type LabelsMutation. If it is not set at runtime, a
// "not implemented" error is returned to API clients who invoke it.
//
// This is contributed by enterprise.
var Labels LabelsResolver

// LabelByID is called to look up a Label given its GraphQL ID.
func LabelByID(ctx context.Context, id graphql.ID) (Label, error) {
	if Labels == nil {
		return nil, errors.New("labels is not implemented")
	}
	return Labels.LabelByID(ctx, id)
}

// LabelsFor returns an instance of the GraphQL LabelConnection type with the list of labels for a
// Labelable.
//
// NOTE: Currently DiscussionThread is the only type to have labels.
func LabelsFor(ctx context.Context, labelable graphql.ID, arg *graphqlutil.ConnectionArgs) (LabelConnection, error) {
	if Labels == nil {
		return nil, errors.New("labels is not implemented")
	}
	return Labels.LabelsFor(ctx, labelable, arg)
}

// LabelsDefinedIn returns an instance of the GraphQL LabelConnection type with the list of labels
// defined in a project.
func LabelsDefinedIn(ctx context.Context, project graphql.ID, arg *graphqlutil.ConnectionArgs) (LabelConnection, error) {
	if Labels == nil {
		return nil, errors.New("labels is not implemented")
	}
	return Labels.LabelsDefinedIn(ctx, project, arg)
}

func (schemaResolver) Labels() (LabelsResolver, error) {
	if Labels == nil {
		return nil, errors.New("labels is not implemented")
	}
	return Labels, nil
}

// LabelsResolver is the interface for the GraphQL type LabelsMutation.
type LabelsResolver interface {
	CreateLabel(context.Context, *CreateLabelArgs) (Label, error)
	UpdateLabel(context.Context, *UpdateLabelArgs) (Label, error)
	DeleteLabel(context.Context, *DeleteLabelArgs) (*EmptyResponse, error)
	AddLabelsToLabelable(context.Context, *AddRemoveLabelsToFromLabelableArgs) (Labelable, error)
	RemoveLabelsFromLabelable(context.Context, *AddRemoveLabelsToFromLabelableArgs) (Labelable, error)

	// LabelByID is called by the LabelByID func but is not in the GraphQL API.
	LabelByID(context.Context, graphql.ID) (Label, error)

	// LabelsFor is called by the LabelsFor func but is not in the GraphQL API.
	LabelsFor(ctx context.Context, labelable graphql.ID, arg *graphqlutil.ConnectionArgs) (LabelConnection, error)

	// LabelsDefinedIn is called by the LabelsDefinedIn func but is not in the GraphQL API.
	LabelsDefinedIn(ctx context.Context, project graphql.ID, arg *graphqlutil.ConnectionArgs) (LabelConnection, error)
}

type CreateLabelArgs struct {
	Input struct {
		Project     graphql.ID
		Name        string
		Description *string
		Color       string
	}
}

type UpdateLabelArgs struct {
	Input struct {
		ID          graphql.ID
		Name        *string
		Description *string
		Color       *string
	}
}

type DeleteLabelArgs struct {
	Label graphql.ID
}

type AddRemoveLabelsToFromLabelableArgs struct {
	Labelable graphql.ID
	Labels    []graphql.ID
}

// Label is the interface for the GraphQL type Label.
type Label interface {
	ID() graphql.ID
	Name() string
	Description() *string
	Color() string
	Project(context.Context) (Project, error)
}

// Labelable is the interface for the GraphQL interface Labelable.
type Labelable interface {
	Labels(context.Context, *graphqlutil.ConnectionArgs) (LabelConnection, error)
	ToDiscussionThread() (*discussionThreadResolver, bool)
}

// LabelConnection is the interface for the GraphQL type LabelConnection.
type LabelConnection interface {
	Nodes(context.Context) ([]Label, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}
