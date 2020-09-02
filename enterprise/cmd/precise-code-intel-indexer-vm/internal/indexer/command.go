package indexer

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	"github.com/inconshreveable/log15"
)

// runCommand invokes the given command on the host machine.
func runCommand(ctx context.Context, command string, args ...string) error {
	cmd, stdout, stderr, err := makeCommand(ctx, command, args...)
	if err != nil {
		return err
	}

	log15.Debug("Running command: %s %s\n", command, strings.Join(args, " "))

	wg := parallel(
		func() { processStream("stdout", stdout) },
		func() { processStream("stderr", stderr) },
	)

	if err := cmd.Start(); err != nil {
		return err
	}

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

// makeCommand returns a new exec.Cmd and pipes to its stdout/stderr streams.
func makeCommand(ctx context.Context, command string, args ...string) (_ *exec.Cmd, stdout, stderr io.Reader, err error) {
	cmd := exec.CommandContext(ctx, command, args...)

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

// parallel runs each function in its own goroutine and returns a wait group that
// blocks until all invocations have returned.
func parallel(funcs ...func()) *sync.WaitGroup {
	var wg sync.WaitGroup

	for _, f := range funcs {
		wg.Add(1)

		go func(f func()) {
			defer wg.Done()
			f()
		}(f)
	}

	return &wg
}

// processStream prefixes and logs each line of the given reader.
func processStream(prefix string, r io.Reader) {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		log15.Info(fmt.Sprintf("%s: %s\n", prefix, scanner.Text()))
	}
}
