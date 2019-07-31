package threads

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

// gqlThread implements the GraphQL type Thread.
type gqlThread struct {
	threadlike.GQLThreadlike
	db *internal.DBThread
}

func newGQLThread(db *internal.DBThread) *gqlThread {
	return &gqlThread{
		GQLThreadlike: threadlike.GQLThreadlike{
			Comment: &comments.GQLIComment{},
			DB:      db,
		},
		db: db,
	}
}

// threadByID looks up and returns the Thread with the given GraphQL ID. If no such Thread exists, it
// returns a non-nil error.
func threadByID(ctx context.Context, id graphql.ID) (*gqlThread, error) {
	dbID, err := threadlike.UnmarshalIDOfType(threadlike.GQLTypeThread, id)
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
	v, err := internal.DBThreads{}.GetByID(ctx, dbID)
	if err != nil {
		return nil, err
	}
	return newGQLThread(v), nil
}

func (v *gqlThread) ID() graphql.ID {
	return threadlike.MarshalID(threadlike.GQLTypeThread, v.db.ID)
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

func (v *gqlThread) Status() graphqlbackend.ThreadStatus {
	return graphqlbackend.ThreadStatus(v.db.Status)
}
