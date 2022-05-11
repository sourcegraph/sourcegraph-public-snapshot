package command

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/inconshreveable/log15"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type command struct {
	Key       string
	Command   []string
	Dir       string
	Env       []string
	Operation *observation.Operation
}

// runCommand invokes the given command on the host machine. The standard output and
// standard error streams of the invoked command are written to the given logger.
func runCommand(ctx context.Context, command command, logger *Logger) (err error) {
	// The context here is used below as a guard against the command finishing before we close
	// the stdout and stderr pipes. This context may not cancel until after logs for the job
	// have been flushed, or after the 30m job deadline, so we enforce a cancellation of a
	// child context at function exit to clean the goroutine up eagerly.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ctx, _, endObservation := command.Operation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	log15.Info(fmt.Sprintf("Running command: %s", strings.Join(command.Command, " ")))

	if err := validateCommand(command.Command); err != nil {
		return err
	}

	cmd, stdout, stderr, err := prepCommand(ctx, command)
	if err != nil {
		return err
	}

	go func() {
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

		<-ctx.Done()
		stdout.Close()
		stderr.Close()
	}()

	handle := logger.Log(command.Key, command.Command)
	defer handle.Close()

	pipeReaderWaitGroup := readProcessPipes(handle, stdout, stderr)
	exitCode, err := monitorCommand(ctx, cmd, pipeReaderWaitGroup)
	handle.Finalize(exitCode)
	if err != nil {
		return err
	}
	if exitCode != 0 {
		// If is context cancelation, forward the ctx.Err().
		if err := ctx.Err(); err != nil {
			return err
		}

		return errors.New("command failed")
	}
	return nil
}

var allowedBinaries = []string{
	"docker",
	"git",
	"ignite",
	"src",
}

var ErrIllegalCommand = errors.New("illegal command")

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

func prepCommand(ctx context.Context, command command) (cmd *exec.Cmd, stdout, stderr io.ReadCloser, err error) {
	cmd = exec.CommandContext(ctx, command.Command[0], command.Command[1:]...)
	cmd.Dir = command.Dir

	env := command.Env
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
// we shell out to, such a docker.
var forwardedHostEnvVars = []string{"HOME", "PATH", "USER"}

func readProcessPipes(logWriter io.WriteCloser, stdout, stderr io.Reader) *errgroup.Group {
	eg := &errgroup.Group{}

	readIntoBuf := func(prefix string, r io.Reader) error {
		scanner := bufio.NewScanner(r)
		// Allocate an initial buffer of 4k.
		buf := make([]byte, 4*1024)
		// And set the maximum size used to buffer a token to 100M.
		// TODO: Tweak this value as needed.
		scanner.Buffer(buf, 100*1024*1024)
		for scanner.Scan() {
			fmt.Fprintf(logWriter, "%s: %s\n", prefix, scanner.Text())
		}
		return scanner.Err()
	}

	eg.Go(func() error {
		return readIntoBuf("stdout", stdout)
	})
	eg.Go(func() error {
		return readIntoBuf("stderr", stderr)
	})

	return eg
}

// monitorCommand starts the given command and waits for the given errgroup to complete.
// This function returns a non-nil error only if there was a system issue - commands that
// run but fail due to a non-zero exit code will return a nil error and the exit code.
func monitorCommand(ctx context.Context, cmd *exec.Cmd, pipeReaderWaitGroup *errgroup.Group) (int, error) {
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

		log15.Error("Non exit-error returned from command", "err", err)
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
