package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// Issues is the implementation of the GraphQL API for issues queries and mutations. If it is not
// set at runtime, a "not implemented" error is returned to API clients who invoke it.
//
// This is contributed by enterprise.
var Issues IssuesResolver

var errIssuesNotImplemented = errors.New("issues is not implemented")

// IssueByID is called to look up a Issue given its GraphQL ID.
func IssueByID(ctx context.Context, id graphql.ID) (Issue, error) {
	if Issues == nil {
		return nil, errors.New("issues is not implemented")
	}
	return Issues.IssueByID(ctx, id)
}

// IssueInRepository returns a specific issue in the specified repository.
func IssueInRepository(ctx context.Context, repository graphql.ID, number string) (Issue, error) {
	if Issues == nil {
		return nil, errIssuesNotImplemented
	}
	return Issues.IssueInRepository(ctx, repository, number)
}

// IssuesForRepository returns an instance of the GraphQL IssueConnection type with the list
// of issues defined in a repository.
func IssuesForRepository(ctx context.Context, repository graphql.ID, arg *graphqlutil.ConnectionArgs) (IssueConnection, error) {
	if Issues == nil {
		return nil, errIssuesNotImplemented
	}
	return Issues.IssuesForRepository(ctx, repository, arg)
}

func (schemaResolver) Issues(ctx context.Context, arg *graphqlutil.ConnectionArgs) (IssueConnection, error) {
	if Issues == nil {
		return nil, errIssuesNotImplemented
	}
	return Issues.Issues(ctx, arg)
}

func (r schemaResolver) CreateIssue(ctx context.Context, arg *CreateIssueArgs) (Issue, error) {
	if Issues == nil {
		return nil, errIssuesNotImplemented
	}
	return Issues.CreateIssue(ctx, arg)
}

func (r schemaResolver) UpdateIssue(ctx context.Context, arg *UpdateIssueArgs) (Issue, error) {
	if Issues == nil {
		return nil, errIssuesNotImplemented
	}
	return Issues.UpdateIssue(ctx, arg)
}

func (r schemaResolver) DeleteIssue(ctx context.Context, arg *DeleteIssueArgs) (*EmptyResponse, error) {
	if Issues == nil {
		return nil, errIssuesNotImplemented
	}
	return Issues.DeleteIssue(ctx, arg)
}

// IssuesResolver is the interface for the GraphQL issues queries and mutations.
type IssuesResolver interface {
	// Queries
	Issues(context.Context, *graphqlutil.ConnectionArgs) (IssueConnection, error)

	// Mutations
	CreateIssue(context.Context, *CreateIssueArgs) (Issue, error)
	UpdateIssue(context.Context, *UpdateIssueArgs) (Issue, error)
	DeleteIssue(context.Context, *DeleteIssueArgs) (*EmptyResponse, error)

	// IssueByID is called by the IssueByID func but is not in the GraphQL API.
	IssueByID(context.Context, graphql.ID) (Issue, error)

	// IssueInRepository is called by the IssueInRepository func but is not in the GraphQL API.
	IssueInRepository(ctx context.Context, repository graphql.ID, number string) (Issue, error)

	// IssuesForRepository is called by the IssuesForRepository func but is not in the GraphQL API.
	IssuesForRepository(ctx context.Context, repository graphql.ID, arg *graphqlutil.ConnectionArgs) (IssueConnection, error)
}

type CreateIssueArgs struct {
	Input struct {
		createThreadlikeInput
		// TODO!(sqs): add diagnostics
	}
}

type UpdateIssueArgs struct {
	Input struct {
		updateThreadlikeInput
	}
}

// TODO!(sqs): add diagnostics update mutation

type DeleteIssueArgs struct {
	Issue graphql.ID
}

type IssueState string

const (
	IssueStateOpen   IssueState = "OPEN"
	IssueStateClosed            = "CLOSED"
)

// Issue is the interface for the GraphQL type Issue.
type Issue interface {
	Threadlike
	State() IssueState
	DiagnosticsData() string // TODO!(sqs)
}

// IssueConnection is the interface for the GraphQL type IssueConnection.
type IssueConnection interface {
	Nodes(context.Context) ([]Issue, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}
