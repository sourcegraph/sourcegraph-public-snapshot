package threads

import (
	"context"
	"path"
	"strconv"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

// 🚨 SECURITY: TODO!(sqs): there needs to be security checks everywhere here! there are none

// gqlThread implements the GraphQL type Thread.
type gqlThread struct{ db *dbThread }

// threadByID looks up and returns the Thread with the given GraphQL ID. If no such Thread exists, it
// returns a non-nil error.
func threadByID(ctx context.Context, id graphql.ID) (*gqlThread, error) {
	dbID, err := unmarshalThreadID(id)
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
	return &gqlThread{db: v}, nil
}

func (v *gqlThread) ID() graphql.ID {
	return marshalThreadID(v.db.ID)
}

func marshalThreadID(id int64) graphql.ID {
	return relay.MarshalID("Thread", id)
}

func unmarshalThreadID(id graphql.ID) (dbID int64, err error) {
	err = relay.UnmarshalSpec(id, &dbID)
	return
}

func (GraphQLResolver) ThreadInRepository(ctx context.Context, repositoryID graphql.ID, threadIDInRepository string) (graphqlbackend.Thread, error) {
	threadID, err := strconv.ParseInt(threadIDInRepository, 10, 64)
	if err != nil {
		return nil, err
	}
	// TODO!(sqs): access checks
	thread, err := threadByDBID(ctx, threadID)
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

func (v *gqlThread) IDInRepository() string { return strconv.FormatInt(v.db.ID, 10) }

func (v *gqlThread) DBID() int64 { return v.db.ID }

func (v *gqlThread) Repository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	return graphqlbackend.RepositoryByDBID(ctx, v.db.RepositoryID)
}

func (v *gqlThread) Title() string { return v.db.Title }

func (v *gqlThread) ExternalURL() *string { return v.db.ExternalURL }

func (v *gqlThread) URL(ctx context.Context) (string, error) {
	repository, err := v.Repository(ctx)
	if err != nil {
		return "", err
	}
	return path.Join(repository.URL(), "-", "threads", v.IDInRepository()), nil
}

func (v *gqlThread) Settings() string {
	if settings := v.db.Settings; settings != nil {
		return *settings
	}
	return "{}"
}

func (v *gqlThread) Status() graphqlbackend.ThreadStatus {
	return v.db.Status
}

func (v *gqlThread) Type() graphqlbackend.ThreadType {
	return v.db.Type
}
