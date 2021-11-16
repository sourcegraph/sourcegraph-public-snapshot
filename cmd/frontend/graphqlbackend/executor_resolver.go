package graphqlbackend

import (
	"time"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

type executorResolver struct {
	executor types.Executor
}

func (e *executorResolver) ID() graphql.ID    { return marshalExecutorID(int64(e.executor.ID)) }
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
func (e *executorResolver) FirstSeenAt() DateTime   { return DateTime{e.executor.FirstSeenAt} }
func (e *executorResolver) LastSeenAt() DateTime    { return DateTime{e.executor.LastSeenAt} }
