package threads

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/externallink"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// 🚨 SECURITY: TODO!(sqs): there needs to be security checks everywhere here! there are none

// gqlThread implements the GraphQL type Thread.
type gqlThread struct {
	db *DBThread
}

func newGQLThread(db *DBThread) *gqlThread {
	return &gqlThread{db: db}
}

// threadByID looks up and returns the Thread with the given GraphQL ID. If no such Thread exists, it
// returns a non-nil error.
func threadByID(ctx context.Context, id graphql.ID) (*gqlThread, error) {
	dbID, err := graphqlbackend.UnmarshalThreadID(id)
	if err != nil {
		return nil, err
	}
	return threadByDBID(ctx, dbID)
}

var MockThreadByID func(id graphql.ID) (graphqlbackend.Thread, error)

func (GraphQLResolver) ThreadByID(ctx context.Context, id graphql.ID) (graphqlbackend.Thread, error) {
	if MockThreadByID != nil {
		return MockThreadByID(id)
	}
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

func (v *gqlThread) ID() graphql.ID {
	return graphqlbackend.MarshalThreadID(v.db.ID)
}

func (v *gqlThread) Repository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	return graphqlbackend.RepositoryByDBID(ctx, v.db.RepositoryID)
}

func (v *gqlThread) Internal_RepositoryID() api.RepoID { return v.db.RepositoryID }

func (v *gqlThread) DBID() int64 { return v.db.ID }

func (v *gqlThread) Title() string { return v.db.Title }

func (v *gqlThread) State() graphqlbackend.ThreadState {
	return graphqlbackend.ThreadState(v.db.State)
}

func (v *gqlThread) BaseRef() *string {
	if v.db.BaseRef == "" {
		return nil
	}
	return &v.db.BaseRef
}

func (v *gqlThread) HeadRef() *string {
	if v.db.HeadRef == "" {
		return nil
	}
	return &v.db.HeadRef
}

func (v *gqlThread) Kind(ctx context.Context) (graphqlbackend.ThreadKind, error) {
	if v.db.BaseRef != "" || v.db.HeadRef != "" {
		return graphqlbackend.ThreadKindChangeset, nil
	}
	return graphqlbackend.ThreadKindIssue, nil
}

func (v *gqlThread) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: v.db.CreatedAt}
}

func (v *gqlThread) UpdatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: v.db.UpdatedAt}
}

func (v *gqlThread) ViewerCanUpdate(ctx context.Context) (bool, error) {
	return false, nil
}

func (v *gqlThread) ExternalURLs(ctx context.Context) ([]*externallink.Resolver, error) {
	// TODO!(sqs): make non-github-specific
	var githubData struct{ URL string }
	if err := json.Unmarshal(v.db.ExternalMetadata, &githubData); err != nil || githubData.URL == "" {
		return nil, nil
	}
	return []*externallink.Resolver{externallink.NewResolver(githubData.URL, "github")}, nil
}

func (v *gqlThread) RepositoryComparison(ctx context.Context) (graphqlbackend.RepositoryComparison, error) {
	if v.db.BaseRef == "" && v.db.HeadRef == "" {
		return nil, nil
	}
	repo, err := v.Repository(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlbackend.NewRepositoryComparison(ctx, repo, &graphqlbackend.RepositoryComparisonInput{
		Base:    &v.db.BaseRef,
		BaseOID: &v.db.BaseRefOID,
		Head:    &v.db.HeadRef,
		HeadOID: &v.db.HeadRefOID,
	})
}

func (v *gqlThread) Campaigns(ctx context.Context, arg *graphqlutil.ConnectionArgs) (graphqlbackend.CampaignConnection, error) {
	return graphqlbackend.CampaignsWithObject(ctx, v.ID(), arg)
}

func (v *gqlThread) Labels(ctx context.Context, arg *graphqlutil.ConnectionArgs) (graphqlbackend.LabelConnection, error) {
	// TODO!(sqs): serialize labels from external service and use ConstLabelConnection here
	return graphqlbackend.ConstLabelConnection(nil), nil
}

func (v *gqlThread) Assignees(ctx context.Context, arg *graphqlutil.ConnectionArgs) (graphqlbackend.ActorConnection, error) {
	actor, err := v.db.Assignee.GQL(ctx)
	if err != nil {
		return nil, err
	}
	if actor == nil {
		return graphqlbackend.ActorConnection{}, nil
	}
	return graphqlbackend.ActorConnection{*actor}, nil
}
