package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"sync"
	"syscall"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server/internal/cacert"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/trace" //nolint:staticcheck // OT is deprecated
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"go.opentelemetry.io/otel/attribute"
)

// unsetExitStatus is a sentinel value for an unknown/unset exit status.
const unsetExitStatus = -10810

// updateRunCommandMock sets the runCommand mock function for use in tests
func updateRunCommandMock(mock func(context.Context, *exec.Cmd) (int, error)) {
	runCommandMockMu.Lock()
	defer runCommandMockMu.Unlock()

	runCommandMock = mock
}

// runCommmandMockMu protects runCommandMock against simultaneous access across
// multiple goroutines
var runCommandMockMu sync.RWMutex

// runCommandMock is set by tests. When non-nil it is run instead of
// runCommand
var runCommandMock func(context.Context, *exec.Cmd) (int, error)

// runCommand runs the command and returns the exit status. All clients of this function should set the context
// in cmd themselves, but we have to pass the context separately here for the sake of tracing.
func runCommand(ctx context.Context, cmd wrexec.Cmder) (exitCode int, err error) {
	runCommandMockMu.RLock()

	if runCommandMock != nil {
		code, err := runCommandMock(ctx, cmd.Unwrap())
		runCommandMockMu.RUnlock()
		return code, err
	}
	runCommandMockMu.RUnlock()

	tr, _ := trace.New(ctx, "gitserver", "runCommand",
		attribute.String("path", cmd.Unwrap().Path),
		attribute.StringSlice("args", cmd.Unwrap().Args),
		attribute.String("dir", cmd.Unwrap().Dir))
	defer func() {
		if err != nil {
			tr.SetAttributes(attribute.Int("exitCode", exitCode))
		}
		tr.FinishWithErr(&err)
	}()

	err = cmd.Run()
	exitStatus := unsetExitStatus
	if cmd.Unwrap().ProcessState != nil { // is nil if process failed to start
		exitStatus = cmd.Unwrap().ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
	}
	return exitStatus, err
}

// runCommandCombinedOutput runs the command with runCommand and returns its
// combined standard output and standard error.
func runCommandCombinedOutput(ctx context.Context, cmd wrexec.Cmder) ([]byte, error) {
	var buf bytes.Buffer
	cmd.Unwrap().Stdout = &buf
	cmd.Unwrap().Stderr = &buf
	_, err := runCommand(ctx, cmd)
	cmd.CombinedOutput()
	return buf.Bytes(), err
}

// runRemoteGitCommand runs the command after applying the remote options. If
// progress is not nil, all output is written to it in a separate goroutine.
func runRemoteGitCommand(ctx context.Context, cmd wrexec.Cmder, configRemoteOpts bool, progress io.Writer) ([]byte, error) {
	if configRemoteOpts {
		// Inherit process environment. This allows admins to configure
		// variables like http_proxy/etc.
		if cmd.Unwrap().Env == nil {
			cmd.Unwrap().Env = os.Environ()
		}
		configureRemoteGitCommand(cmd.Unwrap(), tlsExternal())
	}

	var b interface {
		Bytes() []byte
	}

	if progress != nil {
		var pw progressWriter
		mr := io.MultiWriter(&pw, progress)
		cmd.Unwrap().Stdout = mr
		cmd.Unwrap().Stderr = mr
		b = &pw
	} else {
		var buf bytes.Buffer
		cmd.Unwrap().Stdout = &buf
		cmd.Unwrap().Stderr = &buf
		b = &buf
	}

	// We don't care about exitStatus, we just rely on error.
	_, err := runCommand(ctx, cmd)

	return b.Bytes(), err
}

// tlsExternal will create a new cache for this gitserer process and store the certificates set in
// the site config.
// This creates a long lived
var tlsExternal = conf.Cached(getTlsExternalDoNotInvoke)

// progressWriter is an io.Writer that writes to a buffer.
// '\r' resets the write offset to the index after last '\n' in the buffer,
// or the beginning of the buffer if a '\n' has not been written yet.
type progressWriter struct {
	// writeOffset is the offset in buf where the next write should begin.
	writeOffset int

	// afterLastNewline is the index after the last '\n' in buf
	// or 0 if there is no '\n' in buf.
	afterLastNewline int

	buf []byte
}

func (w *progressWriter) Write(p []byte) (n int, err error) {
	l := len(p)
	for {
		if len(p) == 0 {
			// If p ends in a '\r' we still want to include that in the buffer until it is overwritten.
			break
		}
		idx := bytes.IndexAny(p, "\r\n")
		if idx == -1 {
			w.buf = append(w.buf[:w.writeOffset], p...)
			w.writeOffset = len(w.buf)
			break
		}
		switch p[idx] {
		case '\n':
			w.buf = append(w.buf[:w.writeOffset], p[:idx+1]...)
			w.writeOffset = len(w.buf)
			w.afterLastNewline = len(w.buf)
			p = p[idx+1:]
		case '\r':
			w.buf = append(w.buf[:w.writeOffset], p[:idx+1]...)
			// Record that our next write should overwrite the data after the most recent newline.
			// Don't slice it off immediately here, because we want to be able to return that output
			// until it is overwritten.
			w.writeOffset = w.afterLastNewline
			p = p[idx+1:]
		default:
			panic(fmt.Sprintf("unexpected char %q", p[idx]))
		}
	}
	return l, nil
}

// String returns the contents of the buffer as a string.
func (w *progressWriter) String() string {
	return string(w.buf)
}

// Bytes returns the contents of the buffer.
func (w *progressWriter) Bytes() []byte {
	return w.buf
}

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

	logger := log.Scoped("tlsExternal", "Global TLS/SSL settings for Sourcegraph to use when communicating with code hosts.")

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
