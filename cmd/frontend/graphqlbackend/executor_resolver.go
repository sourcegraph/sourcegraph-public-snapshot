package graphqlbackend

import (
	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

type executorResolver struct {
	executor types.Executor
}

func (e *executorResolver) ID() graphql.ID          { return marshalExecutorID(int64(e.executor.ID)) }
func (e *executorResolver) Hostname() string        { return e.executor.Hostname }
func (e *executorResolver) QueueName() string       { return e.executor.QueueName }
func (e *executorResolver) Os() string              { return e.executor.OS }
func (e *executorResolver) Architecture() string    { return e.executor.Architecture }
func (e *executorResolver) ExecutorVersion() string { return e.executor.ExecutorVersion }
func (e *executorResolver) SrcCliVersion() string   { return e.executor.SrcCliVersion }
func (e *executorResolver) DockerVersion() string   { return e.executor.DockerVersion }
func (e *executorResolver) GitVersion() string      { return e.executor.GitVersion }
func (e *executorResolver) IgniteVersion() string   { return e.executor.IgniteVersion }
func (e *executorResolver) FirstSeenAt() DateTime   { return DateTime{e.executor.FirstSeenAt} }
func (e *executorResolver) LastSeenAt() DateTime    { return DateTime{e.executor.LastSeenAt} }
