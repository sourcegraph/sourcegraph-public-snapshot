package p4server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/p4server/protocol"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// P4Command is an interface describing a git commands to be executed.
type P4Command interface {
	// DividedOutput runs the command and returns its standard output and standard error.
	DividedOutput(ctx context.Context) ([]byte, []byte, error)

	// Output runs the command and returns its standard output.
	Output(ctx context.Context) ([]byte, error)

	// CombinedOutput runs the command and returns its combined standard output and standard error.
	CombinedOutput(ctx context.Context) ([]byte, error)

	// DisableTimeout turns command timeout off
	DisableTimeout()

	// Repo returns repo against which the command is run
	Repo() api.RepoName

	// Args returns arguments of the command
	Args() []string

	// ExitStatus returns exit status of the command
	ExitStatus() int

	// SetEnsureRevision sets the revision which should be ensured when the command is ran
	SetEnsureRevision(r string)

	// EnsureRevision returns ensureRevision parameter of the command
	EnsureRevision() string

	// SetStdin will write b to stdin when running the command.
	SetStdin(b []byte)

	// String returns string representation of the command (in fact prints args parameter of the command)
	String() string

	// StdoutReader returns an io.ReadCloser of stdout of c. If the command has a
	// non-zero return value, Read returns a non io.EOF error. Do not pass in a
	// started command.
	StdoutReader(ctx context.Context) (io.ReadCloser, error)
}

// LocalP4Command is a P4Command interface implementation which runs git commands against local file system.
//
// This struct uses composition with exec.RemoteP4Command which already provides all necessary means to run commands against
// local system.
type LocalP4Command struct {
	Logger log.Logger

	// ReposDir is needed in order to LocalP4Command be used like RemoteP4Command (providing only repo name without its full path)
	// Unlike RemoteP4Command, which is run against server who knows the directory where repos are located, LocalP4Command is
	// run locally, therefore the knowledge about repos location should be provided explicitly by setting this field
	ReposDir       string
	repo           api.RepoName
	ensureRevision string
	args           []string
	stdin          []byte
	exitStatus     int
}

func NewLocalP4Command(repo api.RepoName, arg ...string) *LocalP4Command {
	args := append([]string{git}, arg...)
	return &LocalP4Command{
		repo:   repo,
		args:   args,
		Logger: log.Scoped("local", "local git command logger"),
	}
}

const NoReposDirErrorMsg = "No ReposDir provided, command cannot be run without it"

func (l *LocalP4Command) DividedOutput(ctx context.Context) ([]byte, []byte, error) {
	if l.ReposDir == "" {
		l.Logger.Error(NoReposDirErrorMsg)
		return nil, nil, errors.New(NoReposDirErrorMsg)
	}
	cmd := exec.CommandContext(ctx, git, l.Args()[1:]...) // stripping "git" itself
	var stderrBuf bytes.Buffer
	var stdoutBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	cmd.Stdin = bytes.NewReader(l.stdin)

	dir := protocol.NormalizeRepo(l.Repo())
	repoPath := filepath.Join(l.ReposDir, filepath.FromSlash(string(dir)))
	gitPath := filepath.Join(repoPath, ".git")
	cmd.Dir = repoPath
	if cmd.Env == nil {
		// Do not strip out existing env when setting.
		cmd.Env = os.Environ()
	}
	cmd.Env = append(cmd.Env, "GIT_DIR="+gitPath)

	err := cmd.Run()
	exitStatus := -10810         // sentinel value to indicate not set
	if cmd.ProcessState != nil { // is nil if process failed to start
		exitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
	}
	l.exitStatus = exitStatus

	// We want to treat actions on files that don't exist as an os.ErrNotExist
	if err != nil && strings.Contains(stderrBuf.String(), "does not exist in") {
		err = os.ErrNotExist
	}

	return stdoutBuf.Bytes(), bytes.TrimSpace(stderrBuf.Bytes()), err
}

func (l *LocalP4Command) Output(ctx context.Context) ([]byte, error) {
	stdout, _, err := l.DividedOutput(ctx)
	return stdout, err
}

func (l *LocalP4Command) CombinedOutput(ctx context.Context) ([]byte, error) {
	stdout, stderr, err := l.DividedOutput(ctx)
	return append(stdout, stderr...), err
}

func (l *LocalP4Command) DisableTimeout() {
	// No-op because there is no network request
}

func (l *LocalP4Command) Repo() api.RepoName { return l.repo }

func (l *LocalP4Command) Args() []string { return l.args }

func (l *LocalP4Command) ExitStatus() int { return l.exitStatus }

func (l *LocalP4Command) SetEnsureRevision(r string) { l.ensureRevision = r }

func (l *LocalP4Command) EnsureRevision() string { return l.ensureRevision }

func (l *LocalP4Command) SetStdin(b []byte) { l.stdin = b }

func (l *LocalP4Command) StdoutReader(ctx context.Context) (io.ReadCloser, error) {
	output, err := l.Output(ctx)
	return io.NopCloser(bytes.NewReader(output)), err
}

func (l *LocalP4Command) String() string { return fmt.Sprintf("%q", l.Args()) }

// RemoteP4Command represents a command to be executed remotely.
type RemoteP4Command struct {
	repo           api.RepoName // the repository to execute the command in
	ensureRevision string
	args           []string
	stdin          []byte
	noTimeout      bool
	exitStatus     int
	execFn         func(ctx context.Context, repo api.RepoName, op string, payload any) (resp *http.Response, err error)
}

// DividedOutput runs the command and returns its standard output and standard error.
func (c *RemoteP4Command) DividedOutput(ctx context.Context) ([]byte, []byte, error) {
	rc, trailer, err := c.sendExec(ctx)
	if err != nil {
		return nil, nil, err
	}

	stdout, err := io.ReadAll(rc)
	if err != nil {
		return nil, nil, errors.Wrap(err, "reading exec output")
	}
	if err := rc.Close(); err != nil {
		return nil, nil, errors.Wrap(err, "closing exec reader")
	}

	c.exitStatus, err = strconv.Atoi(trailer.Get("X-Exec-Exit-Status"))
	if err != nil {
		return nil, nil, err
	}

	stderr := []byte(trailer.Get("X-Exec-Stderr"))
	if errorMsg := trailer.Get("X-Exec-Error"); errorMsg != "" {
		return stdout, stderr, errors.New(errorMsg)
	}

	return stdout, stderr, nil
}

// Output runs the command and returns its standard output.
func (c *RemoteP4Command) Output(ctx context.Context) ([]byte, error) {
	stdout, _, err := c.DividedOutput(ctx)
	return stdout, err
}

// CombinedOutput runs the command and returns its combined standard output and standard error.
func (c *RemoteP4Command) CombinedOutput(ctx context.Context) ([]byte, error) {
	stdout, stderr, err := c.DividedOutput(ctx)
	return append(stdout, stderr...), err
}

func (c *RemoteP4Command) DisableTimeout() {
	c.noTimeout = true
}

func (c *RemoteP4Command) Repo() api.RepoName { return c.repo }

func (c *RemoteP4Command) Args() []string { return c.args }

func (c *RemoteP4Command) ExitStatus() int { return c.exitStatus }

func (c *RemoteP4Command) SetEnsureRevision(r string) { c.ensureRevision = r }

func (c *RemoteP4Command) EnsureRevision() string { return c.ensureRevision }

func (c *RemoteP4Command) SetStdin(b []byte) { c.stdin = b }

func (c *RemoteP4Command) String() string { return fmt.Sprintf("%q", c.args) }

// StdoutReader returns an io.ReadCloser of stdout of c. If the command has a
// non-zero return value, Read returns a non io.EOF error. Do not pass in a
// started command.
func (c *RemoteP4Command) StdoutReader(ctx context.Context) (io.ReadCloser, error) {
	rc, trailer, err := c.sendExec(ctx)
	if err != nil {
		return nil, err
	}

	return &cmdReader{
		rc:      rc,
		trailer: trailer,
	}, nil
}

type cmdReader struct {
	rc      io.ReadCloser
	trailer http.Header
}

func (c *cmdReader) Read(p []byte) (int, error) {
	n, err := c.rc.Read(p)
	if err == io.EOF {
		stderr := c.trailer.Get("X-Exec-Stderr")
		if len(stderr) > 100 {
			stderr = stderr[:100] + "... (truncated)"
		}
		if errorMsg := c.trailer.Get("X-Exec-Error"); errorMsg != "" {
			return 0, errors.Errorf("%s (stderr: %q)", errorMsg, stderr)
		}
		if exitStatus := c.trailer.Get("X-Exec-Exit-Status"); exitStatus != "0" {
			return 0, errors.Errorf("non-zero exit status: %s (stderr: %q)", exitStatus, stderr)
		}
	}
	return n, err
}

func (c *cmdReader) Close() error {
	return c.rc.Close()
}
