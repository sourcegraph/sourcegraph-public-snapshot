package command

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
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
	ctx, endObservation := command.Operation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	log15.Info(fmt.Sprintf("Running command: %s", strings.Join(command.Command, " ")))

	if err := validateCommand(command.Command); err != nil {
		return err
	}

	cmd, stdout, stderr, err := prepCommand(ctx, command)
	if err != nil {
		return err
	}

	startTime := time.Now()
	pipeContents, pipeReaderWaitGroup := readProcessPipes(stdout, stderr)
	exitCode, err := monitorCommand(cmd, pipeReaderWaitGroup)
	duration := time.Since(startTime)

	logger.Log(workerutil.ExecutionLogEntry{
		Key:        command.Key,
		Command:    command.Command,
		StartTime:  startTime,
		ExitCode:   exitCode,
		Out:        pipeContents.String(),
		DurationMs: int(duration / time.Millisecond),
	})

	if err != nil {
		return err
	}
	if exitCode != 0 {
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

func prepCommand(ctx context.Context, command command) (cmd *exec.Cmd, stdout io.Reader, stderr io.Reader, err error) {
	cmd = exec.CommandContext(ctx, command.Command[0], command.Command[1:]...)
	cmd.Dir = command.Dir
	cmd.Env = command.Env

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

func readProcessPipes(stdout, stderr io.Reader) (*bytes.Buffer, *sync.WaitGroup) {
	var m sync.Mutex
	out := &bytes.Buffer{}
	wg := &sync.WaitGroup{}

	readIntoBuf := func(prefix string, r io.Reader) {
		defer wg.Done()

		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			m.Lock()
			fmt.Fprintf(out, "%s: %s\n", prefix, scanner.Text())
			m.Unlock()
		}
	}

	wg.Add(2)
	go readIntoBuf("stdout", stdout)
	go readIntoBuf("stderr", stderr)

	return out, wg
}

func monitorCommand(cmd *exec.Cmd, pipeReaderWaitGroup *sync.WaitGroup) (int, error) {
	if err := cmd.Start(); err != nil {
		return 0, err
	}

	pipeReaderWaitGroup.Wait()

	if err := cmd.Wait(); err != nil {
		return 0, err
	}

	return cmd.ProcessState.ExitCode(), nil
}
