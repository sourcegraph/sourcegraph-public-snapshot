package graphqlbackend

import (
	"context"
	"errors"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// Labels is the implementation of the GraphQL type LabelsMutation. If it is not set at runtime, a
// "not implemented" error is returned to API clients who invoke it.
//
// This is contributed by enterprise.
var Labels LabelsResolver

const GQLTypeLabel = "Label"

func MarshalLabelID(id int64) graphql.ID {
	return relay.MarshalID(GQLTypeLabel, id)
}

func UnmarshalLabelID(id graphql.ID) (dbID int64, err error) {
	if typ := relay.UnmarshalKind(id); typ != GQLTypeLabel {
		return 0, fmt.Errorf("label ID has unexpected type type %q", typ)
	}
	err = relay.UnmarshalSpec(id, &dbID)
	return
}

var errLabelsNotImplemented = errors.New("labels is not implemented")

// LabelByID is called to look up a Label given its GraphQL ID.
func LabelByID(ctx context.Context, id graphql.ID) (Label, error) {
	if Labels == nil {
		return nil, errors.New("labels is not implemented")
	}
	return Labels.LabelByID(ctx, id)
}

// LabelsForLabelable returns an instance of the GraphQL LabelConnection type with the list of
// labels for a Labelable.
func LabelsForLabelable(ctx context.Context, labelable graphql.ID, arg *graphqlutil.ConnectionArgs) (LabelConnection, error) {
	if Labels == nil {
		return nil, errors.New("labels is not implemented")
	}
	return Labels.LabelsForLabelable(ctx, labelable, arg)
}

// LabelsInRepository returns an instance of the GraphQL LabelConnection type with the list of
// labels defined in a repository.
func LabelsInRepository(ctx context.Context, repository graphql.ID, arg *graphqlutil.ConnectionArgs) (LabelConnection, error) {
	if Labels == nil {
		return nil, errors.New("labels is not implemented")
	}
	return Labels.LabelsInRepository(ctx, repository, arg)
}

func (schemaResolver) CreateLabel(ctx context.Context, arg *CreateLabelArgs) (Label, error) {
	if Labels == nil {
		return nil, errLabelsNotImplemented
	}
	return Labels.CreateLabel(ctx, arg)
}

func (schemaResolver) UpdateLabel(ctx context.Context, arg *UpdateLabelArgs) (Label, error) {
	if Labels == nil {
		return nil, errLabelsNotImplemented
	}
	return Labels.UpdateLabel(ctx, arg)
}

func (schemaResolver) DeleteLabel(ctx context.Context, arg *DeleteLabelArgs) (*EmptyResponse, error) {
	if Labels == nil {
		return nil, errLabelsNotImplemented
	}
	return Labels.DeleteLabel(ctx, arg)
}

func (schemaResolver) AddLabelsToLabelable(ctx context.Context, arg *AddRemoveLabelsToFromLabelableArgs) (*ToLabelable, error) {
	if Labels == nil {
		return nil, errLabelsNotImplemented
	}
	return Labels.AddLabelsToLabelable(ctx, arg)
}

func (schemaResolver) RemoveLabelsFromLabelable(ctx context.Context, arg *AddRemoveLabelsToFromLabelableArgs) (*ToLabelable, error) {
	if Labels == nil {
		return nil, errLabelsNotImplemented
	}
	return Labels.RemoveLabelsFromLabelable(ctx, arg)
}

// LabelsResolver is the interface for the GraphQL type LabelsMutation.
type LabelsResolver interface {
	CreateLabel(context.Context, *CreateLabelArgs) (Label, error)
	UpdateLabel(context.Context, *UpdateLabelArgs) (Label, error)
	DeleteLabel(context.Context, *DeleteLabelArgs) (*EmptyResponse, error)
	AddLabelsToLabelable(context.Context, *AddRemoveLabelsToFromLabelableArgs) (*ToLabelable, error)
	RemoveLabelsFromLabelable(context.Context, *AddRemoveLabelsToFromLabelableArgs) (*ToLabelable, error)

	// LabelByID is called by the LabelByID func but is not in the GraphQL API.
	LabelByID(context.Context, graphql.ID) (Label, error)

	// LabelsFor is called by the LabelsFor func but is not in the GraphQL API.
	LabelsForLabelable(ctx context.Context, labelable graphql.ID, arg *graphqlutil.ConnectionArgs) (LabelConnection, error)

	// LabelsInRepository is called by the LabelsInRepository func but is not in the GraphQL API.
	LabelsInRepository(ctx context.Context, repository graphql.ID, arg *graphqlutil.ConnectionArgs) (LabelConnection, error)
}

type CreateLabelArgs struct {
	Input struct {
		Repository  graphql.ID
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
	Repository(context.Context) (*RepositoryResolver, error)
}

// Labelable is the interface for the Labelable GraphQL type.
type Labelable interface {
	Labels(context.Context, *graphqlutil.ConnectionArgs) (LabelConnection, error)
}

type ToLabelable struct {
	Thread Thread
}

func (v ToLabelable) Labelable() Labelable {
	switch {
	case v.Thread != nil:
		return v.Thread
	default:
		panic("no labelable")
	}
}

func (v ToLabelable) ToThread() (Thread, bool) {
	return v.Thread, v.Thread != nil
}

func (v ToLabelable) Labels(ctx context.Context, arg *graphqlutil.ConnectionArgs) (LabelConnection, error) {
	return v.Labelable().Labels(ctx, arg)
}

// LabelConnection is the interface for the GraphQL type LabelConnection.
type LabelConnection interface {
	Nodes(context.Context) ([]Label, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}

var EmptyLabelConnection emptyLabelConnection

type emptyLabelConnection struct{}

func (emptyLabelConnection) Nodes(context.Context) ([]Label, error)    { return nil, nil }
func (emptyLabelConnection) TotalCount(context.Context) (int32, error) { return 0, nil }
func (emptyLabelConnection) PageInfo(context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}
