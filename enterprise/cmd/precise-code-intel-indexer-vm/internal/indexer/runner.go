package indexer

import (
	"context"
)

type Runner interface {
	Startup(ctx context.Context) error
	Teardown(ctx context.Context) error
	Invoke(ctx context.Context, image string, cs *CommandSpec) error
	MakeArgs(ctx context.Context, image string, cs *CommandSpec, mountPoint string) []string
}

type CommandSpec struct {
	command []string
	env     map[string]string
}

func FromArgs(args []string) *CommandSpec {
	cs := &CommandSpec{
		env: map[string]string{},
	}

	return cs.AddArgs(args...)
}

func (cs *CommandSpec) AddArgs(args ...string) *CommandSpec {
	cs.command = append(cs.command, args...)
	return cs
}

func (cs *CommandSpec) AddFlag(name, value string) *CommandSpec {
	cs.command = append(cs.command, name, value)
	return cs
}

func (cs *CommandSpec) AddEnv(name, value string) *CommandSpec {
	cs.env[name] = value
	return cs
}
