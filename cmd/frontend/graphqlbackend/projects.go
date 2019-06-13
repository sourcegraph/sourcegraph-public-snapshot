package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// Projects is the implementation of the GraphQL type ProjectsMutation. If it is not set at runtime, a
// "not implemented" error is returned to API clients who invoke it.
var Projects ProjectsResolver

// ProjectByID is called to look up a Project given its GraphQL ID.
func ProjectByID(ctx context.Context, id graphql.ID) (Project, error) {
	if Projects == nil {
		return nil, errors.New("projects is not implemented")
	}
	return Projects.ProjectByID(ctx, id)
}

// ProjectByDBID is called to look up a Project given its database ID.
func ProjectByDBID(ctx context.Context, id int64) (Project, error) {
	if Projects == nil {
		return nil, errors.New("projects is not implemented")
	}
	return Projects.ProjectByDBID(ctx, id)
}

func (schemaResolver) Project(ctx context.Context, args *struct{ IDWithoutKind string }) (Project, error) {
	if Projects == nil {
		return nil, errors.New("projects is not implemented")
	}
	return Projects.ProjectByIDWithoutKind(ctx, args.IDWithoutKind)
}

// ProjectsInNamespace returns an instance of the GraphQL ProjectConnection type with the list of
// projects in a namespace.
func ProjectsInNamespace(ctx context.Context, namespace graphql.ID, arg *graphqlutil.ConnectionArgs) (ProjectConnection, error) {
	if Projects == nil {
		return nil, errors.New("projects is not implemented")
	}
	return Projects.ProjectsInNamespace(ctx, namespace, arg)
}

func (schemaResolver) Projects() (ProjectsResolver, error) {
	if Projects == nil {
		return nil, errors.New("projects is not implemented")
	}
	return Projects, nil
}

// ProjectsResolver is the interface for the GraphQL type ProjectsMutation.
type ProjectsResolver interface {
	CreateProject(context.Context, *CreateProjectArgs) (Project, error)
	UpdateProject(context.Context, *UpdateProjectArgs) (Project, error)
	DeleteProject(context.Context, *DeleteProjectArgs) (*EmptyResponse, error)

	// ProjectByID is called by the ProjectByID func but is not in the GraphQL API.
	ProjectByID(context.Context, graphql.ID) (Project, error)

	// ProjectByIDWithoutKind is called by (schemaResolver).Project but is not in the GraphQL API.
	ProjectByIDWithoutKind(context.Context, string) (Project, error)

	// ProjectByDBID is called by the ProjectByDBID func but is not in the GraphQL API.
	ProjectByDBID(context.Context, int64) (Project, error)

	// ProjectsInNamespace is called by the ProjectsIn func but is not in the GraphQL API.
	ProjectsInNamespace(ctx context.Context, namespace graphql.ID, arg *graphqlutil.ConnectionArgs) (ProjectConnection, error)
}

type CreateProjectArgs struct {
	Input struct {
		Namespace graphql.ID
		Name      string
	}
}

type UpdateProjectArgs struct {
	Input struct {
		ID   graphql.ID
		Name *string
	}
}

type DeleteProjectArgs struct {
	Project graphql.ID
}

// Project is the interface for the GraphQL type Project.
type Project interface {
	ID() graphql.ID
	IDWithoutKind() string
	Name() string
	Namespace(context.Context) (*NamespaceResolver, error)
	Labels(context.Context, *graphqlutil.ConnectionArgs) (LabelConnection, error)
	URL() string

	// DBID is exposed for internal use but is not in the GraphQL API.
	DBID() int64
}

// ProjectConnection is the interface for the GraphQL type ProjectConnection.
type ProjectConnection interface {
	Nodes(context.Context) ([]Project, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}
