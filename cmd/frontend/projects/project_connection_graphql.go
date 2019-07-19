package projects

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

func (GraphQLResolver) ProjectsInNamespace(ctx context.Context, namespace graphql.ID, arg *graphqlutil.ConnectionArgs) (graphqlbackend.ProjectConnection, error) {
	var opt dbProjectsListOptions
	var err error
	opt.NamespaceUserID, opt.NamespaceOrgID, err = graphqlbackend.NamespaceDBIDByID(ctx, namespace)
	if err != nil {
		return nil, err
	}

	list, err := dbProjects{}.List(ctx, opt)
	if err != nil {
		return nil, err
	}

	projects := make([]*gqlProject, len(list))
	for i, p := range list {
		projects[i] = &gqlProject{db: p}
	}
	return &projectConnection{arg: arg, projects: projects}, nil
}

type projectConnection struct {
	arg      *graphqlutil.ConnectionArgs
	projects []*gqlProject
}

func (r *projectConnection) Nodes(ctx context.Context) ([]graphqlbackend.Project, error) {
	projects := r.projects
	if first := r.arg.First; first != nil && len(projects) > int(*first) {
		projects = projects[:int(*first)]
	}

	projects2 := make([]graphqlbackend.Project, len(projects))
	for i, l := range projects {
		projects2[i] = l
	}
	return projects2, nil
}

func (r *projectConnection) TotalCount(ctx context.Context) (int32, error) {
	return int32(len(r.projects)), nil
}

func (r *projectConnection) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(r.arg.First != nil && int(*r.arg.First) < len(r.projects)), nil
}
