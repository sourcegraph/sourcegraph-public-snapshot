package issues

import (
	"context"
	"strconv"

	"github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/internal"
)

// 🚨 SECURITY: TODO!(sqs): there needs to be security checks everywhere here! there are none

// gqlIssue implements the GraphQL type Issue.
type gqlIssue struct {
	threadlike.GQLThreadlike
	db *internal.DBThread
}

func newGQLIssue(db *internal.DBThread) *gqlIssue {
	return &gqlIssue{
		GQLThreadlike: threadlike.GQLThreadlike{
			DB:             db,
			PartialComment: comments.GraphQLResolver{}.LazyCommentByID(threadlike.MarshalID(threadlike.GQLTypeIssue, db.ID)),
		},
		db: db,
	}
}

// issueByID looks up and returns the Issue with the given GraphQL ID. If no such Issue exists, it
// returns a non-nil error.
func issueByID(ctx context.Context, id graphql.ID) (*gqlIssue, error) {
	dbID, err := threadlike.UnmarshalIDOfType(threadlike.GQLTypeIssue, id)
	if err != nil {
		return nil, err
	}
	return issueByDBID(ctx, dbID)
}

func (GraphQLResolver) IssueByID(ctx context.Context, id graphql.ID) (graphqlbackend.Issue, error) {
	return issueByID(ctx, id)
}

// issueByDBID looks up and returns the Issue with the given database ID. If no such Issue exists,
// it returns a non-nil error.
func issueByDBID(ctx context.Context, dbID int64) (*gqlIssue, error) {
	v, err := internal.DBThreads{}.GetByID(ctx, dbID)
	if err != nil {
		return nil, err
	}
	return newGQLIssue(v), nil
}

func (GraphQLResolver) IssueInRepository(ctx context.Context, repositoryID graphql.ID, number string) (graphqlbackend.Issue, error) {
	issueDBID, err := strconv.ParseInt(number, 10, 64)
	if err != nil {
		return nil, err
	}
	// TODO!(sqs): access checks
	issue, err := issueByDBID(ctx, issueDBID)
	if err != nil {
		return nil, err
	}

	// TODO!(sqs): check that the issue is indeed in the repo. When we make the issue number
	// sequence per-repo, this will become necessary to even retrieve the issue. for now, the ID is
	// global, so we need to perform this check.
	assertedRepo, err := graphqlbackend.RepositoryByID(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	if issue.db.RepositoryID != assertedRepo.DBID() {
		return nil, errors.New("issue does not exist in repository")
	}

	return issue, nil
}

func (v *gqlIssue) State() graphqlbackend.IssueState {
	return graphqlbackend.IssueState(v.db.State)
}

func (v *gqlIssue) DiagnosticsData() string {
	return string(v.db.DiagnosticsData)
}
