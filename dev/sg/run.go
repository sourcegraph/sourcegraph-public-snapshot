package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/rjeczalik/notify"

	// TODO - deduplicate me
	"github.com/sourcegraph/batch-change-utils/output"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
)

func run(ctx context.Context, cmds ...Command) error {
	chs := make([]<-chan struct{}, 0, len(cmds))
	monitor := &changeMonitor{}
	for _, cmd := range cmds {
		chs = append(chs, monitor.register(cmd))
	}

	pathChanges, err := watch()
	if err != nil {
		return err
	}
	go monitor.run(pathChanges)

	root, err := root.RepositoryRoot()
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	failures := make(chan failedRun, len(cmds))
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for i, cmd := range cmds {
		wg.Add(1)

		go func(cmd Command, ch <-chan struct{}) {
			defer wg.Done()

			if err := runWatch(ctx, cmd, root, ch); err != nil {
				if err != ctx.Err() {
					failures <- failedRun{cmdName: cmd.Name, err: err}
					cancel()
				}
			}
		}(cmd, chs[i])
	}

	wg.Wait()

	failure := <-failures
	printCmdError(out, failure.cmdName, failure.err)
	return failure
}

// failedRun is returned by run when a command failed to run and run exits
type failedRun struct {
	cmdName string
	err     error
}

func (e failedRun) Error() string {
	return fmt.Sprintf("failed to run %s", e.cmdName)
}

// installErr is returned by runWatch if the cmd.Install step fails.
type installErr struct {
	cmdName string
	output  string
}

func (e installErr) Error() string {
	return fmt.Sprintf("install of %s failed: %s", e.cmdName, e.output)
}

// reinstallErr is used internally by runWatch to print a message when a
// command failed to reinstall.
type reinstallErr struct {
	cmdName string
	output  string
}

func (e reinstallErr) Error() string {
	return fmt.Sprintf("reinstalling %s failed: %s", e.cmdName, e.output)
}

func printCmdError(out *output.Output, cmdName string, err error) {
	var message, cmdOut string

	switch e := errors.Cause(err).(type) {
	case installErr:
		message = "Failed to build " + cmdName
		cmdOut = e.output
	case reinstallErr:
		message = "Failed to rebuild " + cmdName
		cmdOut = e.output
	default:
		message = fmt.Sprintf("Failed to run %s: %s", cmdName, err)
	}

	separator := strings.Repeat("-", 80)
	if cmdOut != "" {
		line := output.Linef(
			"", output.StyleWarning,
			"%s\n%s%s:\n%s%s%s%s%s",
			separator, output.StyleBold, message, output.StyleReset,
			cmdOut, output.StyleWarning, separator, output.StyleReset,
		)
		out.WriteLine(line)
	} else {
		line := output.Linef(
			"", output.StyleWarning,
			"%s\n%s%s\n%s%s",
			separator, output.StyleBold, message,
			separator, output.StyleReset,
		)
		out.WriteLine(line)
	}
}

func runWatch(ctx context.Context, cmd Command, root string, reload <-chan struct{}) error {
	startedOnce := false

	for {
		// Build it
		if cmd.Install != "" {
			out.WriteLine(output.Linef("", output.StylePending, "Installing %s...", cmd.Name))

			c := exec.CommandContext(ctx, "bash", "-c", cmd.Install)
			c.Dir = root
			c.Env = makeEnv(conf.Env, cmd.Env)
			cmdOut, err := c.CombinedOutput()
			if err != nil {
				if !startedOnce {
					return installErr{cmdName: cmd.Name, output: string(cmdOut)}
				} else {
					printCmdError(out, cmd.Name, reinstallErr{cmdName: cmd.Name, output: string(cmdOut)})
					// Now we wait for a reload signal before we start to build it again
					select {
					case <-reload:
						continue
					}
				}
			}

			// clear this signal before starting
			select {
			case <-reload:
			default:
			}

			out.WriteLine(output.Linef("", output.StyleSuccess, "%sSuccessfully installed %s%s", output.StyleBold, cmd.Name, output.StyleReset))
		}

		// Run it
		out.WriteLine(output.Linef("", output.StylePending, "Running %s...", cmd.Name))

		commandCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		c := exec.CommandContext(commandCtx, "bash", "-c", cmd.Cmd)
		c.Dir = root
		c.Env = makeEnv(conf.Env, cmd.Env)

		logger := newCmdLogger(cmd.Name, out)
		c.Stdout = logger
		c.Stderr = logger

		if err := c.Start(); err != nil {
			return err
		}

		errs := make(chan error, 1)
		go func() {
			defer close(errs)

			errs <- (func() error {
				if err := c.Wait(); err != nil {
					if exitErr, ok := err.(*exec.ExitError); ok {
						return fmt.Errorf("exited with %d", exitErr.ExitCode())
					}

					return err
				}

				return nil
			})()
		}()

		// TODO: We should probably only set this after N seconds (or when
		// we're sure that the command has booted up -- maybe healthchecks?)
		startedOnce = true
	outer:
		for {
			select {
			case <-reload:
				out.WriteLine(output.Linef("", output.StylePending, "Change detected. Reloading %s...", cmd.Name))

				cancel()    // Stop command
				<-errs      // Wait for exit
				break outer // Reinstall

			case err := <-errs:
				// Exited on its own or errored
				return err
			}
		}
	}
}

func makeEnv(envs ...map[string]string) []string {
	combined := os.Environ()

	expandedEnv := map[string]string{}

	for _, env := range envs {
		for k, v := range env {
			// Expand env vars and keep track of previously set env vars
			// so they can be used when expanding too.
			// TODO: using range to iterate over the env is not stable and thus
			// this won't work
			expanded := os.Expand(v, func(lookup string) string {
				if e, ok := env[lookup]; ok {
					return e
				}
				return os.Getenv(lookup)
			})
			expandedEnv[k] = expanded
			combined = append(combined, fmt.Sprintf("%s=%s", k, expanded))
		}
	}

	return combined
}

//
//

type changeMonitor struct {
	subscriptions []subscription
}

type subscription struct {
	cmd Command
	ch  chan struct{}
}

func (m *changeMonitor) run(paths <-chan string) {
	for path := range paths {
		for _, sub := range m.subscriptions {
			m.notify(sub, path)
		}
	}
}

func (m *changeMonitor) notify(sub subscription, path string) {
	found := false
	for _, prefix := range sub.cmd.Watch {
		if strings.HasPrefix(path, prefix) {
			found = true
		}
	}
	if !found {
		return
	}

	select {
	case sub.ch <- struct{}{}:
	default:
	}
}

func (m *changeMonitor) register(cmd Command) <-chan struct{} {
	ch := make(chan struct{}, 0)
	m.subscriptions = append(m.subscriptions, subscription{cmd, ch})
	return ch
}

//
//

var watchIgnorePatterns = []*regexp.Regexp{
	regexp.MustCompile(`_test\.go$`),
	regexp.MustCompile(`^.bin/`),
	regexp.MustCompile(`^.git/`),
	regexp.MustCompile(`^dev/`),
	regexp.MustCompile(`^node_modules/`),
}

func watch() (<-chan string, error) {
	root, err := root.RepositoryRoot()
	if err != nil {
		return nil, err
	}

	paths := make(chan string)
	events := make(chan notify.EventInfo, 1)

	if err := notify.Watch(root+"/...", events, notify.All); err != nil {
		return nil, err
	}

	go func() {
		defer close(events)
		defer notify.Stop(events)

	outer:
		for event := range events {
			path := strings.TrimPrefix(strings.TrimPrefix(event.Path(), root), "/")

			for _, pattern := range watchIgnorePatterns {
				if pattern.MatchString(path) {
					continue outer
				}
			}

			paths <- path
		}
	}()

	return paths, nil
}

func runTest(ctx context.Context, cmd Command) error {
	root, err := root.RepositoryRoot()
	if err != nil {
		return err
	}

	out.WriteLine(output.Linef("", output.StylePending, "Starting testsuite %s", cmd.Name))
	out.WriteLine(output.Linef("", output.StylePending, "Running %q in %q...", cmd.Cmd, root))
	commandCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	c := exec.CommandContext(commandCtx, "bash", "-c", cmd.Cmd)
	c.Dir = root
	c.Env = makeEnv(conf.Env, cmd.Env)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
