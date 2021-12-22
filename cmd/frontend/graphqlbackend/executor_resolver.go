package graphqlbackend

import (
	"context"
	"time"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type ExecutorResolver struct {
	executor types.Executor
}

// ExecutorByHostname returns an executor resolver for the given hostname, or
// nil when there is no executor record matching the given hostname.
//
// 🚨 SECURITY: This always returns nil for non-site admins.
func ExecutorByHostname(ctx context.Context, db database.DB, hostname string) (*ExecutorResolver, error) {
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
		if err != backend.ErrMustBeSiteAdmin {
			return nil, err
		}
		return nil, nil
	}

	e, found, err := db.Executors().GetByHostname(ctx, hostname)
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, nil
	}

	return &ExecutorResolver{executor: e}, nil
}

func (e *ExecutorResolver) ID() graphql.ID    { return marshalExecutorID(int64(e.executor.ID)) }
func (e *ExecutorResolver) Hostname() string  { return e.executor.Hostname }
func (e *ExecutorResolver) QueueName() string { return e.executor.QueueName }
func (e *ExecutorResolver) Active() bool {
	// TODO: Read the value of the executor worker heartbeat interval in here.
	heartbeatInterval := 5 * time.Second
	return time.Since(e.executor.LastSeenAt) <= 3*heartbeatInterval
}
func (e *ExecutorResolver) Os() string              { return e.executor.OS }
func (e *ExecutorResolver) Architecture() string    { return e.executor.Architecture }
func (e *ExecutorResolver) DockerVersion() string   { return e.executor.DockerVersion }
func (e *ExecutorResolver) ExecutorVersion() string { return e.executor.ExecutorVersion }
func (e *ExecutorResolver) GitVersion() string      { return e.executor.GitVersion }
func (e *ExecutorResolver) IgniteVersion() string   { return e.executor.IgniteVersion }
func (e *ExecutorResolver) SrcCliVersion() string   { return e.executor.SrcCliVersion }
func (e *ExecutorResolver) FirstSeenAt() DateTime   { return DateTime{e.executor.FirstSeenAt} }
func (e *ExecutorResolver) LastSeenAt() DateTime    { return DateTime{e.executor.LastSeenAt} }
