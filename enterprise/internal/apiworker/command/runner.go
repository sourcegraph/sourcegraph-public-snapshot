package command

import "context"

// Runner is the interface between an executor and the host on which commands
// are invoked. Having this interface at this level allows us to use the same
// code paths for local development (via shell + docker) as well as production
// usage (via Firecracker).
type Runner interface {
	// Setup prepares the runner to invoke a series of commands.
	Setup(ctx context.Context, images []string) error

	// Teardown disposes of any resources created in Setup.
	Teardown(ctx context.Context) error

	// Run invokes a command in the runner context.
	Run(ctx context.Context, command CommandSpec) error
}

// CommandSpec represents a command that can be run on a machine, whether that
// is the host, in a virtual machine, or in a docker container. If an image is
// supplied, then the command will be run in a one-shot docker container.
type CommandSpec struct {
	Key      string
	Image    string
	Commands []string
	Dir      string
	Env      []string
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

	// ImageArchivesPath is a path on the host where docker image tarfiles will be stored.
	ImageArchivesPath string
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
func NewRunner(dir string, logger *Logger, options Options) Runner {
	if !options.FirecrackerOptions.Enabled {
		return &dockerRunner{dir: dir, logger: logger, options: options}
	}

	return &firecrackerRunner{name: options.ExecutorName, dir: dir, logger: logger, options: options}
}

type dockerRunner struct {
	dir     string
	logger  *Logger
	options Options
}

var _ Runner = &dockerRunner{}

func (r *dockerRunner) Setup(ctx context.Context, images []string) error {
	return nil
}

func (r *dockerRunner) Teardown(ctx context.Context) error {
	return nil
}

func (r *dockerRunner) Run(ctx context.Context, command CommandSpec) error {
	return runCommand(ctx, r.logger, formatRawOrDockerCommand(command, r.dir, r.options))
}

type firecrackerRunner struct {
	name    string
	dir     string
	logger  *Logger
	options Options
}

var _ Runner = &firecrackerRunner{}

func (r *firecrackerRunner) Setup(ctx context.Context, images []string) error {
	return setupFirecracker(ctx, defaultRunner, r.logger, r.name, r.dir, images, r.options)
}

func (r *firecrackerRunner) Teardown(ctx context.Context) error {
	return teardownFirecracker(ctx, defaultRunner, r.logger, r.name)
}

func (r *firecrackerRunner) Run(ctx context.Context, cx CommandSpec) error {
	return runCommand(ctx, r.logger, formatFirecrackerCommand(cx, r.name, r.dir, r.options))
}

type runnerWrapper struct{}

var defaultRunner = &runnerWrapper{}

func (runnerWrapper) RunCommand(ctx context.Context, logger *Logger, command command) error {
	return runCommand(ctx, logger, command)
}
