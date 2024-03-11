package run

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/grafana/regexp"
	"github.com/sourcegraph/conc/pool"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/interrupt"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
	"github.com/sourcegraph/sourcegraph/lib/process"
)

type Command struct {
	Name                string
	Cmd                 string            `yaml:"cmd"`
	Install             string            `yaml:"install"`
	InstallFunc         string            `yaml:"install_func"`
	CheckBinary         string            `yaml:"checkBinary"`
	Env                 map[string]string `yaml:"env"`
	Watch               []string          `yaml:"watch"`
	IgnoreStdout        bool              `yaml:"ignoreStdout"`
	IgnoreStderr        bool              `yaml:"ignoreStderr"`
	DefaultArgs         string            `yaml:"defaultArgs"`
	ContinueWatchOnExit bool              `yaml:"continueWatchOnExit"`
	// Preamble is a short and visible message, displayed when the command is launched.
	Preamble string `yaml:"preamble"`

	ExternalSecrets map[string]secrets.ExternalSecret `yaml:"external_secrets"`
	Description     string                            `yaml:"description"`

	// ATTENTION: If you add a new field here, be sure to also handle that
	// field in `Merge` (below).
}

func (cmd Command) GetName() string {
	return cmd.Name
}

func (cmd Command) GetContinueWatchOnExit() bool {
	return cmd.ContinueWatchOnExit
}

func (cmd Command) GetBinaryLocation() (string, error) {
	if cmd.CheckBinary != "" {
		repoRoot, err := root.RepositoryRoot()
		if err != nil {
			return "", err
		}
		return filepath.Join(repoRoot, cmd.CheckBinary), nil
	}
	return "", noBinaryError{name: cmd.Name}
}

func (cmd Command) GetExternalSecrets() map[string]secrets.ExternalSecret {
	return cmd.ExternalSecrets
}

func (cmd Command) GetIgnoreStdout() bool {
	return cmd.IgnoreStdout
}

func (cmd Command) GetIgnoreStderr() bool {
	return cmd.IgnoreStderr
}

func (cmd Command) GetPreamble() string {
	return cmd.Preamble
}

func (cmd Command) GetEnv() map[string]string {
	return cmd.Env
}

func (cmd Command) GetExecCmd(ctx context.Context) (*exec.Cmd, error) {
	return exec.CommandContext(ctx, "bash", "-c", cmd.Cmd), nil
}

func (cmd Command) RunInstall(ctx context.Context, parentEnv map[string]string) error {
	if cmd.requiresInstall() {
		if cmd.hasBashInstaller() {
			return cmd.bashInstall(ctx, parentEnv)
		} else {
			return cmd.functionInstall(ctx, parentEnv)
		}
	}

	return nil
}

// Standard commands ignore installer
func (cmd Command) SetInstallerOutput(chan<- output.FancyLine) {}

func (cmd Command) Count() int {
	return 1
}

func (cmd Command) requiresInstall() bool {
	return cmd.Install != "" || cmd.InstallFunc != ""
}

func (cmd Command) hasBashInstaller() bool {
	return cmd.Install != "" || cmd.InstallFunc == ""
}

func (cmd Command) bashInstall(ctx context.Context, parentEnv map[string]string) error {
	output, err := BashInRoot(ctx, cmd.Install, makeEnv(parentEnv, cmd.Env))
	if err != nil {
		return installErr{cmdName: cmd.Name, output: output, originalErr: err}
	}
	return nil
}

func (cmd Command) functionInstall(ctx context.Context, parentEnv map[string]string) error {
	fn, ok := installFuncs[cmd.InstallFunc]
	if !ok {
		return installErr{cmdName: cmd.Name, originalErr: errors.Newf("no install func with name %q found", cmd.InstallFunc)}
	}
	if err := fn(ctx, makeEnvMap(parentEnv, cmd.Env)); err != nil {
		return installErr{cmdName: cmd.Name, originalErr: err}
	}

	return nil
}

func (cmd Command) getWatchPaths() ([]string, error) {
	root, err := root.RepositoryRoot()
	if err != nil {
		return nil, err
	}

	fullPaths := make([]string, len(cmd.Watch))
	for i, path := range cmd.Watch {
		fullPaths[i] = filepath.Join(root, path)
	}

	return fullPaths, nil
}

func (cmd Command) StartWatch(ctx context.Context) (<-chan struct{}, error) {
	if watchPaths, err := cmd.getWatchPaths(); err != nil {
		return nil, err
	} else {
		return WatchPaths(ctx, watchPaths)
	}
}

func (c Command) Merge(other Command) Command {
	merged := c

	if other.Name != merged.Name && other.Name != "" {
		merged.Name = other.Name
	}
	if other.Cmd != merged.Cmd && other.Cmd != "" {
		merged.Cmd = other.Cmd
	}
	if other.Install != merged.Install && other.Install != "" {
		merged.Install = other.Install
	}
	if other.InstallFunc != merged.InstallFunc && other.InstallFunc != "" {
		merged.InstallFunc = other.InstallFunc
	}
	if other.IgnoreStdout != merged.IgnoreStdout && !merged.IgnoreStdout {
		merged.IgnoreStdout = other.IgnoreStdout
	}
	if other.IgnoreStderr != merged.IgnoreStderr && !merged.IgnoreStderr {
		merged.IgnoreStderr = other.IgnoreStderr
	}
	if other.DefaultArgs != merged.DefaultArgs && other.DefaultArgs != "" {
		merged.DefaultArgs = other.DefaultArgs
	}
	if other.Preamble != merged.Preamble && other.Preamble != "" {
		merged.Preamble = other.Preamble
	}
	if other.Description != merged.Description && other.Description != "" {
		merged.Description = other.Description
	}
	merged.ContinueWatchOnExit = other.ContinueWatchOnExit || merged.ContinueWatchOnExit

	for k, v := range other.Env {
		if merged.Env == nil {
			merged.Env = make(map[string]string)
		}
		merged.Env[k] = v
	}

	for k, v := range other.ExternalSecrets {
		if merged.ExternalSecrets == nil {
			merged.ExternalSecrets = make(map[string]secrets.ExternalSecret)
		}
		merged.ExternalSecrets[k] = v
	}

	if !equal(merged.Watch, other.Watch) && len(other.Watch) != 0 {
		merged.Watch = other.Watch
	}

	return merged
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

var sgConn net.Conn

func OpenUnixSocket() error {
	var err error
	sgConn, err = net.Dial("unix", "/tmp/sg.sock")
	return err
}

func getSecrets(ctx context.Context, name string, extSecrets map[string]secrets.ExternalSecret) (map[string]string, error) {
	secretsEnv := map[string]string{}

	if len(extSecrets) == 0 {
		return secretsEnv, nil
	}

	secretsStore, err := secrets.FromContext(ctx)
	if err != nil {
		return nil, errors.Errorf("failed to get secrets store: %v", err)
	}

	var errs error
	for envName, secret := range extSecrets {
		secretsEnv[envName], err = secretsStore.GetExternal(ctx, secret)
		if err != nil {
			errs = errors.Append(errs,
				errors.Wrapf(err, "failed to access secret %q for command %q", envName, name))
		}
	}
	return secretsEnv, errs
}

type startedCmd struct {
	*exec.Cmd
	opts   commandOptions
	cancel func()

	outEg  *pool.ErrorPool
	result chan error
}

type commandOptions struct {
	name   string
	exec   *exec.Cmd
	dir    string
	env    []string
	stdout outputOptions
	stderr outputOptions
}

type outputOptions struct {
	// When true, output will be ignored and not written to any writers
	ignore bool

	// when enabled, output will not be streamed to the writers until
	// after the process is begun, only captured for later retrieval
	buffer bool

	// Buffer that captures the output for error logging
	captured io.ReadWriter

	// Additional writers to write output to
	additionalWriters []io.Writer

	// Channel that is used to signal that output should start streaming
	// when buffer is true
	start chan struct{}
}

func startSgCmd(ctx context.Context, cmd SGConfigCommand, dir string, parentEnv map[string]string) (*startedCmd, error) {
	exec, err := cmd.GetExecCmd(ctx)
	if err != nil {
		return nil, err
	}

	secretsEnv, err := getSecrets(ctx, cmd.GetName(), cmd.GetExternalSecrets())
	if err != nil {
		std.Out.WriteLine(output.Styledf(output.StyleWarning, "[%s] %s %s",
			cmd.GetName(), output.EmojiFailure, err.Error()))
	}

	opts := commandOptions{
		name:   cmd.GetName(),
		exec:   exec,
		env:    makeEnv(parentEnv, secretsEnv, cmd.GetEnv()),
		dir:    dir,
		stdout: outputOptions{ignore: cmd.GetIgnoreStdout()},
		stderr: outputOptions{ignore: cmd.GetIgnoreStderr()},
	}

	if cmd.GetPreamble() != "" {
		std.Out.WriteLine(output.Styledf(output.StyleOrange, "[%s] %s %s", cmd.GetName(), output.EmojiInfo, cmd.GetPreamble()))
	}

	return startCmd(ctx, opts)
}

func startCmd(ctx context.Context, opts commandOptions) (*startedCmd, error) {
	sc := &startedCmd{
		opts: opts,
	}

	ctx, cancel := context.WithCancel(ctx)
	sc.cancel = func() {
		// The default cancel function will use a SIGKILL (9) which does
		// not allow processes to cleanup. If they have spawned child processes
		// those child processes will be orphaned and continue running.
		// SIGINT will instead gracefully shut down the process and child processes.
		if sc.Cmd.Process != nil {
			// We created a process group above which we kill here.
			pgid, err := syscall.Getpgid(sc.Cmd.Process.Pid)
			if err != nil {
				// Ignore Errno 3 (No such process) as this means the process has already exited
				if !errors.Is(err, syscall.Errno(0x3)) {
					panic(errors.Wrapf(err, "failed to get process group ID for %s (PID %d)", sc.opts.name, sc.Cmd.Process.Pid))
				}
				// note the minus sign; this signals that we want to kill the whole process group
			} else if err := syscall.Kill(-pgid, syscall.SIGINT); err != nil {
				panic(errors.Wrapf(err, "failed kill process group ID %d for cmd %s ", pgid, sc.opts.name))
			}
		}

		cancel()
	}
	// Register an interrput handler
	interrupt.Register(sc.cancel)

	sc.Cmd = opts.exec
	sc.Cmd.Dir = opts.dir
	sc.Cmd.Env = opts.env

	// This sets up a process group which we kill later.
	// This allows us to ensure that any child processes are killed as well when this exits
	// This will only work on POSIX systems
	sc.Cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := sc.connectOutput(ctx); err != nil {
		sc.cancel()
		return nil, err
	}

	if err := sc.Start(); err != nil {
		sc.cancel()
		return nil, err
	}
	return sc, nil
}

func (sc *startedCmd) connectOutput(ctx context.Context) error {
	stdoutWriter := sc.getOutputWriter(ctx, &sc.opts.stdout, "stdout")
	stderrWriter := sc.getOutputWriter(ctx, &sc.opts.stderr, "stderr")

	eg, err := process.PipeOutputUnbuffered(ctx, sc.Cmd, stdoutWriter, stderrWriter)
	if err != nil {
		return err
	}
	sc.outEg = eg

	return nil
}

func (sc *startedCmd) getOutputWriter(ctx context.Context, opts *outputOptions, outputName string) io.Writer {
	writers := opts.additionalWriters
	if writers == nil {
		writers = []io.Writer{}
	}
	if opts.captured == nil {
		opts.captured = &prefixSuffixSaver{N: 32 << 10}
	}
	writers = append(writers, opts.captured)

	if opts.ignore {
		std.Out.WriteLine(output.Styledf(output.StyleSuggestion, "Ignoring %s of %s", outputName, sc.opts.name))
	} else {
		// Create a channel to signal when output should start. If buffering is disabled, close
		// the channel so output starts immediately.
		opts.start = make(chan struct{})
		if !opts.buffer {
			close(opts.start)
		}

		writers = append(writers, newBufferedCmdLogger(ctx, sc.opts.name, std.Out.Output, opts.start))
	}

	if sgConn != nil {
		sink := func(data string) {
			sgConn.Write([]byte(fmt.Sprintf("%s: %s\n", sc.opts.name, data)))
		}
		writers = append(writers, process.NewLogger(ctx, sink))
	}

	return io.MultiWriter(writers...)
}

func (sc *startedCmd) Exit() <-chan error {
	if sc.result == nil {
		sc.result = make(chan error)
		go func() {
			sc.result <- sc.Wait()
		}()
	}
	return sc.result
}

func (sc *startedCmd) Wait() error {
	err := sc.wait()
	var e *exec.ExitError
	if errors.As(err, &e) {
		err = runErr{
			cmdName:  sc.opts.name,
			exitCode: e.ExitCode(),
			stderr:   sc.CapturedStderr(),
			stdout:   sc.CapturedStdout(),
		}
	}

	return err
}

func (sc *startedCmd) wait() error {
	if err := sc.outEg.Wait(); err != nil {
		return err
	}
	return sc.Cmd.Wait()
}

func (sc *startedCmd) CapturedStdout() string {
	return captured(sc.opts.stdout)
}

func (sc *startedCmd) CapturedStderr() string {
	return captured(sc.opts.stderr)
}

func captured(opts outputOptions) string {
	if opts.captured == nil {
		return ""
	}

	if output, err := io.ReadAll(opts.captured); err == nil {
		return string(output)
	}

	return ""
}

// Begins writing output to StdOut and StdErr if it was previously buffered
func (sc *startedCmd) StartOutput() {
	sc.startOutput(sc.opts.stdout)
	sc.startOutput(sc.opts.stderr)
}

func (sc *startedCmd) startOutput(opts outputOptions) {
	if opts.buffer && opts.start != nil {
		close(opts.start)
	}
}

// patternMatcher is writer which looks for a regular expression in the
// written bytes and calls a callback if a match is found
// by default it only looks for the matched pattern once
type patternMatcher struct {
	regex    *regexp.Regexp
	callback func()
	buffer   bytes.Buffer
	multi    bool
	disabled bool
}

func (writer *patternMatcher) Write(p []byte) (int, error) {
	if writer.disabled {
		return len(p), nil
	}
	n, err := writer.buffer.Write(p)
	if err != nil {
		return n, err
	}
	if writer.regex.MatchReader(&writer.buffer) {
		writer.callback()
		if !writer.multi {
			writer.disabled = true
		}
	}
	return n, err
}
