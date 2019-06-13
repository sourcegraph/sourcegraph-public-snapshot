package projects

import (
	"context"
	"fmt"
	"strconv"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// 🚨 SECURITY: TODO!(sqs): There are virtually no security checks here and they MUST be added.

// gqlProject implements the GraphQL type Project.
type gqlProject struct{ db *dbProject }

// ProjectByID looks up and returns the Project with the given GraphQL ID. If no such Project
// exists, it returns a non-nil error.
func (r GraphQLResolver) ProjectByID(ctx context.Context, id graphql.ID) (graphqlbackend.Project, error) {
	dbID, err := unmarshalProjectID(id)
	if err != nil {
		return nil, err
	}
	return r.ProjectByDBID(ctx, dbID)
}

// ProjectByIDWithoutKind looks up and returns the Project with the given GraphQL
// Project.idWithoutKind value. If no such Project exists, it returns a non-nil error.
func (r GraphQLResolver) ProjectByIDWithoutKind(ctx context.Context, idWithoutKind string) (graphqlbackend.Project, error) {
	dbID, err := strconv.ParseInt(idWithoutKind, 10, 64)
	if err != nil {
		return nil, err
	}
	return r.ProjectByDBID(ctx, dbID)
}

// ProjectByDBID looks up and returns the Project with the given database ID. If no such Project
// exists, it returns a non-nil error.
func (GraphQLResolver) ProjectByDBID(ctx context.Context, id int64) (graphqlbackend.Project, error) {
	if MockProjectByDBID != nil {
		return MockProjectByDBID(id)
	}
	v, err := dbProjects{}.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &gqlProject{db: v}, nil
}

func (v *gqlProject) ID() graphql.ID {
	return marshalProjectID(v.db.ID)
}

func (v *gqlProject) IDWithoutKind() string { return strconv.FormatInt(v.db.ID, 10) }

func (v *gqlProject) DBID() int64 { return v.db.ID }

func marshalProjectID(id int64) graphql.ID {
	return relay.MarshalID("Project", id)
}

func unmarshalProjectID(id graphql.ID) (dbID int64, err error) {
	err = relay.UnmarshalSpec(id, &dbID)
	return
}

func (v *gqlProject) Name() string { return v.db.Name }

func (v *gqlProject) Namespace(ctx context.Context) (*graphqlbackend.NamespaceResolver, error) {
	switch {
	case v.db.NamespaceUserID != 0:
		org, err := graphqlbackend.OrgByIDInt32(ctx, v.db.NamespaceOrgID)
		if err != nil {
			return nil, err
		}
		return &graphqlbackend.NamespaceResolver{Namespace: org}, nil
	case v.db.NamespaceOrgID != 0:
		org, err := graphqlbackend.OrgByIDInt32(ctx, v.db.NamespaceOrgID)
		if err != nil {
			return nil, err
		}
		return &graphqlbackend.NamespaceResolver{Namespace: org}, nil
	default:
		return nil, fmt.Errorf("unrecognized namespace for project %s", v.ID())
	}
}

func (v *gqlProject) Labels(ctx context.Context, args *graphqlutil.ConnectionArgs) (graphqlbackend.LabelConnection, error) {
	return graphqlbackend.LabelsDefinedIn(ctx, v.ID(), args)
}

func (v *gqlProject) URL() string { return fmt.Sprintf("/p/%d", v.db.ID) }

// MockProjectByDBID mocks (GraphQLResolver).ProjectByDBID, for use in tests only.
var MockProjectByDBID func(int64) (graphqlbackend.Project, error)

// TestNewProject creates a graphqlbackend.Project value, for use in tests only.
func TestNewProject(id int64, name string, namespaceUserID, namespaceOrgID int32) graphqlbackend.Project {
	return &gqlProject{
		db: &dbProject{
			ID:              id,
			Name:            name,
			NamespaceUserID: namespaceUserID,
			NamespaceOrgID:  namespaceOrgID,
		},
	}
}
