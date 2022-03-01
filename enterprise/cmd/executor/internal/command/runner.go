package command

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Runner is the interface between an executor and the host on which commands
// are invoked. Having this interface at this level allows us to use the same
// code paths for local development (via shell + docker) as well as production
// usage (via Firecracker).
type Runner interface {
	// Setup prepares the runner to invoke a series of commands.
	Setup(ctx context.Context) error

	// Teardown disposes of any resources created in Setup.
	Teardown(ctx context.Context) error

	// Run invokes a command in the runner context.
	Run(ctx context.Context, command CommandSpec) error
}

// CommandSpec represents a command that can be run on a machine, whether that
// is the host, in a virtual machine, or in a docker container. If an image is
// supplied, then the command will be run in a one-shot docker container.
type CommandSpec struct {
	Key        string
	Image      string
	ScriptPath string
	Command    []string
	Dir        string
	Env        []string
	Operation  *observation.Operation
}

type Options struct {
	// ExecutorName is a unique identifier for the requesting executor.
	ExecutorName string

	// FirecrackerOptions configures the behavior of Firecracker virtual machine creation.
	FirecrackerOptions FirecrackerOptions

	// ResourceOptions configures the resource limits of docker container and Firecracker
	// virtual machines running on the executor.
	ResourceOptions ResourceOptions
}

type FirecrackerOptions struct {
	// Enabled determines if commands will be run in Firecracker virtual machines.
	Enabled bool

	// Image is the base image used for all Firecracker virtual machines.
	Image string

	// VMStartupScriptPath is a path to a file on the host that is loaded into a fresh
	// virtual machine and executed on startup.
	VMStartupScriptPath string
}

type ResourceOptions struct {
	// NumCPUs is the number of virtual CPUs a container or VM can use.
	NumCPUs int

	// Memory is the maximum amount of memory a container or VM can use.
	Memory string

	// DiskSpace is the maximum amount of disk a container or VM can use.
	DiskSpace string
}

// NewRunner creates a new runner with the given options.
func NewRunner(dir string, logger *Logger, options Options, operations *Operations) Runner {
	if !options.FirecrackerOptions.Enabled {
		return &dockerRunner{dir: dir, logger: logger, options: options}
	}

	return &firecrackerRunner{
		name:       options.ExecutorName,
		dir:        dir,
		logger:     logger,
		options:    options,
		operations: operations,
	}
}

type dockerRunner struct {
	dir     string
	logger  *Logger
	options Options
}

var _ Runner = &dockerRunner{}

func (r *dockerRunner) Setup(ctx context.Context) error {
	return nil
}

func (r *dockerRunner) Teardown(ctx context.Context) error {
	return nil
}

func (r *dockerRunner) Run(ctx context.Context, command CommandSpec) error {
	return runCommand(ctx, formatRawOrDockerCommand(command, r.dir, r.options), r.logger)
}

type firecrackerRunner struct {
	name       string
	dir        string
	logger     *Logger
	options    Options
	operations *Operations
}

var _ Runner = &firecrackerRunner{}

func (r *firecrackerRunner) Setup(ctx context.Context) error {
	return setupFirecracker(ctx, defaultRunner, r.logger, r.name, r.dir, r.options, r.operations)
}

func (r *firecrackerRunner) Teardown(ctx context.Context) error {
	return teardownFirecracker(ctx, defaultRunner, r.logger, r.name, r.operations)
}

func (r *firecrackerRunner) Run(ctx context.Context, command CommandSpec) error {
	return runCommand(ctx, formatFirecrackerCommand(command, r.name, r.options), r.logger)
}

type runnerWrapper struct{}

var defaultRunner = &runnerWrapper{}

func (runnerWrapper) RunCommand(ctx context.Context, command command, logger *Logger) error {
	return runCommand(ctx, command, logger)
}
