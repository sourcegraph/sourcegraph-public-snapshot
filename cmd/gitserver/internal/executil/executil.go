package executil

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/sourcegraph/conc/pool"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/cacert"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/trace" //nolint:staticcheck // OT is deprecated
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/process"
)

// ShortGitCommandTimeout returns the timeout for git commands that should not
// take a long time. Some commands such as "git archive" are allowed more time
// than "git rev-parse", so this will return an appropriate timeout given the
// command.
func ShortGitCommandTimeout(args []string) time.Duration {
	if len(args) < 1 {
		return time.Minute
	}
	switch args[0] {
	case "archive":
		// This is a long time, but this never blocks a user request for this
		// long. Even repos that are not that large can take a long time, for
		// example a search over all repos in an organization may have several
		// large repos. All of those repos will be competing for IO => we need
		// a larger timeout.
		return conf.GitLongCommandTimeout()

	case "ls-remote":
		return 30 * time.Second

	default:
		return time.Minute
	}
}

// UnsetExitStatus is a sentinel value for an unknown/unset exit status.
const UnsetExitStatus = -10810

// UpdateRunCommandMock sets the runCommand mock function for use in tests
func UpdateRunCommandMock(mock func(context.Context, *exec.Cmd) (int, error)) {
	runCommandMockMu.Lock()
	defer runCommandMockMu.Unlock()

	RunCommandMock = mock
}

// runCommmandMockMu protects runCommandMock against simultaneous access across
// multiple goroutines
var runCommandMockMu sync.RWMutex

// RunCommandMock is set by tests. When non-nil it is run instead of
// runCommand
var RunCommandMock func(context.Context, *exec.Cmd) (int, error)

// RunCommand runs the command and returns the exit status. All clients of this function should set the context
// in cmd themselves, but we have to pass the context separately here for the sake of tracing.
func RunCommand(ctx context.Context, cmd wrexec.Cmder) (exitCode int, err error) {
	runCommandMockMu.RLock()

	if RunCommandMock != nil {
		code, err := RunCommandMock(ctx, cmd.Unwrap())
		runCommandMockMu.RUnlock()
		return code, err
	}
	runCommandMockMu.RUnlock()

	tr, _ := trace.New(ctx, "runCommand",
		attribute.String("path", cmd.Unwrap().Path),
		attribute.StringSlice("args", cmd.Unwrap().Args),
		attribute.String("dir", cmd.Unwrap().Dir))
	defer func() {
		tr.SetAttributes(attribute.Int("exitCode", exitCode))
		tr.EndWithErr(&err)
	}()

	err = cmd.Run()
	exitStatus := UnsetExitStatus
	if cmd.Unwrap().ProcessState != nil { // is nil if process failed to start
		exitStatus = cmd.Unwrap().ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
	}
	return exitStatus, err
}

// RunCommandCombinedOutput runs the command with runCommand and returns its
// combined standard output and standard error.
func RunCommandCombinedOutput(ctx context.Context, cmd wrexec.Cmder) ([]byte, error) {
	var buf bytes.Buffer
	cmd.Unwrap().Stdout = &buf
	cmd.Unwrap().Stderr = &buf
	_, err := RunCommand(ctx, cmd)
	return buf.Bytes(), err
}

// RunRemoteGitCommand runs the command after applying the remote options.
func RunRemoteGitCommand(ctx context.Context, cmd wrexec.Cmder, configRemoteOpts bool) ([]byte, error) {
	if configRemoteOpts {
		// Inherit process environment. This allows admins to configure
		// variables like http_proxy/etc.
		if cmd.Unwrap().Env == nil {
			cmd.Unwrap().Env = os.Environ()
		}
		configureRemoteGitCommand(cmd.Unwrap(), tlsExternal())
	}

	var buf bytes.Buffer
	cmd.Unwrap().Stdout = &buf
	cmd.Unwrap().Stderr = &buf

	// We don't care about exitStatus, we just rely on error.
	_, err := RunCommand(ctx, cmd)

	return buf.Bytes(), err
}

// tlsExternal will create a new cache for this gitserer process and store the certificates set in
// the site config.
// This creates a long lived
var tlsExternal = conf.Cached(getTlsExternalDoNotInvoke)

type tlsConfig struct {
	// Whether to not verify the SSL certificate when fetching or pushing over
	// HTTPS.
	//
	// https://git-scm.com/docs/git-config#Documentation/git-config.txt-httpsslVerify
	SSLNoVerify bool

	// File containing the certificates to verify the peer with when fetching
	// or pushing over HTTPS.
	//
	// https://git-scm.com/docs/git-config#Documentation/git-config.txt-httpsslCAInfo
	SSLCAInfo string
}

// writeTempFile writes data to the TempFile with pattern. Returns the path of
// the tempfile.
func writeTempFile(pattern string, data []byte) (path string, err error) {
	f, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}

	defer func() {
		if err1 := f.Close(); err == nil {
			err = err1
		}
		// Cleanup if we fail to write
		if err != nil {
			path = ""
			os.Remove(f.Name())
		}
	}()

	n, err := f.Write(data)
	if err == nil && n < len(data) {
		return "", io.ErrShortWrite
	}

	return f.Name(), err
}

// getTlsExternalDoNotInvoke as the name suggests, exists as a function instead of being passed
// directly to conf.Cached below just so that we can test it.
func getTlsExternalDoNotInvoke() *tlsConfig {
	exp := conf.ExperimentalFeatures()
	c := exp.TlsExternal

	logger := log.Scoped("tlsExternal")

	if c == nil {
		return &tlsConfig{}
	}

	sslCAInfo := ""
	if len(c.Certificates) > 0 {
		var b bytes.Buffer
		for _, cert := range c.Certificates {
			b.WriteString(cert)
			b.WriteString("\n")
		}

		// git will ignore the system certificates when specifying SSLCAInfo,
		// so we additionally include the system certificates. Note: this only
		// works on linux, see cacert package for more information.
		root, err := cacert.System()
		if err != nil {
			logger.Error("failed to load system certificates for inclusion in SSLCAInfo. Git will now fail to speak to TLS services not specified in your TlsExternal site configuration.", log.Error(err))
		} else if len(root) == 0 {
			logger.Warn("no system certificates found for inclusion in SSLCAInfo. Git will now fail to speak to TLS services not specified in your TlsExternal site configuration.")
		}
		for _, cert := range root {
			b.Write(cert)
			b.WriteString("\n")
		}

		// We don't clean up the file since it has a process life time.
		p, err := writeTempFile("gitserver*.crt", b.Bytes())
		if err != nil {
			logger.Error("failed to create file holding tls.external.certificates for git", log.Error(err))
		} else {
			sslCAInfo = p
		}
	}

	return &tlsConfig{
		SSLNoVerify: c.InsecureSkipVerify,
		SSLCAInfo:   sslCAInfo,
	}
}

func ConfigureRemoteGitCommand(cmd *exec.Cmd) {
	configureRemoteGitCommand(cmd, tlsExternal())
}

func configureRemoteGitCommand(cmd *exec.Cmd, tlsConf *tlsConfig) {
	// We split here in case the first command is an absolute path to the executable
	// which allows us to safely match lower down
	_, executable := path.Split(cmd.Args[0])
	// As a special case we also support the experimental p4-fusion client which is
	// not run as a subcommand of git.
	if executable != "git" && executable != "p4-fusion" {
		panic(fmt.Sprintf("Only git or p4-fusion commands are supported, got %q", executable))
	}

	cmd.Env = append(cmd.Env, "GIT_ASKPASS=true") // disable password prompt

	// Suppress asking to add SSH host key to known_hosts (which will hang because
	// the command is non-interactive).
	//
	// And set a timeout to avoid indefinite hangs if the server is unreachable.
	cmd.Env = append(cmd.Env, "GIT_SSH_COMMAND=ssh -o BatchMode=yes -o ConnectTimeout=30")

	// Identify HTTP requests with a user agent. Please keep the git/ prefix because GitHub breaks the protocol v2
	// negotiation of clone URLs without a `.git` suffix (which we use) without it. Don't ask.
	cmd.Env = append(cmd.Env, "GIT_HTTP_USER_AGENT=git/Sourcegraph-Bot")

	if tlsConf.SSLNoVerify {
		cmd.Env = append(cmd.Env, "GIT_SSL_NO_VERIFY=true")
	}
	if tlsConf.SSLCAInfo != "" {
		cmd.Env = append(cmd.Env, "GIT_SSL_CAINFO="+tlsConf.SSLCAInfo)
	}

	extraArgs := []string{
		// Unset credential helper because the command is non-interactive.
		"-c", "credential.helper=",
	}

	if len(cmd.Args) > 1 && cmd.Args[1] != "ls-remote" {
		// Use Git protocol version 2 for all commands except for ls-remote because it actually decreases the performance of ls-remote.
		// https://opensource.googleblog.com/2018/05/introducing-git-protocol-version-2.html
		extraArgs = append(extraArgs, "-c", "protocol.version=2")
	}

	if executable == "p4-fusion" {
		extraArgs = removeUnsupportedP4Args(extraArgs)
	}

	cmd.Args = append(cmd.Args[:1], append(extraArgs, cmd.Args[1:]...)...)
}

// removeUnsupportedP4Args removes all -c arguments as `p4-fusion` command doesn't
// support -c argument and passing this causes warning logs.
func removeUnsupportedP4Args(args []string) []string {
	if len(args) == 0 {
		return args
	}

	idx := 0
	foundC := false
	for _, arg := range args {
		if arg == "-c" {
			// removing any -c
			foundC = true
		} else if foundC {
			// removing the argument following -c and resetting the flag
			foundC = false
		} else {
			// keep the argument
			args[idx] = arg
			idx++
		}
	}
	args = args[:idx]
	return args
}

// WrapCmdError will wrap errors for cmd to include the arguments. If the error
// is an exec.ExitError and cmd was invoked with Output(), it will also include
// the captured stderr.
func WrapCmdError(cmd *exec.Cmd, err error) error {
	if err == nil {
		return nil
	}
	var e *exec.ExitError
	if errors.As(err, &e) {
		return errors.Wrapf(err, "%s %s failed with stderr: %s", cmd.Path, strings.Join(cmd.Args, " "), string(e.Stderr))
	}
	return errors.Wrapf(err, "%s %s failed", cmd.Path, strings.Join(cmd.Args, " "))
}

type RedactorFunc func(string) string

// The passed cmd should be bound to the passed context.
func RunCommandWriteOutput(ctx context.Context, cmd wrexec.Cmder, writer io.Writer, redactor RedactorFunc) (int, error) {
	exitStatus := UnsetExitStatus

	// Create a cancel context so that on exit we always properly close the command
	// pipes attached later by process.PipeOutputUnbuffered.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Make sure we only write to the writer from one goroutine at a time, either
	// stdout or stderr.
	syncWriter := newSynchronizedWriter(writer)

	outputRedactor := func(w io.Writer, r io.Reader) error {
		sc := process.NewOutputScannerWithSplit(r, scanLinesWithCRLF)
		for sc.Scan() {
			line := sc.Text()
			if _, err := fmt.Fprint(w, redactor(line)); err != nil {
				return err
			}
		}
		// We can ignore ErrClosed because we get that if a process crashes, it will
		// be handled by cmd.Wait.
		if err := sc.Err(); err != nil && !errors.Is(err, fs.ErrClosed) {
			return err
		}
		return nil
	}

	eg, err := process.PipeProcessOutput(
		ctx,
		cmd,
		syncWriter,
		syncWriter,
		outputRedactor,
	)
	if err != nil {
		return exitStatus, errors.Wrap(err, "failed to pipe output")
	}

	if err = cmd.Start(); err != nil {
		return exitStatus, errors.Wrap(err, "failed to start command")
	}

	// Wait for either the command to finish (aka the pipewriters get closed), or
	// for a context cancelation.
	select {
	case <-ctx.Done():
	case err := <-watchErrGroup(eg):
		if err != nil {
			return exitStatus, errors.Wrap(err, "failed to read output")
		}
	}

	err = cmd.Wait()

	if ps := cmd.Unwrap().ProcessState; ps != nil && ps.Sys() != nil {
		if ws, ok := ps.Sys().(syscall.WaitStatus); ok {
			exitStatus = ws.ExitStatus()
		}
	}

	return exitStatus, errors.Wrap(err, "command failed")
}

// watchErrGroup turns a pool.ErrorPool into a channel that will receive the error
// returned from the pool once it returns.
func watchErrGroup(g *pool.ErrorPool) <-chan error {
	ch := make(chan error)
	go func() {
		ch <- g.Wait()
		close(ch)
	}()

	return ch
}

// scanLinesWithCRLF is a modified version of bufio.ScanLines that retains
// the trailing newline byte(s) in the returned token and splits on either CR
// or LF.
func scanLinesWithCRLF(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexAny(data, "\r\n"); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0 : i+1], nil
	}

	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}

	// Request more data.
	return 0, nil, nil
}

func newSynchronizedWriter(w io.Writer) *synchronizedWriter {
	return &synchronizedWriter{writer: w}
}

type synchronizedWriter struct {
	mu     sync.Mutex
	writer io.Writer
}

func (sw *synchronizedWriter) Write(p []byte) (n int, err error) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return sw.writer.Write(p)
}
