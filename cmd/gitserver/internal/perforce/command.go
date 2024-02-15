package perforce

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/wrexec"
)

// p4Options holds configuration options for executing Perforce commands.
type p4Options struct {
	arguments []string // arguments to pass to the command

	overrideEnvironment map[string]string // these environment variables will override any existing environment variables with the same name
	environment         []string          // these environment variables will be set for the command

	stdin  io.Reader // alternative stdin for the command
	stdout io.Writer // alternative stdout for the command
	stderr io.Writer // alternative stderr for the command
}

type P4OptionFunc func(*p4Options)

// WithArguments sets the given arguments to the command arguments.
func WithArguments(args ...string) P4OptionFunc {
	return func(o *p4Options) {
		o.arguments = args
	}
}

// WithHost specifies the host to use for the Perforce command.
//
// Example: WithHost("ssl:perforce.example.com:1666")
func WithHost(p4port string) P4OptionFunc {
	return func(o *p4Options) {
		if o.overrideEnvironment == nil {
			o.overrideEnvironment = make(map[string]string)
		}

		o.overrideEnvironment["P4PORT"] = p4port
	}
}

// WithAuthentication specifies the user and password to use for the Perforce command.
//
// Example: WithAuthentication("alice", "hunter2")
func WithAuthentication(user, password string) P4OptionFunc {
	return func(o *p4Options) {
		if o.overrideEnvironment == nil {
			o.overrideEnvironment = make(map[string]string)
		}

		o.overrideEnvironment["P4USER"] = user
		o.overrideEnvironment["P4PASSWD"] = password
	}
}

// WithEnvironment specifies the environment variables to set for the Perforce command.
// Each string should be in the form "key=value".
//
// If multiple environment variables with the same key are provided, only the last one will be used.
//
// If no environment variables are provided, the command will inherit the current process's environment.
//
// Example: WithEnvironment("CONFIG_DIR=/etc/perforce", "P4CONFIG=.p4config")
func WithEnvironment(env ...string) P4OptionFunc {
	return func(o *p4Options) {
		o.environment = append(o.environment, env...)
	}
}

// WithClient specifies the client to use for the Perforce command.
//
// Example: WithClient("alice-client")
func WithClient(client string) P4OptionFunc {
	return func(o *p4Options) {
		if o.overrideEnvironment == nil {
			o.overrideEnvironment = make(map[string]string)
		}

		o.overrideEnvironment["P4CLIENT"] = client
	}
}

// WithStderr specifies the writer to use for the command's stderr output.
func WithStderr(stderr io.Writer) P4OptionFunc {
	return func(o *p4Options) {
		o.stderr = stderr
	}
}

// WithStdin specifies the reader to use for the command's stdin input.
func WithStdin(stdin io.Reader) P4OptionFunc {
	return func(o *p4Options) {
		o.stdin = stdin
	}
}

// WithStdout specifies the writer to use for the command's stdout output.
func WithStdout(stdout io.Writer) P4OptionFunc {
	return func(o *p4Options) {
		o.stdout = stdout
	}
}

func NewBaseCommand(ctx context.Context, homeDir, cwd string, options ...P4OptionFunc) wrexec.Cmder {
	opts := p4Options{}

	// Apply options
	for _, option := range options {
		option(&opts)
	}

	c := exec.CommandContext(ctx, "p4", opts.arguments...)

	c.Env = os.Environ()

	// Apply environment variables if specified
	if opts.environment != nil {
		c.Env = opts.environment
	}

	// Make sure that the override environment variables
	// take precedence over any duplicates in existing environment variables.
	for k, v := range opts.overrideEnvironment {
		c.Env = append(c.Env, fmt.Sprintf("%s=%s", k, v))
	}

	c.Env = append(c.Env,
		fmt.Sprintf("P4CLIENTPATH=%s", cwd),
		fmt.Sprintf("HOME=%s", homeDir),
	)

	// Apply alternate stdin, stdout, and stderr if specified

	if opts.stdin != nil {
		c.Stdin = opts.stdin
	}

	if opts.stdout != nil {
		c.Stdout = opts.stdout
	}

	if opts.stderr != nil {
		c.Stderr = opts.stderr
	}

	// Set the working directory
	c.Dir = cwd

	return wrexec.Wrap(ctx, log.NoOp(), c)
}
