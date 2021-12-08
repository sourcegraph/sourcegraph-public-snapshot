package graphqlbackend

import (
	"encoding/json"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

type executorResolver struct {
	executor types.Executor
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
func (e *executorResolver) FirstSeenAt() DateTime   { return DateTime{e.executor.FirstSeenAt} }
func (e *executorResolver) LastSeenAt() DateTime    { return DateTime{e.executor.LastSeenAt} }

// TODO: abstract this out into a common package

// DateTime implements the DateTime GraphQL scalar type.
type DateTime struct{ time.Time }

// DateTimeOrNil is a helper function that returns nil for time == nil and otherwise wraps time in
// DateTime.
func DateTimeOrNil(time *time.Time) *DateTime {
	if time == nil {
		return nil
	}
	return &DateTime{Time: *time}
}

func (DateTime) ImplementsGraphQLType(name string) bool {
	return name == "DateTime"
}

func (v DateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Time.Format(time.RFC3339))
}

func (v *DateTime) UnmarshalGraphQL(input interface{}) error {
	s, ok := input.(string)
	if !ok {
		return errors.Errorf("invalid GraphQL DateTime scalar value input (got %T, expected string)", input)
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	*v = DateTime{Time: t}
	return nil
}
