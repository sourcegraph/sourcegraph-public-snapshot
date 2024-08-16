package command

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/sourcegraph/log"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/util"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/cmdlogger"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var allowedBinaries = []string{
	"docker",
	"git",
	"ignite",
	"src",
}

func init() {
	// We run /bin/sh to execute scripts locally in the shell runtime, so we need
	// to allow that, too.
	if util.HasShellBuildTag() {
		allowedBinaries = append(allowedBinaries, "/bin/sh")
	}
}

type Command interface {
	Run(ctx context.Context, cmdLogger cmdlogger.Logger, spec Spec) error
}

type RealCommand struct {
	CmdRunner util.CmdRunner
	Logger    log.Logger
}

var _ Command = &RealCommand{}

type Spec struct {
	Key       string
	Name      string
	Command   []string
	Dir       string
	Env       []string
	Image     string
	Operation *observation.Operation
}

func (c *RealCommand) Run(ctx context.Context, cmdLogger cmdlogger.Logger, spec Spec) (err error) {
	// The context here is used below as a guard against the command finishing before we close
	// the stdout and stderr pipes. This context may not cancel until after logs for the job
	// have been flushed, or after the 30m job deadline, so we enforce a cancellation of a
	// child context at function exit to clean the goroutine up eagerly.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ctx, _, endObservation := spec.Operation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	c.Logger.Info(
		"Running command",
		log.String("key", spec.Key),
		log.String("workingDir", spec.Dir),
		log.String("image", spec.Image),
	)

	// Check if we can even run the command.
	if err := validateCommand(spec.Command); err != nil {
		return err
	}

	cmd, stdout, stderr, err := c.prepCommand(ctx, spec)
	if err != nil {
		return err
	}

	context.AfterFunc(ctx, func() {
		// There is a deadlock condition due the following strange decisions:
		//
		// 1. The pipes attached to a command are not closed if the context
		//    attached to the command is canceled. The pipes are only closed
		//    after Wait has been called.
		// 2. According to the docs, we are not meant to call cmd.Wait() until
		//    we have complete read the pipes attached to the command.
		//
		// Since we're following the expected usage, we block on a wait group
		// tracking the consumption of stdout and stderr pipes in two separate
		// goroutines between calls to Start and Wait. This means that if there
		// is a reason the command is abandoned but the pipes are not closed
		// (such as context cancellation), we will hang indefinitely.
		//
		// To be defensive, we'll forcibly close both pipes when the context has
		// finished. These may return an ErrClosed condition, but we don't really
		// care: the command package doesn't surface errors when closing the pipes
		// either.
		stdout.Close()
		stderr.Close()
	})

	// Create the log entry that we will be writing stdout and stderr to.
	logEntry := cmdLogger.LogEntry(spec.Key, spec.Command)
	defer logEntry.Close()

	// Starts writing the stdout and stderr of the command to the log entry.
	pipeReaderWaitGroup := readProcessPipes(logEntry, stdout, stderr)
	// Start the command and wait for it to finish.
	exitCode, err := startCommand(ctx, cmd, pipeReaderWaitGroup)
	// Finalize the log entry with the exit code.
	logEntry.Finalize(exitCode)

	if err != nil {
		return err
	}
	if exitCode != 0 {
		// If is context cancellation, forward the ctx.Err().
		if err = ctx.Err(); err != nil {
			return err
		}

		return errors.Newf("command failed with exit code %d", exitCode)
	}

	return nil
}

func validateCommand(command []string) error {
	if len(command) == 0 {
		return ErrIllegalCommand
	}

	for _, candidate := range allowedBinaries {
		if command[0] == candidate {
			return nil
		}
	}

	return ErrIllegalCommand
}

// ErrIllegalCommand is returned when a command is not allowed to be run.
var ErrIllegalCommand = errors.New("illegal command")

func (c *RealCommand) prepCommand(ctx context.Context, options Spec) (cmd *exec.Cmd, stdout, stderr io.ReadCloser, err error) {
	cmd = c.CmdRunner.CommandContext(ctx, options.Command[0], options.Command[1:]...)
	cmd.Dir = options.Dir

	env := options.Env
	for _, k := range forwardedHostEnvVars {
		env = append(env, fmt.Sprintf("%s=%s", k, os.Getenv(k)))
	}

	cmd.Env = env

	stdout, err = cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, err
	}

	stderr, err = cmd.StderrPipe()
	if err != nil {
		return nil, nil, nil, err
	}

	return cmd, stdout, stderr, nil
}

// forwardedHostEnvVars is a list of environment variable names that are inherited
// when executing a command on the host. These are commonly required by programs
// we invoke, such as calling docker commands.
var forwardedHostEnvVars = []string{"HOME", "PATH", "USER", "DOCKER_HOST"}

func readProcessPipes(w io.WriteCloser, stdout, stderr io.Reader) *errgroup.Group {
	eg := &errgroup.Group{}

	eg.Go(func() error {
		return readIntoBuffer("stdout", w, stdout)
	})
	eg.Go(func() error {
		return readIntoBuffer("stderr", w, stderr)
	})

	return eg
}

func readIntoBuffer(prefix string, w io.WriteCloser, r io.Reader) error {
	scanner := bufio.NewScanner(r)
	// Allocate an initial buffer of 4k.
	buf := make([]byte, 4*1024)
	// And set the maximum size used to buffer a token to 100M.
	// TODO: Tweak this value as needed.
	scanner.Buffer(buf, maxBuffer)
	for scanner.Scan() {
		_, err := fmt.Fprintf(w, "%s: %s\n", prefix, scanner.Text())
		if err != nil {
			return err
		}
	}
	return scanner.Err()
}

const maxBuffer = 100 * 1024 * 1024

// startCommand starts the given command and waits for the given errgroup to complete.
// This function returns a non-nil error only if there was a system issue - commands that
// run but fail due to a non-zero exit code will return a nil error and the exit code.
func startCommand(ctx context.Context, cmd *exec.Cmd, pipeReaderWaitGroup *errgroup.Group) (int, error) {
	if err := cmd.Start(); err != nil {
		return 0, errors.Wrap(err, "starting command")
	}

	select {
	case <-ctx.Done():
	case err := <-watchErrGroup(pipeReaderWaitGroup):
		if err != nil {
			return 0, errors.Wrap(err, "reading process pipes")
		}
	}

	if err := cmd.Wait(); err != nil {
		var e *exec.ExitError
		if errors.As(err, &e) {
			return e.ExitCode(), nil
		}

		return 0, errors.Wrap(err, "waiting for command")
	}

	// All good, command ran successfully.
	return 0, nil
}

func watchErrGroup(eg *errgroup.Group) <-chan error {
	ch := make(chan error)
	go func() {
		ch <- eg.Wait()
		close(ch)
	}()

	return ch
}
