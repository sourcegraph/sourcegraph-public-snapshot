package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// ThreadOrIssueOrChangesets is the implementation of the GraphQL API for
// threads-or-issues-or-changesets queries and mutations. If it is not set at runtime, a "not
// implemented" error is returned to API clients who invoke it.
//
// This is contributed by enterprise.
var ThreadOrIssueOrChangesets ThreadOrIssueOrChangesetsResolver

var errThreadOrIssueOrChangesetsNotImplemented = errors.New("threadOrIssueOrChangesets is not implemented")

// ThreadOrIssueOrChangesetByID is called to look up a ThreadOrIssueOrChangeset given its GraphQL
// ID.
func ThreadOrIssueOrChangesetByID(ctx context.Context, id graphql.ID) (*ThreadOrIssueOrChangeset, error) {
	if ThreadOrIssueOrChangesets == nil {
		return nil, errThreadOrIssueOrChangesetsNotImplemented
	}
	return ThreadOrIssueOrChangesets.ThreadOrIssueOrChangesetByID(ctx, id)
}

// ThreadOrIssueOrChangesetInRepository returns a specific threadOrIssueOrChangeset in the specified
// repository.
func ThreadOrIssueOrChangesetInRepository(ctx context.Context, repository graphql.ID, number string) (*ThreadOrIssueOrChangeset, error) {
	if ThreadOrIssueOrChangesets == nil {
		return nil, errThreadOrIssueOrChangesetsNotImplemented
	}
	return ThreadOrIssueOrChangesets.ThreadOrIssueOrChangesetInRepository(ctx, repository, number)
}

// ThreadOrIssueOrChangesetsForRepository returns an instance of the GraphQL
// ThreadOrIssueOrChangesetConnection type with the list of threads, issues, and changesets defined
// in a repository.
func ThreadOrIssueOrChangesetsForRepository(ctx context.Context, repository graphql.ID, arg *graphqlutil.ConnectionArgs) (ThreadOrIssueOrChangesetConnection, error) {
	if ThreadOrIssueOrChangesets == nil {
		return nil, errThreadOrIssueOrChangesetsNotImplemented
	}
	return ThreadOrIssueOrChangesets.ThreadOrIssueOrChangesetsForRepository(ctx, repository, arg)
}

func (schemaResolver) ThreadOrIssueOrChangesets(ctx context.Context, arg *graphqlutil.ConnectionArgs) (ThreadOrIssueOrChangesetConnection, error) {
	if ThreadOrIssueOrChangesets == nil {
		return nil, errThreadOrIssueOrChangesetsNotImplemented
	}
	return ThreadOrIssueOrChangesets.ThreadOrIssueOrChangesets(ctx, arg)
}

// ThreadOrIssueOrChangesetsResolver is the interface for the GraphQL threads-or-issues-or-changesets queries and
// mutations.
type ThreadOrIssueOrChangesetsResolver interface {
	// Queries
	ThreadOrIssueOrChangesets(context.Context, *graphqlutil.ConnectionArgs) (ThreadOrIssueOrChangesetConnection, error)

	// ThreadOrIssueOrChangesetByID is called by the ThreadOrIssueOrChangesetByID func but is not in
	// the GraphQL API.
	ThreadOrIssueOrChangesetByID(context.Context, graphql.ID) (*ThreadOrIssueOrChangeset, error)

	// ThreadOrIssueOrChangesetInRepository is called by the ThreadOrIssueOrChangesetInRepository
	// func but is not in the GraphQL API.
	ThreadOrIssueOrChangesetInRepository(ctx context.Context, repository graphql.ID, number string) (*ThreadOrIssueOrChangeset, error)

	// ThreadOrIssueOrChangesetsForRepository is called by the
	// ThreadOrIssueOrChangesetsForRepository func but is not in the GraphQL API.
	ThreadOrIssueOrChangesetsForRepository(ctx context.Context, repository graphql.ID, arg *graphqlutil.ConnectionArgs) (ThreadOrIssueOrChangesetConnection, error)
}

// ThreadOrIssueOrChangeset implements the GraphQL union ThreadOrIssueOrChangeset. Exactly 1 of the
// fields must be non-nil.
type ThreadOrIssueOrChangeset struct {
	Thread    Thread
	Changeset Changeset
}

func (v ThreadOrIssueOrChangeset) ToThread() (Thread, bool) {
	return v.Thread, v.Thread != nil
}

func (v ThreadOrIssueOrChangeset) ToChangeset() (Changeset, bool) {
	return v.Changeset, v.Changeset != nil
}

// ThreadOrIssueOrChangesetConnection is the interface for the GraphQL type
// ThreadOrIssueOrChangesetConnection.
type ThreadOrIssueOrChangesetConnection interface {
	Nodes(context.Context) ([]ThreadOrIssueOrChangeset, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}
