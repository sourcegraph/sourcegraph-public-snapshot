package run

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sourcegraph/conc/pool"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/internal/download"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func Commands(ctx context.Context, parentEnv map[string]string, verbose bool, cmds ...ConfigCommand) (err error) {
	if len(cmds) == 0 {
		// no Bazel commands so we return
		return nil
	}
	std.Out.WriteLine(output.Styled(output.StylePending, fmt.Sprintf("Starting %d cmds", len(cmds))))

	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return err
	}

	runner := cmdRunner{
		std.Out,
		cmds,
		repoRoot,
		parentEnv,
		verbose,
	}

	return runner.run(ctx)
}

func (runner *cmdRunner) run(ctx context.Context) error {
	p := pool.New().WithContext(ctx).WithCancelOnError()
	// Start each Bazel command concurrently
	for _, cmd := range runner.cmds {
		cmd := cmd
		p.Go(func(ctx context.Context) error {
			std.Out.WriteLine(output.Styledf(output.StylePending, "Running %s...", cmd.GetName()))

			// Start watching the commands dependencies
			wantRestart, err := cmd.StartWatch(ctx)
			if err != nil {
				runner.Write("Failed to watch " + cmd.GetName())
				runner.printError(cmd, err)
				return err
			}

			// start up the binary
			sc, err := runner.start(ctx, cmd)
			if err != nil {
				runner.Write("Failed to start " + cmd.GetName())
				runner.printError(cmd, err)
				return errors.Wrapf(err, "failed to start command %q", cmd.GetName())
			}
			defer sc.cancel()

			// Wait forever until we're asked to stop or that restarting returns an error.
			for {

				select {
				// Handle context cancelled
				case <-ctx.Done():
					runner.debug("context error" + cmd.GetName())
					return ctx.Err()

				// Handle process exit
				case err := <-sc.ErrorChannel():
					// Exited on its own or errored
					if err != nil {
						runner.debug("Error channel " + cmd.GetName())
						return err
					}
					runner.WriteLine(output.Styledf(output.StyleSuccess, "%s%s exited without error%s", output.StyleBold, cmd.GetName(), output.StyleReset))

					// If we shouldn't restart when the process exits, return
					if !cmd.GetContinueWatchOnExit() {
						return nil
					}

				// handle file watcher triggered
				case <-wantRestart:
					// If the command has an installer, re-run the install and determine if we should restart
					runner.WriteLine(output.Styledf(output.StylePending, "Change detected. Reloading %s...", cmd.GetName()))
					shouldRestart, err := runner.reinstall(ctx, cmd)
					if err != nil {
						runner.debug("reinstall failure: %s", cmd.GetName())
						runner.printError(cmd, err)
						return err
					}

					if shouldRestart {
						std.Out.WriteLine(output.Styledf(output.StylePending, "Restarting %s...", cmd.GetName()))
						sc.cancel()
						sc, err = runner.start(ctx, cmd)
						defer sc.cancel()
						if err != nil {
							runner.debug("restart failure " + cmd.GetName())
							return err
						}
					} else {
						std.Out.WriteLine(output.Styledf(output.StylePending, "Binary for %s did not change. Not restarting.", cmd.GetName()))
					}
				}
			}
		})
	}

	err := p.Wait()
	runner.Write("Completed all commands")
	return err
}

type cmdRunner struct {
	*std.Output
	cmds           []ConfigCommand
	repositoryRoot string
	parentEnv      map[string]string
	verbose        bool
}

func (runner *cmdRunner) printError(cmd ConfigCommand, err error) {
	printCmdError(runner.Output.Output, cmd.GetName(), err)
}

func (runner *cmdRunner) debug(msg string, args ...any) {
	if runner.verbose {
		message := fmt.Sprintf(msg, args...)
		runner.WriteLine(output.Styledf(output.StylePending, "%s[DEBUG]: %s %s", output.StyleBold, output.StyleReset, message))
	}
}

func (runner *cmdRunner) start(ctx context.Context, cmd ConfigCommand) (*startedCmd, error) {
	return startConfigCmd(ctx, cmd, runner.repositoryRoot, runner.parentEnv)
}

func (runner *cmdRunner) reinstall(ctx context.Context, cmd ConfigCommand) (bool, error) {
	if installer, ok := cmd.(Installer); ok {
		bin, err := cmd.GetBinaryLocation()
		if err != nil {
			noBinary := noBinaryError{}
			// If the command doesn't have a CheckBinary, we just ignore it
			if errors.As(err, &noBinary) {
				return false, nil
			} else {
				return false, err
			}
		}

		oldHash, err := md5HashFile(bin)
		if err != nil {
			return false, err
		}

		if err := installer.RunInstall(ctx, runner.parentEnv); err != nil {
			printCmdError(std.Out.Output, cmd.GetName(), err)
			return false, err
		}
		newHash, err := md5HashFile(bin)
		if err != nil {
			return false, err
		}

		return oldHash != newHash, nil
	}

	// If there is no installer, then we always restart
	return true, nil
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

func Test(ctx context.Context, cmd ConfigCommand, parentEnv map[string]string) error {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return err
	}

	std.Out.WriteLine(output.Styledf(output.StylePending, "Starting testsuite %q.", cmd.GetName()))
	sc, err := startConfigCmd(ctx, cmd, repoRoot, parentEnv)
	if err != nil {
		printCmdError(std.Out.Output, cmd.GetName(), err)
	}
	return sc.Wait()
}
