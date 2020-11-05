package command

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/inconshreveable/log15"
)

type command struct {
	Commands []string
	Dir      string
	Env      []string
}

var allowedCommands = []string{
	"docker",
	"git",
	"ignite",
	"src",
}

var ErrIllegalCommand = errors.New("illegal command")

// runCommand invokes the given command on the host machine. The standard output and
// standard error streams of the invoked command are written to the given logger.
func runCommand(ctx context.Context, logger *Logger, command command) error {
	if len(command.Commands) == 0 {
		return ErrIllegalCommand
	}

	found := false
	for _, candidate := range allowedCommands {
		if command.Commands[0] == candidate {
			found = true
		}
	}
	if !found {
		return ErrIllegalCommand
	}

	cmd := exec.CommandContext(ctx, command.Commands[0], command.Commands[1:]...)
	cmd.Dir = command.Dir
	cmd.Env = command.Env

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	log15.Info(fmt.Sprintf("Running command: %s", strings.Join(command.Commands, " ")))

	wg := wgWrap(func() {
		logger.RecordCommand(command.Commands, stdout, stderr)
	})

	if err := cmd.Start(); err != nil {
		return err
	}

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func wgWrap(f func()) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		f()
	}()

	return &wg
}
