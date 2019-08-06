package issues

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/internal"
)

func (GraphQLResolver) Issues(ctx context.Context, arg *graphqlutil.ConnectionArgs) (graphqlbackend.IssueConnection, error) {
	return issuesByOptions(ctx, internal.DBThreadsListOptions{}, arg)
}

func (GraphQLResolver) IssuesForRepository(ctx context.Context, repositoryID graphql.ID, arg *graphqlutil.ConnectionArgs) (graphqlbackend.IssueConnection, error) {
	repo, err := graphqlbackend.RepositoryByID(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	return issuesByOptions(ctx, internal.DBThreadsListOptions{
		RepositoryID: repo.DBID(),
	}, arg)
}

func issuesByOptions(ctx context.Context, options internal.DBThreadsListOptions, arg *graphqlutil.ConnectionArgs) (graphqlbackend.IssueConnection, error) {
	options.Type = internal.DBThreadTypeIssue
	list, err := internal.DBThreads{}.List(ctx, options)
	if err != nil {
		return nil, err
	}
	issues := make([]*gqlIssue, len(list))
	for i, a := range list {
		issues[i] = newGQLIssue(a)
	}
	return &issueConnection{arg: arg, issues: issues}, nil
}

type issueConnection struct {
	arg    *graphqlutil.ConnectionArgs
	issues []*gqlIssue
}

func (r *issueConnection) Nodes(ctx context.Context) ([]graphqlbackend.Issue, error) {
	issues := r.issues
	if first := r.arg.First; first != nil && len(issues) > int(*first) {
		issues = issues[:int(*first)]
	}

	issues2 := make([]graphqlbackend.Issue, len(issues))
	for i, l := range issues {
		issues2[i] = l
	}
	return issues2, nil
}

func (r *issueConnection) TotalCount(ctx context.Context) (int32, error) {
	return int32(len(r.issues)), nil
}

func (r *issueConnection) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(r.arg.First != nil && int(*r.arg.First) < len(r.issues)), nil
}
