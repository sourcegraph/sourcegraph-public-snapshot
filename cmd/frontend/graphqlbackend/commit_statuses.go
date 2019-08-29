package graphqlbackend

import (
	"context"
	"errors"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// CommitStatuses is the implementation of the GraphQL API for commit status-related queries and
// mutations. If it is not set at runtime, a "not implemented" error is returned to API clients who
// invoke it.
//
// This is contributed by enterprise.
var CommitStatuses interface {
	// CommitStatusContextByID is called by the CommitStatusContextByID func but is not in the
	// GraphQL API.
	CommitStatusContextByID(context.Context, graphql.ID) (CommitStatusContext, error)

	// CommitStatusForCommit is called by the CommitStatusForCommit func but is not in the GraphQL
	// API.
	CommitStatusForCommit(ctx context.Context, repository graphql.ID, commitID api.CommitID) (CommitStatus, error)
}

const (
	GQLTypeCommitStatus        = "CommitStatus"
	GQLTypeCommitStatusContext = "CommitStatusContext"
)

func MarshalCommitStatusContextID(id int64) graphql.ID {
	return relay.MarshalID(GQLTypeCommitStatusContext, id)
}

func UnmarshalCommitStatusContextID(id graphql.ID) (dbID int64, err error) {
	if typ := relay.UnmarshalKind(id); typ != GQLTypeCommitStatusContext {
		return 0, fmt.Errorf("CommitStatusContext ID has unexpected type type %q", typ)
	}
	err = relay.UnmarshalSpec(id, &dbID)
	return
}

var errCommitStatusesNotImplemented = errors.New("commit statuses are not implemented")

// CommitStatusContextByID is called to look up a CommitStatusContext given its GraphQL ID.
func CommitStatusContextByID(ctx context.Context, id graphql.ID) (CommitStatusContext, error) {
	if CommitStatuses == nil {
		return nil, errCommitStatusesNotImplemented
	}
	return CommitStatuses.CommitStatusContextByID(ctx, id)
}

// CommitStatusForCommit returns the combined CommitStatus for the specified repository and commit.
func CommitStatusContextsForCommit(ctx context.Context, repository graphql.ID, commitID api.CommitID) (CommitStatus, error) {
	if CommitStatuses == nil {
		return nil, errCommitStatusesNotImplemented
	}
	return CommitStatuses.CommitStatusForCommit(ctx, repository, commitID)
}

type CommitStatusState string

const (
	CommitStatusStateExpected CommitStatusState = "EXPECTED"
	CommitStatusStateError                      = "ERROR"
	CommitStatusStateFailure                    = "FAILURE"
	CommitStatusStatePending                    = "PENDING"
	CommitStatusStateSuccess                    = "SUCCESS"
)

// CommitStatus is the interface for the GraphQL type CommitStatus.
type CommitStatus interface {
	Repository(context.Context) (*RepositoryResolver, error)
	Commit(context.Context) (*GitCommitResolver, error)
	Contexts(context.Context) ([]CommitStatusContext, error)
	State() CommitStatusState
}

// CommitStatusContext is the interface for the GraphQL type CommitStatusContext.
type CommitStatusContext interface {
	ID() graphql.ID
	DBID() int64
	Repository(context.Context) (*RepositoryResolver, error)
	Commit(context.Context) (*GitCommitResolver, error)
	Context() string
	State() CommitStatusState
	Description() *string
	TargetURL() *string
	Actor(context.Context) (*Actor, error)
	CreatedAt() DateTime
}
