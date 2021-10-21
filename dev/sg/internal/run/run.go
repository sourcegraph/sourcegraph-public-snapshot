package run

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/rjeczalik/notify"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func Commands(ctx context.Context, globalEnv map[string]string, verbose bool, cmds ...Command) error {
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
			var err error
			for first := true; cmd.ContinueWatchOnExit || first; first = false {
				if err = runWatch(ctx, cmd, root, globalEnv, ch, verbose); err != nil {
					if errors.Is(err, ctx.Err()) { // if error caused by context, terminate
						return
					}
					if cmd.ContinueWatchOnExit {
						printCmdError(stdout.Out, cmd.Name, err)
						time.Sleep(time.Second * 10) // backoff
					} else {
						failures <- failedRun{cmdName: cmd.Name, err: err}
					}
				}
			}
			if err != nil {
				cancel()
			}
		}(cmd, chs[i])
	}

	wg.Wait()

	select {
	case failure := <-failures:
		printCmdError(stdout.Out, failure.cmdName, failure.err)
		return failure
	default:
		return nil
	}
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

// runErr is used internally by runWatch to print a message when a
// command failed to reinstall.
type runErr struct {
	cmdName  string
	exitCode int
	stderr   string
	stdout   string
}

func (e runErr) Error() string {
	return fmt.Sprintf("failed to run %s.\nstderr:\n%s\nstdout:\n%s\n", e.cmdName, e.stderr, e.stdout)
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
	case runErr:
		message = "Failed to run " + cmdName
		cmdOut = fmt.Sprintf("Exit code: %d\n\n", e.exitCode)

		if len(strings.TrimSpace(e.stdout)) > 0 {
			formattedStdout := "\t" + strings.Join(strings.Split(e.stdout, "\n"), "\n\t")
			cmdOut += fmt.Sprintf("Standard out:\n%s\n", formattedStdout)
		}

		if len(strings.TrimSpace(e.stderr)) > 0 {
			formattedStderr := "\t" + strings.Join(strings.Split(e.stderr, "\n"), "\n\t")
			cmdOut += fmt.Sprintf("Standard err:\n%s\n", formattedStderr)
		}

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

func runWatch(ctx context.Context, cmd Command, root string, globalEnv map[string]string, reload <-chan struct{}, verbose bool) error {
	printDebug := func(f string, args ...interface{}) {
		if !verbose {
			return
		}
		message := fmt.Sprintf(f, args...)
		stdout.Out.WriteLine(output.Linef("", output.StylePending, "%s[DEBUG] %s: %s %s", output.StyleBold, cmd.Name, output.StyleReset, message))
	}

	startedOnce := false

	var (
		md5hash    string
		md5changed bool
	)

	var wg sync.WaitGroup
	var cancelFuncs []context.CancelFunc

	errs := make(chan error, 1)
	defer func() {
		wg.Wait()
		close(errs)
	}()

	for {
		// Build it
		if cmd.Install != "" {
			stdout.Out.WriteLine(output.Linef("", output.StylePending, "Installing %s...", cmd.Name))

			cmdOut, err := BashInRoot(ctx, cmd.Install, makeEnv(globalEnv, cmd.Env))
			if err != nil {
				if !startedOnce {
					return installErr{cmdName: cmd.Name, output: cmdOut}
				} else {
					printCmdError(stdout.Out, cmd.Name, reinstallErr{cmdName: cmd.Name, output: cmdOut})
					// Now we wait for a reload signal before we start to build it again
					<-reload
					continue
				}
			}

			// clear this signal before starting
			select {
			case <-reload:
			default:
			}

			stdout.Out.WriteLine(output.Linef("", output.StyleSuccess, "%sSuccessfully installed %s%s", output.StyleBold, cmd.Name, output.StyleReset))

			if cmd.CheckBinary != "" {
				newHash, err := md5HashFile(filepath.Join(root, cmd.CheckBinary))
				if err != nil {
					return installErr{cmdName: cmd.Name, output: cmdOut}
				}

				md5changed = md5hash != newHash
				md5hash = newHash
			}
		}

		if cmd.CheckBinary == "" || md5changed {
			for _, cancel := range cancelFuncs {
				printDebug("Canceling previous process and waiting for it to exit...")
				cancel() // Stop command
				<-errs   // Wait for exit
				printDebug("Previous command exited")
			}
			cancelFuncs = nil

			// Run it
			stdout.Out.WriteLine(output.Linef("", output.StylePending, "Running %s...", cmd.Name))

			sc, err := startCmd(ctx, root, cmd, globalEnv)
			defer sc.cancel()

			if err != nil {
				return err
			}

			cancelFuncs = append(cancelFuncs, sc.cancel)

			wg.Add(1)
			go func() {
				defer wg.Done()

				err := sc.Wait()

				var e *exec.ExitError
				if errors.As(err, &e) {
					err = runErr{
						cmdName:  cmd.Name,
						exitCode: e.ExitCode(),
						stderr:   sc.CapturedStderr(),
						stdout:   sc.CapturedStdout(),
					}
				}
				if err == nil && cmd.ContinueWatchOnExit {
					stdout.Out.WriteLine(output.Linef("", output.StyleSuccess, "Command %s completed", cmd.Name))
					<-reload // on success, wait for next reload before restarting
					errs <- nil
				} else {
					errs <- err
				}
			}()

			// TODO: We should probably only set this after N seconds (or when
			// we're sure that the command has booted up -- maybe healthchecks?)
			startedOnce = true
		} else {
			stdout.Out.WriteLine(output.Linef("", output.StylePending, "Binary did not change. Not restarting."))
		}

		select {
		case <-reload:
			stdout.Out.WriteLine(output.Linef("", output.StylePending, "Change detected. Reloading %s...", cmd.Name))
			continue // Reinstall

		case err := <-errs:
			// Exited on its own or errored
			if err == nil {
				stdout.Out.WriteLine(output.Linef("", output.StyleSuccess, "%s%s exited without error%s", output.StyleBold, cmd.Name, output.StyleReset))
			}
			return err
		}
	}
}

func makeEnv(envs ...map[string]string) []string {
	combined := os.Environ()

	expandedEnv := map[string]string{}

	for _, env := range envs {
		for k, v := range env {
			if _, ok := os.LookupEnv(k); ok {
				// If the key is already set in the process env, we don't
				// overwrite it. That way we can do something like:
				//
				//	SRC_LOG_LEVEL=debug sg run enterprise-frontend
				//
				// to overwrite the default value in sg.config.yaml
				continue
			}

			// Expand env vars and keep track of previously set env vars
			// so they can be used when expanding too.
			// TODO: using range to iterate over the env is not stable and thus
			// this won't work
			expanded := os.Expand(v, func(lookup string) string {
				// If we're looking up the key that we're trying to define, we
				// skip the self-reference and look in the OS
				if lookup == k {
					return os.Getenv(lookup)
				}

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

func md5HashFile(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return string(h.Sum(nil)), nil
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
	ch := make(chan struct{})
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

func Test(ctx context.Context, cmd Command, args []string, globalEnv map[string]string) error {
	root, err := root.RepositoryRoot()
	if err != nil {
		return err
	}

	stdout.Out.WriteLine(output.Linef("", output.StylePending, "Starting testsuite %q.", cmd.Name))
	if len(args) != 0 {
		stdout.Out.WriteLine(output.Linef("", output.StylePending, "\tAdditional arguments: %s", args))
	}
	commandCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	cmdArgs := []string{cmd.Cmd}
	if len(args) != 0 {
		cmdArgs = append(cmdArgs, args...)
	} else {
		cmdArgs = append(cmdArgs, cmd.DefaultArgs)
	}

	c := exec.CommandContext(commandCtx, "bash", "-c", strings.Join(cmdArgs, " "))
	c.Dir = root
	c.Env = makeEnv(globalEnv, cmd.Env)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	stdout.Out.WriteLine(output.Linef("", output.StylePending, "Running %s in %q...", c, root))

	return c.Run()
}

func Checks(ctx context.Context, globalEnv map[string]string, checks ...Check) (bool, error) {
	success := true

	for _, check := range checks {
		commandCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		c := exec.CommandContext(commandCtx, "bash", "-c", check.Cmd)
		c.Env = makeEnv(globalEnv)

		p := stdout.Out.Pending(output.Linef(output.EmojiLightbulb, output.StylePending, "Running check %q...", check.Name))

		if cmdOut, err := InRoot(c); err != nil {
			success = false

			p.Complete(output.Linef(output.EmojiFailure, output.StyleWarning, "Check %q failed: %s", check.Name, err))

			stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "%s", check.FailMessage))
			if len(cmdOut) != 0 {
				stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "Check produced the following output:"))
				separator := strings.Repeat("-", 80)
				line := output.Linef(
					"", output.StyleWarning,
					"%s\n%s%s%s%s%s",
					separator, output.StyleReset, cmdOut, output.StyleWarning, separator, output.StyleReset,
				)
				stdout.Out.WriteLine(line)
			}
		} else {
			p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Check %q success!", check.Name))
		}
	}

	return success, nil
}
