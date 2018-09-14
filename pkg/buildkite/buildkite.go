package buildkite

import (
	"io"
	"strings"

	"github.com/ghodss/yaml"
)

type Pipeline struct {
	Steps []interface{} `json:"steps"`
}

type Step struct {
	Label            string                 `json:"label"`
	Command          string                 `json:"command"`
	Env              map[string]string      `json:"env"`
	Plugins          map[string]interface{} `json:"plugins"`
	ArtifactPaths    string                 `json:"artifact_paths,omitempty"`
	ConcurrencyGroup string                 `json:"concurrency_group,omitempty"`
	Concurrency      int                    `json:"concurrency,omitempty"`
}

var Plugins = make(map[string]interface{})

func (p *Pipeline) AddStep(label string, opts ...StepOpt) {
	step := &Step{
		Label:   label,
		Env:     make(map[string]string),
		Plugins: Plugins,
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
