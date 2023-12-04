package run

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/grafana/regexp"
	"github.com/rjeczalik/notify"
	"golang.org/x/sync/semaphore"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/analytics"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/interrupt"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/internal/download"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

const MAX_CONCURRENT_BUILD_PROCS = 4

func Commands(ctx context.Context, parentEnv map[string]string, verbose bool, cmds ...Command) error {
	if len(cmds) == 0 {
		// Exit early if there are no commands to run.
		return nil
	}

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

	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return err
	}

	// binaries get installed to <repository-root>/.bin. If the binary is installed with go build, then go
	// will create .bin directory. Some binaries (like docsite) get downloaded instead of built and therefore
	// need the directory to exist before hand.
	binDir := filepath.Join(repoRoot, ".bin")
	if err := os.Mkdir(binDir, 0755); err != nil && !os.IsExist(err) {
		return err
	}

	wg := sync.WaitGroup{}
	installSemaphore := semaphore.NewWeighted(MAX_CONCURRENT_BUILD_PROCS)
	failures := make(chan failedRun, len(cmds))
	installed := make(chan string, len(cmds))
	okayToStart := make(chan struct{})

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	runner := &cmdRunner{
		verbose:          verbose,
		installSemaphore: installSemaphore,
		failures:         failures,
		installed:        installed,
		okayToStart:      okayToStart,
		repositoryRoot:   repoRoot,
		parentEnv:        parentEnv,
	}

	cmdNames := make(map[string]struct{}, len(cmds))

	for i, cmd := range cmds {
		cmdNames[cmd.Name] = struct{}{}

		wg.Add(1)

		go func(cmd Command, ch <-chan struct{}) {
			defer wg.Done()
			var err error
			for first := true; cmd.ContinueWatchOnExit || first; first = false {
				if err = runner.runAndWatch(ctx, cmd, ch); err != nil {
					if errors.Is(err, ctx.Err()) { // if error caused by context, terminate
						return
					}
					if cmd.ContinueWatchOnExit {
						printCmdError(std.Out.Output, cmd.Name, err)
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

	err = runner.waitForInstallation(ctx, cmdNames)
	if err != nil {
		return err
	}

	if err := writePid(); err != nil {
		return err
	}

	wg.Wait()

	select {
	case <-ctx.Done():
		printCmdError(std.Out.Output, "other", ctx.Err())
		return ctx.Err()
	case failure := <-failures:
		printCmdError(std.Out.Output, failure.cmdName, failure.err)
		return failure
	default:
		return nil
	}
}

type cmdRunner struct {
	verbose bool

	installSemaphore *semaphore.Weighted
	failures         chan failedRun
	installed        chan string
	okayToStart      chan struct{}

	repositoryRoot string
	parentEnv      map[string]string
}

func (c *cmdRunner) runAndWatch(ctx context.Context, cmd Command, reload <-chan struct{}) error {
	printDebug := func(f string, args ...any) {
		if !c.verbose {
			return
		}
		message := fmt.Sprintf(f, args...)
		std.Out.WriteLine(output.Styledf(output.StylePending, "%s[DEBUG] %s: %s %s", output.StyleBold, cmd.Name, output.StyleReset, message))
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
		if cmd.Install != "" || cmd.InstallFunc != "" {
			install := func() (string, error) {
				if err := c.installSemaphore.Acquire(ctx, 1); err != nil {
					return "", errors.Wrap(err, "lockfiles semaphore")
				}
				defer c.installSemaphore.Release(1)

				if startedOnce {
					std.Out.WriteLine(output.Styledf(output.StylePending, "Installing %s...", cmd.Name))
				}
				if cmd.Install != "" && cmd.InstallFunc == "" {
					return BashInRoot(ctx, cmd.Install, makeEnv(c.parentEnv, cmd.Env))
				} else if cmd.Install == "" && cmd.InstallFunc != "" {
					fn, ok := installFuncs[cmd.InstallFunc]
					if !ok {
						return "", errors.Newf("no install func with name %q found", cmd.InstallFunc)
					}
					return "", fn(ctx, makeEnvMap(c.parentEnv, cmd.Env))
				}

				return "", nil
			}

			cmdOut, err := install()
			if err != nil {
				if !startedOnce {
					return installErr{cmdName: cmd.Name, output: cmdOut, originalErr: err}
				} else {
					printCmdError(std.Out.Output, cmd.Name, reinstallErr{cmdName: cmd.Name, output: cmdOut})
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

			if startedOnce {
				std.Out.WriteLine(output.Styledf(output.StyleSuccess, "%sSuccessfully installed %s%s", output.StyleBold, cmd.Name, output.StyleReset))
			}

			if cmd.CheckBinary != "" {
				newHash, err := md5HashFile(filepath.Join(c.repositoryRoot, cmd.CheckBinary))
				if err != nil {
					return installErr{cmdName: cmd.Name, output: cmdOut, originalErr: err}
				}

				md5changed = md5hash != newHash
				md5hash = newHash
			}

		}

		if !startedOnce {
			c.installed <- cmd.Name
			<-c.okayToStart
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
			std.Out.WriteLine(output.Styledf(output.StylePending, "Running %s...", cmd.Name))

			sc, err := startCmd(ctx, c.repositoryRoot, cmd, c.parentEnv)
			if err != nil {
				return err
			}
			defer sc.cancel()

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
					std.Out.WriteLine(output.Styledf(output.StyleSuccess, "Command %s completed", cmd.Name))
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
			std.Out.WriteLine(output.Styled(output.StylePending, "Binary did not change. Not restarting."))
		}

		select {
		case <-reload:
			std.Out.WriteLine(output.Styledf(output.StylePending, "Change detected. Reloading %s...", cmd.Name))
			continue // Reinstall

		case err := <-errs:
			// Exited on its own or errored
			if err == nil {
				std.Out.WriteLine(output.Styledf(output.StyleSuccess, "%s%s exited without error%s", output.StyleBold, cmd.Name, output.StyleReset))
			}
			return err
		}
	}
}

func (c *cmdRunner) waitForInstallation(ctx context.Context, cmdNames map[string]struct{}) error {
	installationStart := time.Now()
	installationSpans := make(map[string]*analytics.Span, len(cmdNames))
	for name := range cmdNames {
		_, installationSpans[name] = analytics.StartSpan(ctx, fmt.Sprintf("install %s", name), "install_command")
	}
	interrupt.Register(func() {
		for _, span := range installationSpans {
			if span.IsRecording() {
				span.Cancelled()
				span.End()
			}
		}
	})

	std.Out.Write("")
	std.Out.WriteLine(output.Linef(output.EmojiLightbulb, output.StyleBold, "Installing %d commands...", len(cmdNames)))
	std.Out.Write("")

	waitingMessages := []string{
		"Still waiting for %s to finish installing...",
		"Yup, still waiting for %s to finish installing...",
		"Here's the bad news: still waiting for %s to finish installing. The good news is that we finally have a chance to talk, no?",
		"Still waiting for %s to finish installing...",
		"Hey, %s, there's people waiting for you, pal",
		"Sooooo, how are ya? Yeah, waiting. I hear you. Wish %s would hurry up.",
		"I mean, what is %s even doing?",
		"I now expect %s to mean 'producing a miracle' with 'installing'",
		"Still waiting for %s to finish installing...",
		"Before this I think the longest I ever had to wait was at Disneyland in '99, but %s is now #1",
		"Still waiting for %s to finish installing...",
		"At this point it could be anything - does your computer still have power? Come on, %s",
		"Might as well check Slack. %s is taking its time...",
		"In German there's a saying: ein guter Käse braucht seine Zeit - a good cheese needs its time. Maybe %s is cheese?",
		"If %ss turns out to be cheese I'm gonna lose it. Hey, hurry up, will ya",
		"Still waiting for %s to finish installing...",
	}
	messageCount := 0

	const tickInterval = 15 * time.Second
	ticker := time.NewTicker(tickInterval)

	done := 0.0
	total := float64(len(cmdNames))
	progress := std.Out.Progress([]output.ProgressBar{
		{Label: fmt.Sprintf("Installing %d commands", len(cmdNames)), Max: total},
	}, nil)

	for {
		select {
		case cmdName := <-c.installed:
			ticker.Reset(tickInterval)

			delete(cmdNames, cmdName)
			done += 1.0
			installationSpans[cmdName].Succeeded()
			installationSpans[cmdName].End()

			progress.WriteLine(output.Styledf(output.StyleSuccess, "%s installed", cmdName))

			progress.SetValue(0, done)
			progress.SetLabelAndRecalc(0, fmt.Sprintf("%d/%d commands installed", int(done), int(total)))

			// Everything installed!
			if len(cmdNames) == 0 {
				progress.Complete()

				duration := time.Since(installationStart)

				std.Out.Write("")
				if c.verbose {
					std.Out.WriteLine(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Everything installed! Took %s. Booting up the system!", duration))
				} else {
					std.Out.WriteLine(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Everything installed! Booting up the system!"))
				}
				std.Out.Write("")

				close(c.okayToStart)
				return nil
			}

		case failure := <-c.failures:
			progress.Destroy()
			installationSpans[failure.cmdName].RecordError("failed", failure.err)
			installationSpans[failure.cmdName].End()

			// Something went wrong with an installation, no need to wait for the others
			printCmdError(std.Out.Output, failure.cmdName, failure.err)
			return failure

		case <-ticker.C:
			names := []string{}
			for name := range cmdNames {
				names = append(names, name)
			}

			idx := messageCount
			if idx > len(waitingMessages)-1 {
				idx = len(waitingMessages) - 1
			}
			msg := waitingMessages[idx]

			emoji := output.EmojiHourglass
			if idx > 3 {
				emoji = output.EmojiShrug
			}

			progress.WriteLine(output.Linef(emoji, output.StyleBold, msg, strings.Join(names, ", ")))
			messageCount += 1
		}
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

	originalErr error
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
		if e.originalErr != nil {
			if errWithout, ok := e.originalErr.(errorWithoutOutputer); ok {
				// If we can, let's strip away the output, otherwise this gets
				// too noisy.
				message += ": " + errWithout.ErrorWithoutOutput()
			} else {
				message += ": " + e.originalErr.Error()
			}
		}
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

type installFunc func(context.Context, map[string]string) error

var installFuncs = map[string]installFunc{
	"installCaddy": func(ctx context.Context, env map[string]string) error {
		version := env["CADDY_VERSION"]
		if version == "" {
			return errors.New("could not find CADDY_VERSION in env")
		}

		root, err := root.RepositoryRoot()
		if err != nil {
			return err
		}

		var os string
		switch runtime.GOOS {
		case "linux":
			os = "linux"
		case "darwin":
			os = "mac"
		}

		archiveName := fmt.Sprintf("caddy_%s_%s_%s", version, os, runtime.GOARCH)
		url := fmt.Sprintf("https://github.com/caddyserver/caddy/releases/download/v%s/%s.tar.gz", version, archiveName)

		target := filepath.Join(root, fmt.Sprintf(".bin/caddy_%s", version))

		return download.ArchivedExecutable(ctx, url, target, "caddy")
	},
	"installJaeger": func(ctx context.Context, env map[string]string) error {
		version := env["JAEGER_VERSION"]

		// Make sure the data folder exists.
		disk := env["JAEGER_DISK"]
		if err := os.MkdirAll(disk, 0755); err != nil {
			return err
		}

		if version == "" {
			return errors.New("could not find JAEGER_VERSION in env")
		}

		root, err := root.RepositoryRoot()
		if err != nil {
			return err
		}

		archiveName := fmt.Sprintf("jaeger-%s-%s-%s", version, runtime.GOOS, runtime.GOARCH)
		url := fmt.Sprintf("https://github.com/jaegertracing/jaeger/releases/download/v%s/%s.tar.gz", version, archiveName)

		target := filepath.Join(root, fmt.Sprintf(".bin/jaeger-all-in-one-%s", version))

		return download.ArchivedExecutable(ctx, url, target, fmt.Sprintf("%s/jaeger-all-in-one", archiveName))
	},
}

// makeEnv merges environments starting from the left, meaning the first environment will be overriden by the second one, skipping
// any key that has been explicitly defined in the current environment of this process. This enables users to manually overrides
// environment variables explictly, i.e FOO=1 sg start will have FOO=1 set even if a command or commandset sets FOO.
func makeEnv(envs ...map[string]string) (combined []string) {
	for k, v := range makeEnvMap(envs...) {
		combined = append(combined, fmt.Sprintf("%s=%s", k, v))
	}
	return combined
}

func makeEnvMap(envs ...map[string]string) map[string]string {
	combined := map[string]string{}
	for _, pair := range os.Environ() {
		elems := strings.SplitN(pair, "=", 2)
		if len(elems) != 2 {
			panic("space/time continuum wrong")
		}

		combined[elems[0]] = elems[1]
	}

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
			combined[k] = expanded
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
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return nil, err
	}

	paths := make(chan string)
	events := make(chan notify.EventInfo, 1)

	if err := notify.Watch(repoRoot+"/...", events, notify.All); err != nil {
		return nil, err
	}

	go func() {
		defer close(events)
		defer notify.Stop(events)

	outer:
		for event := range events {
			path := strings.TrimPrefix(strings.TrimPrefix(event.Path(), repoRoot), "/")

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

func Test(ctx context.Context, cmd Command, args []string, parentEnv map[string]string) error {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return err
	}

	std.Out.WriteLine(output.Styledf(output.StylePending, "Starting testsuite %q.", cmd.Name))
	if len(args) != 0 {
		std.Out.WriteLine(output.Styledf(output.StylePending, "\tAdditional arguments: %s", args))
	}
	commandCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	cmdArgs := []string{cmd.Cmd}
	if len(args) != 0 {
		cmdArgs = append(cmdArgs, args...)
	} else {
		cmdArgs = append(cmdArgs, cmd.DefaultArgs)
	}

	secretsEnv, err := getSecrets(ctx, cmd.Name, cmd.ExternalSecrets)
	if err != nil {
		std.Out.WriteLine(output.Styledf(output.StyleWarning, "[%s] %s %s",
			cmd.Name, output.EmojiFailure, err.Error()))
	}

	if cmd.Preamble != "" {
		std.Out.WriteLine(output.Styledf(output.StyleOrange, "[%s] %s %s", cmd.Name, output.EmojiInfo, cmd.Preamble))
	}

	c := exec.CommandContext(commandCtx, "bash", "-c", strings.Join(cmdArgs, " "))
	c.Dir = repoRoot
	c.Env = makeEnv(parentEnv, secretsEnv, cmd.Env)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	std.Out.WriteLine(output.Styledf(output.StylePending, "Running %s in %q...", c, repoRoot))

	return c.Run()
}
