// Package buildkite defines data types that reflect Buildkite's YAML pipeline format.
//
// Usage:
//
//    pipeline := buildkite.Pipeline{}
//    pipeline.AddStep("check_mark", buildkite.Cmd("./dev/check/all.sh"))
package buildkite

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/ghodss/yaml"
)

type Pipeline struct {
	Env   map[string]string `json:"env,omitempty"`
	Steps []interface{}     `json:"steps"`
}

type BuildOptions struct {
	Message  string                 `json:"message,omitempty"`
	Commit   string                 `json:"commit,omitempty"`
	Branch   string                 `json:"branch,omitempty"`
	MetaData map[string]interface{} `json:"meta_data,omitempty"`
	Env      map[string]string      `json:"env,omitempty"`
}

func (bo BuildOptions) MarshalJSON() ([]byte, error) {
	type buildOptions BuildOptions
	boCopy := buildOptions(bo)
	// Buildkite pipeline upload command will interpolate if it sees a $var
	// which can cause the pipeline generation to fail because that
	// variable do not exists.
	// By replacing $ into $$ in the commit messages we can prevent those
	// failures to happen.
	//
	// https://buildkite.com/docs/agent/v3/cli-pipeline#environment-variable-substitution
	boCopy.Message = strings.ReplaceAll(boCopy.Message, "$", `$$`)
	return json.Marshal(boCopy)
}

func (bo BuildOptions) MarshalYAML() ([]byte, error) {
	type buildOptions BuildOptions
	boCopy := buildOptions(bo)
	// Buildkite pipeline upload command will interpolate if it sees a $var
	// which can cause the pipeline generation to fail because that
	// variable do not exists.
	// By replacing $ into $$ in the commit messages we can prevent those
	// failures to happen.
	//
	// https://buildkite.com/docs/agent/v3/cli-pipeline#environment-variable-substitution
	boCopy.Message = strings.ReplaceAll(boCopy.Message, "$", `$$`)
	return yaml.Marshal(boCopy)
}

// Matches Buildkite pipeline JSON schema:
// https://github.com/buildkite/pipeline-schema/blob/master/schema.json
type Step struct {
	Label                  string                 `json:"label"`
	Key                    string                 `json:"key,omitempty"`
	Command                []string               `json:"command,omitempty"`
	DependsOn              []string               `json:"depends_on,omitempty"`
	AllowDependencyFailure bool                   `json:"allow_dependency_failure,omitempty"`
	TimeoutInMinutes       string                 `json:"timeout_in_minutes,omitempty"`
	Trigger                string                 `json:"trigger,omitempty"`
	Async                  bool                   `json:"async,omitempty"`
	Build                  *BuildOptions          `json:"build,omitempty"`
	Env                    map[string]string      `json:"env,omitempty"`
	Plugins                map[string]interface{} `json:"plugins,omitempty"`
	ArtifactPaths          string                 `json:"artifact_paths,omitempty"`
	ConcurrencyGroup       string                 `json:"concurrency_group,omitempty"`
	Concurrency            int                    `json:"concurrency,omitempty"`
	Parallelism            int                    `json:"parallelism,omitempty"`
	Skip                   string                 `json:"skip,omitempty"`
	SoftFail               []softFailExitStatus   `json:"soft_fail,omitempty"`
	Retry                  *RetryOptions          `json:"retry,omitempty"`
	Agents                 map[string]string      `json:"agents,omitempty"`
}

var nonAlphaNumeric = regexp.MustCompile("[^a-zA-Z0-9]+")

// GenerateKey will automatically generate a key based on the
// step label, and return it.
func (s *Step) GenerateKey() string {
	s.Key = nonAlphaNumeric.ReplaceAllString(s.Label, "")
	return s.Key
}

type RetryOptions struct {
	Automatic *AutomaticRetryOptions `json:"automatic,omitempty"`
	Manual    *ManualRetryOptions    `json:"manual,omitempty"`
}

type AutomaticRetryOptions struct {
	Limit int `json:"limit,omitempty"`
}

type ManualRetryOptions struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

// BeforeEveryStepOpts are e.g. commands that are run before every AddStep, similar to
// Plugins.
var BeforeEveryStepOpts []StepOpt

// AfterEveryStepOpts are e.g. that are run at the end of every AddStep, helpful for
// post-processing
var AfterEveryStepOpts []StepOpt

func (p *Pipeline) AddStep(label string, opts ...StepOpt) {
	step := &Step{
		Label:   label,
		Env:     make(map[string]string),
		Agents:  make(map[string]string),
		Plugins: make(map[string]interface{}),
	}
	for _, opt := range BeforeEveryStepOpts {
		opt(step)
	}
	for _, opt := range opts {
		opt(step)
	}
	for _, opt := range AfterEveryStepOpts {
		opt(step)
	}

	if step.Key == "" {
		step.GenerateKey()
	}

	// Set a default agent queue to assign this job to
	if len(step.Agents) == 0 {
		step.Agents["queue"] = "standard"
	}

	p.Steps = append(p.Steps, step)
}

// AddEnsureStep adds a step that has a dependency on all other steps prior to this step,
// up until a wait step.
//
// We do not go past the closest "wait" because it won't work anyway - a failure before a
// "wait" will not allow this step to run.
func (p *Pipeline) AddEnsureStep(label string, opts ...StepOpt) {
	// Collect all keys to make this step depends on all others, traversing in reverse
	// until we reach a "wait", if there is one.
	keys := []string{}
	for i := len(p.Steps) - 1; i >= 0; i-- {
		step := p.Steps[i]
		if s, ok := step.(*Step); ok {
			keys = append(keys, s.Key)
		} else if wait, ok := step.(string); ok {
			if wait == "wait" {
				break // we are done
			}
		}
	}
	p.AddStep(label, append(opts, DependsOn(keys...))...)
}

func (p *Pipeline) AddTrigger(label string, opts ...StepOpt) {
	step := &Step{
		Label: label,
	}
	for _, opt := range opts {
		opt(step)
	}
	if step.Key == "" {
		step.GenerateKey()
	}
	p.Steps = append(p.Steps, step)
}

func (p *Pipeline) WriteJSONTo(w io.Writer) (int64, error) {
	output, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return 0, err
	}
	n, err := w.Write([]byte(output))
	return int64(n), err
}

func (p *Pipeline) WriteYAMLTo(w io.Writer) (int64, error) {
	output, err := yaml.Marshal(p)
	if err != nil {
		return 0, err
	}
	n, err := w.Write([]byte(output))
	return int64(n), err
}

type StepOpt func(step *Step)

func Cmd(command string) StepOpt {
	return func(step *Step) {
		// step.Command = append(step.Command, command)

		// BUILDKITE_STEP_ID is prefixed by $$ so it's not interpolated at pipeline generation time
		// but instead when the step is executed.
		tracedCmd := fmt.Sprintf(`(
  BUILDEVENT_APIKEY="$$CI_BUILDEVENT_API_KEY"
  BUILDEVENT_DATASET="buildkite"
  export BUILDEVENT_APIKEY
  export BUILDEVENT_DATASET
  ./buildevents cmd $BUILDKITE_BUILD_ID $$BUILDKITE_STEP_ID '%s' -- %s
)
`, command, command)
		step.Command = append(step.Command, tracedCmd)
	}
}

func Trigger(pipeline string) StepOpt {
	return func(step *Step) {
		step.Trigger = pipeline
	}
}

func Async(async bool) StepOpt {
	return func(step *Step) {
		step.Async = async
	}
}

func Build(buildOptions BuildOptions) StepOpt {
	return func(step *Step) {
		step.Build = &buildOptions
	}
}

func ConcurrencyGroup(group string) StepOpt {
	return func(step *Step) {
		step.ConcurrencyGroup = group
	}
}

func Concurrency(limit int) StepOpt {
	return func(step *Step) {
		step.Concurrency = limit
	}
}

// Parallelism tells Buildkite to run this job multiple time in parallel,
// which is very useful to QA a flake fix.
func Parallelism(count int) StepOpt {
	return func(step *Step) {
		step.Parallelism = count
	}
}

func Env(name, value string) StepOpt {
	return func(step *Step) {
		step.Env[name] = value
	}
}

func Skip(reason string) StepOpt {
	return func(step *Step) {
		step.Skip = reason
	}
}

type softFailExitStatus struct {
	ExitStatus int `json:"exit_status"`
}

// SoftFail indicates the specified exit codes should trigger a soft fail.
// https://buildkite.com/docs/pipelines/command-step#command-step-attributes
// This function also adds a specific env var named SOFT_FAIL_EXIT_CODES, enabling
// to get exit codes from the scripts until https://github.com/sourcegraph/sourcegraph/issues/27264
// is fixed.
func SoftFail(exitCodes ...int) StepOpt {
	return func(step *Step) {
		var codes []string
		for _, code := range exitCodes {
			codes = append(codes, strconv.Itoa(code))
			step.SoftFail = append(step.SoftFail, softFailExitStatus{
				ExitStatus: code,
			})
		}
		// https://github.com/sourcegraph/sourcegraph/issues/27264
		step.Env["SOFT_FAIL_EXIT_CODES"] = strings.Join(codes, " ")
	}
}

// AutomaticRetry enables automatic retry for the step with the number of times this job can be retried.
// The maximum value this can be set to is 10.
// Docs: https://buildkite.com/docs/pipelines/command-step#automatic-retry-attributes
func AutomaticRetry(limit int) StepOpt {
	return func(step *Step) {
		step.Retry = &RetryOptions{
			Automatic: &AutomaticRetryOptions{
				Limit: limit,
			},
		}
	}
}

// DisableManualRetry disables manual retry for the step. The reason string passed
// will be displayed in a tooltip on the Retry button in the Buildkite interface.
// Docs: https://buildkite.com/docs/pipelines/command-step#manual-retry-attributes
func DisableManualRetry(reason string) StepOpt {
	return func(step *Step) {
		step.Retry = &RetryOptions{
			Manual: &ManualRetryOptions{
				Allowed: false,
				Reason:  reason,
			},
		}
	}
}

func ArtifactPaths(paths ...string) StepOpt {
	return func(step *Step) {
		step.ArtifactPaths = strings.Join(paths, ";")
	}
}

func Agent(key, value string) StepOpt {
	return func(step *Step) {
		step.Agents[key] = value
	}
}

func (p *Pipeline) AddWait() {
	p.Steps = append(p.Steps, "wait")
}

func Key(key string) StepOpt {
	return func(step *Step) {
		step.Key = key
	}
}

func Plugin(name string, plugin interface{}) StepOpt {
	return func(step *Step) {
		step.Plugins[name] = plugin
	}
}

func DependsOn(dependency ...string) StepOpt {
	return func(step *Step) {
		step.DependsOn = append(step.DependsOn, dependency...)
	}
}

// AllowDependencyFailure enables `allow_dependency_failure` attribute on the step.
// Such a step will run when the depended-on jobs complete, fail or even did not run.
// See extended docs here: https://buildkite.com/docs/pipelines/dependencies#allowing-dependency-failures
func AllowDependencyFailure() StepOpt {
	return func(step *Step) {
		step.AllowDependencyFailure = true
	}
}
