package graphql

import (
	"encoding/json"
	"regexp"
	"time"

	"github.com/Masterminds/semver"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ExecutorResolver struct {
	executor types.Executor
}

func NewExecutorResolver(executor Executor) *ExecutorResolver {
	return &ExecutorResolver{executor: executor}
}

func (e *ExecutorResolver) ID() graphql.ID {
	return relay.MarshalID("Executor", (int64(e.executor.ID)))
}
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

func (e *ExecutorResolver) IsOutdated() (bool, error) {
	ev := e.executor.ExecutorVersion
    if !e.Active() {
        return false, nil
    }
	return isExecutorOutdated(ev)
}

func isExecutorOutdated(ev string, active bool) (bool, error) {
	sv := version.Version()

	isExecutorDev := ev != "" && version.IsDev(ev)
	isSgDev := sv != "" && version.IsDev(sv)

	if !active || isSgDev || isExecutorDev {
		return false, nil
	}

	r := regexp.MustCompile(`^[\w-]+_(\d{4}-\d{2}-\d{2})_\w+`)
	evm := r.FindStringSubmatch(ev)
	svm := r.FindStringSubmatch(sv)
	if len(evm) > 1 && len(svm) > 1 {
		layout := "2006-01-02"

		st, err := time.Parse(layout, svm[1])
		if err != nil {
			return false, err
		}

		et, err := time.Parse(layout, evm[1])
		if err != nil {
			return false, err
		}

		return et.Before(st), nil
	}

	s, err := semver.NewVersion(sv)
	if err != nil {
		return false, err
	}

	e, err := semver.NewVersion(ev)
	if err != nil {
		return true, err
	}

	// it's okay for an executor to be one version behind the sourcegraph version.
	iev := e.IncMinor()

	return s.GreaterThan(&iev), nil
}

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

func (v *DateTime) UnmarshalGraphQL(input any) error {
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
