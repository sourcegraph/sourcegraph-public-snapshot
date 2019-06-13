package projects

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (GraphQLResolver) CreateProject(ctx context.Context, arg *graphqlbackend.CreateProjectArgs) (graphqlbackend.Project, error) {
	v := &dbProject{
		Name: arg.Input.Name,
	}

	var err error
	v.NamespaceUserID, v.NamespaceOrgID, err = getNamespaceDBID(ctx, arg.Input.Namespace)
	if err != nil {
		return nil, err
	}

	project, err := dbProjects{}.Create(ctx, v)
	if err != nil {
		return nil, err
	}
	return &gqlProject{db: project}, nil
}

func (r GraphQLResolver) UpdateProject(ctx context.Context, arg *graphqlbackend.UpdateProjectArgs) (graphqlbackend.Project, error) {
	l, err := r.ProjectByID(ctx, arg.Input.ID)
	if err != nil {
		return nil, err
	}
	project, err := dbProjects{}.Update(ctx, l.DBID(), dbProjectUpdate{
		Name: arg.Input.Name,
	})
	if err != nil {
		return nil, err
	}
	return &gqlProject{db: project}, nil
}

func (r GraphQLResolver) DeleteProject(ctx context.Context, arg *graphqlbackend.DeleteProjectArgs) (*graphqlbackend.EmptyResponse, error) {
	gqlProject, err := r.ProjectByID(ctx, arg.Project)
	if err != nil {
		return nil, err
	}
	return nil, dbProjects{}.DeleteByID(ctx, gqlProject.DBID())
}
