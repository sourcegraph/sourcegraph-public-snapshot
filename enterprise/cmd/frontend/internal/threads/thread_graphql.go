package threads

import (
	"context"
	"strconv"

	"github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads"
)

// 🚨 SECURITY: TODO!(sqs): there needs to be security checks everywhere here! there are none

// gqlThread implements the GraphQL type Thread.
type gqlThread struct {
	threads.GQLThreadlike
	db *dbThread
}

func newGQLThread(db *dbThread) *gqlThread {
	return &gqlThread{
		GQLThreadlike: threads.GQLThreadlike{
			DB:             db,
			PartialComment: comments.GraphQLResolver{}.LazyCommentByID(threads.MarshalID(threads.GQLTypeThread, db.ID)),
		},
		db: db,
	}
}

// threadByID looks up and returns the Thread with the given GraphQL ID. If no such Thread exists, it
// returns a non-nil error.
func threadByID(ctx context.Context, id graphql.ID) (*gqlThread, error) {
	dbID, err := threads.UnmarshalIDOfType(threads.GQLTypeThread, id)
	if err != nil {
		return nil, err
	}
	return threadByDBID(ctx, dbID)
}

func (GraphQLResolver) ThreadByID(ctx context.Context, id graphql.ID) (graphqlbackend.Thread, error) {
	return threadByID(ctx, id)
}

// threadByDBID looks up and returns the Thread with the given database ID. If no such Thread exists,
// it returns a non-nil error.
func threadByDBID(ctx context.Context, dbID int64) (*gqlThread, error) {
	v, err := dbThreads{}.GetByID(ctx, dbID)
	if err != nil {
		return nil, err
	}
	return newGQLThread(v), nil
}

func (GraphQLResolver) ThreadInRepository(ctx context.Context, repositoryID graphql.ID, number string) (graphqlbackend.Thread, error) {
	threadDBID, err := strconv.ParseInt(number, 10, 64)
	if err != nil {
		return nil, err
	}
	// TODO!(sqs): access checks
	thread, err := threadByDBID(ctx, threadDBID)
	if err != nil {
		return nil, err
	}

	// TODO!(sqs): check that the thread is indeed in the repo. When we make the thread number
	// sequence per-repo, this will become necessary to even retrieve the thread. for now, the ID is
	// global, so we need to perform this check.
	assertedRepo, err := graphqlbackend.RepositoryByID(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	if thread.db.RepositoryID != assertedRepo.DBID() {
		return nil, errors.New("thread does not exist in repository")
	}

	return thread, nil
}

func (v *gqlThread) State() graphqlbackend.ThreadState {
	return graphqlbackend.ThreadState(v.db.State)
}

func (v *gqlThread) BaseRef() string { return v.db.BaseRef }

func (v *gqlThread) HeadRef() string { return v.db.HeadRef }

func (v *gqlThread) IsPreview() bool { return v.db.IsPreview }

func (v *gqlThread) RepositoryComparison(ctx context.Context) (*graphqlbackend.RepositoryComparisonResolver, error) {
	repo, err := v.Repository(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlbackend.NewRepositoryComparison(ctx, repo, &graphqlbackend.RepositoryComparisonInput{
		Base: &v.db.BaseRef,
		Head: &v.db.HeadRef,
	})
}
