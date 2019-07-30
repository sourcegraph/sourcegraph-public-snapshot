package threadlike

import (
	"context"
	"errors"
	"strconv"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/internal"
)

func newGQLThreadOrIssueOrChangeset(v *internal.DBThread) graphqlbackend.ThreadOrIssueOrChangeset {
	switch v.Type {
	case graphqlbackend.ThreadlikeTypeThread:
		return graphqlbackend.ThreadOrIssueOrChangeset{Thread: internal.ToGQLThread(v)}
	case graphqlbackend.ThreadlikeTypeIssue:
		panic("TODO!(sqs): not implemented")
	case graphqlbackend.ThreadlikeTypeChangeset:
		return graphqlbackend.ThreadOrIssueOrChangeset{Changeset: internal.ToGQLChangeset(v)}
	}
	panic("unrecognized thread type: " + v.Type)
}

// threadOrIssueOrChangesetByID looks up and returns the ThreadOrIssueOrChangeset with the given
// GraphQL ID. If no such ThreadOrIssueOrChangeset exists, it returns a non-nil error.
func threadOrIssueOrChangesetByID(ctx context.Context, id graphql.ID) (*graphqlbackend.ThreadOrIssueOrChangeset, error) {
	switch relay.UnmarshalKind(id) {
	case "Thread":
		thread, err := graphqlbackend.ThreadByID(ctx, id)
		if err != nil {
			return nil, err
		}
		return &graphqlbackend.ThreadOrIssueOrChangeset{Thread: thread}, nil
	case "Issue":
		panic("TODO!(sqs): not implemented")
	case "Changeset":
		changeset, err := graphqlbackend.ChangesetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		return &graphqlbackend.ThreadOrIssueOrChangeset{Changeset: changeset}, nil
	}
	return nil, errors.New("invalid ThreadOrIssueOrChangeset ID")
}

func (GraphQLResolver) ThreadOrIssueOrChangesetByID(ctx context.Context, id graphql.ID) (*graphqlbackend.ThreadOrIssueOrChangeset, error) {
	return threadOrIssueOrChangesetByID(ctx, id)
}

func (GraphQLResolver) ThreadOrIssueOrChangesetInRepository(ctx context.Context, repositoryID graphql.ID, number string) (*graphqlbackend.ThreadOrIssueOrChangeset, error) {
	dbID, err := strconv.ParseInt(number, 10, 64)
	if err != nil {
		return nil, err
	}

	dbThread, err := internal.DBThreads{}.GetByID(ctx, dbID)
	if err != nil {
		return nil, err
	}

	// TODO!(sqs): access checks

	// TODO!(sqs): check that the changeset is indeed in the repo. When we make the changeset number
	// sequence per-repo, this will become necessary to even retrieve the changeset. for now, the ID is
	// global, so we need to perform this check.
	assertedRepo, err := graphqlbackend.RepositoryByID(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	if dbThread.RepositoryID != assertedRepo.DBID() {
		return nil, errors.New("changeset does not exist in repository")
	}

	tmp := newGQLThreadOrIssueOrChangeset(dbThread)
	return &tmp, nil
}
