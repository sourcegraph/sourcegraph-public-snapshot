// Package buildkite defines data types that reflect Buildkite's YAML pipeline format.
//
// Usage:
//
//    pipeline := buildkite.Pipeline{}
//    pipeline.AddStep("check_mark", buildkite.Cmd("./dev/check/all.sh"))
package buildkite

import (
	"io"
	"strings"

	"github.com/ghodss/yaml"
)

type Pipeline struct {
	Steps []interface{} `json:"steps"`
}

type BuildOptions struct {
	Message  string                 `json:"message,omitempty"`
	Commit   string                 `json:"commit,omitempty"`
	MetaData map[string]interface{} `json:"meta_data,omitempty"`
	Env      map[string]string      `json:"env,omitempty"`
}

type Step struct {
	Label            string                 `json:"label"`
	Command          string                 `json:"command,omitempty"`
	Trigger          string                 `json:"trigger,omitempty"`
	Async            bool                   `json:"async,omitempty"`
	Build            *BuildOptions          `json:"build,omitempty"`
	Env              map[string]string      `json:"env,omitempty"`
	Plugins          map[string]interface{} `json:"plugins,omitempty"`
	ArtifactPaths    string                 `json:"artifact_paths,omitempty"`
	ConcurrencyGroup string                 `json:"concurrency_group,omitempty"`
	Concurrency      int                    `json:"concurrency,omitempty"`
}

var Plugins = make(map[string]interface{})

// OnEveryStepOpts are e.g. commands that are run on every AddStep, similar to
// Plugins.
var OnEveryStepOpts []StepOpt

func (p *Pipeline) AddStep(label string, opts ...StepOpt) {
	step := &Step{
		Label:   label,
		Env:     make(map[string]string),
		Plugins: Plugins,
	}
	for _, opt := range OnEveryStepOpts {
		opt(step)
	}
	for _, opt := range opts {
		opt(step)
	}
	p.Steps = append(p.Steps, step)
}

func (p *Pipeline) AddTrigger(label string, opts ...StepOpt) {
	step := &Step{
		Label: label,
	}
	for _, opt := range OnEveryStepOpts {
		opt(step)
	}
	for _, opt := range opts {
		opt(step)
	}
	p.Steps = append(p.Steps, step)
}

func (p *Pipeline) WriteTo(w io.Writer) (int64, error) {
	output, err := yaml.Marshal(p)
	if err != nil {
		return 0, err
	}

	n, err := w.Write(output)
	return int64(n), err
}

type StepOpt func(step *Step)

func Cmd(command string) StepOpt {
	return func(step *Step) {
		step.Command = strings.TrimSpace(step.Command + "\n" + command)
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

func Env(name, value string) StepOpt {
	return func(step *Step) {
		step.Env[name] = value
	}
}

func ArtifactPaths(paths string) StepOpt {
	return func(step *Step) {
		step.ArtifactPaths = paths
	}
}

func (p *Pipeline) AddWait() {
	p.Steps = append(p.Steps, "wait")
}
