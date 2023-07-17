package graphqlbackend

import (
	"time"

	"github.com/Masterminds/semver"
	"github.com/grafana/regexp"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const oneReleaseCycle = 35 * 24 * time.Hour

var insiderBuildRegex = regexp.MustCompile(`^[\w-]+_(\d{4}-\d{2}-\d{2})_(\d+\.\d+-)?\w+`)

func NewExecutorResolver(executor types.Executor) *ExecutorResolver {
	return &ExecutorResolver{executor: executor}
}

type ExecutorResolver struct {
	executor types.Executor
}

func (e *ExecutorResolver) ID() graphql.ID {
	return relay.MarshalID("Executor", int64(e.executor.ID))
}
func (e *ExecutorResolver) Hostname() string { return e.executor.Hostname }
func (e *ExecutorResolver) QueueName() *string {
	queueName := e.executor.QueueName
	if queueName == "" {
		return nil
	}
	return &queueName
}
func (e *ExecutorResolver) QueueNames() *[]string {
	return &e.executor.QueueNames
}
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
func (e *ExecutorResolver) FirstSeenAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: e.executor.FirstSeenAt}
}
func (e *ExecutorResolver) LastSeenAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: e.executor.LastSeenAt}
}

func (e *ExecutorResolver) Compatibility() (*string, error) {
	ev := e.executor.ExecutorVersion
	if !e.Active() {
		return nil, nil
	}
	return calculateExecutorCompatibility(ev)
}

func calculateExecutorCompatibility(ev string) (*string, error) {
	compatibility := ExecutorCompatibilityUpToDate
	sv := version.Version()

	isExecutorDev := ev != "" && version.IsDev(ev)
	isSgDev := sv != "" && version.IsDev(sv)

	if isSgDev || isExecutorDev {
		return nil, nil
	}

	evm := insiderBuildRegex.FindStringSubmatch(ev)
	svm := insiderBuildRegex.FindStringSubmatch(sv)

	// check for version mismatch
	if len(evm) > 1 && len(svm) <= 1 {
		// this means that the executor is an insider version while the Sourcegraph
		// instance is not.
		return nil, nil
	}

	if len(evm) <= 1 && len(svm) > 1 {
		// this means that the Sourcegraph instance is an insider version while the
		// executor is not.
		return nil, nil
	}

	if len(evm) > 1 && len(svm) > 1 {
		layout := "2006-01-02"

		st, err := time.Parse(layout, svm[1])
		if err != nil {
			return nil, err
		}

		et, err := time.Parse(layout, evm[1])
		if err != nil {
			return nil, err
		}

		hst := st.Add(oneReleaseCycle)
		lst := st.Add(-1 * oneReleaseCycle)

		if et.After(hst) {
			// We check if the executor build date is after a release cycle + sourcegraph build date.
			// if this is true then we assume the executor's version is ahead.
			compatibility = ExecutorCompatibilityVersionAhead
		} else if et.Before(lst) {
			// if the executor date is a release cycle behind the current build date of the Sourcegraph
			// instance then we assume that the executor is outdated.
			compatibility = ExecutorCompatibilityOutdated
		}

		return compatibility.ToGraphQL(), nil
	}

	s, err := getSemVer("sourcegraph", sv)
	if err != nil {
		return nil, err
	}

	e, err := getSemVer("executor", ev)
	if err != nil {
		return nil, err
	}

	// it's okay for an executor to be one minor version behind or ahead of the sourcegraph version.
	iev := e.IncMinor()

	isv := s.IncMinor()

	if s.GreaterThan(&iev) {
		compatibility = ExecutorCompatibilityOutdated
	} else if isv.LessThan(e) {
		compatibility = ExecutorCompatibilityVersionAhead
	}

	return compatibility.ToGraphQL(), nil
}

func getSemVer(source string, version string) (*semver.Version, error) {
	v, err := semver.NewVersion(version)
	if err != nil {
		// Maybe the version is a daily build and need to extract the version from there.
		// We don't care about the error from getDailyBuildVersion because we already have the error.
		v, _ = getDailyBuildVersion(version)
		if v == nil {
			return nil, errors.Wrapf(err, "failed to parse %s version %q", source, version)
		}
	}
	return v, nil
}

func getDailyBuildVersion(version string) (*semver.Version, error) {
	matches := api.BuildDateRegex.FindStringSubmatch(version)
	if len(matches) > 2 {
		return semver.NewVersion(matches[2])
	}
	return nil, nil
}
