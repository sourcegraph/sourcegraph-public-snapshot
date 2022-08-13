package resolvers

import (
	"context"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type executorResolver struct {
	executor types.Executor
}

func NewExecutorResolver(executor types.Executor) graphqlbackend.ExecutorResolver {
	return &executorResolver{executor}
}

func (e *executorResolver) ID() graphql.ID {
	return relay.MarshalID("Executor", (int64(e.executor.ID)))
}
func (e *executorResolver) Hostname() string  { return e.executor.Hostname }
func (e *executorResolver) QueueName() string { return e.executor.QueueName }
func (e *executorResolver) Active() bool {
	// TODO: Read the value of the executor worker heartbeat interval in here.
	heartbeatInterval := 5 * time.Second
	return time.Since(e.executor.LastSeenAt) <= 3*heartbeatInterval
}
func (e *executorResolver) Os() string              { return e.executor.OS }
func (e *executorResolver) Architecture() string    { return e.executor.Architecture }
func (e *executorResolver) DockerVersion() string   { return e.executor.DockerVersion }
func (e *executorResolver) ExecutorVersion() string { return e.executor.ExecutorVersion }
func (e *executorResolver) GitVersion() string      { return e.executor.GitVersion }
func (e *executorResolver) IgniteVersion() string   { return e.executor.IgniteVersion }
func (e *executorResolver) SrcCliVersion() string   { return e.executor.SrcCliVersion }

func (e *executorResolver) FirstSeenAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: e.executor.FirstSeenAt}
}

func (e *executorResolver) LastSeenAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: e.executor.LastSeenAt}
}

func (e *executorResolver) ActiveJobs(
	ctx context.Context,
	args graphqlbackend.ExecutorActiveJobArgs,
) (graphqlbackend.ExecutorJobConnectionResolver, error) {
	return nil, errors.New("not implemented")
}
