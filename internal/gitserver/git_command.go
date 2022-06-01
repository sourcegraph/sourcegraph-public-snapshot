package gitserver

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
	"syscall"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GitCommand is an interface describing a git commands to be executed.
type GitCommand interface {
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

	// String returns string representation of the command (in fact prints args parameter of the command)
	String() string

	// StdoutReader returns an io.ReadCloser of stdout of c. If the command has a
	// non-zero return value, Read returns a non io.EOF error. Do not pass in a
	// started command.
	StdoutReader(ctx context.Context) (io.ReadCloser, error)
}

// LocalGitCommand is a GitCommand interface implementation which runs git commands against local file system.
//
// This struct uses composition with exec.RemoteGitCommand which already provides all necessary means to run commands against
// local system.
type LocalGitCommand struct {
	command *exec.Cmd

	// ReposDir is needed in order to LocalGitCommand be used like RemoteGitCommand (providing only repo name without its full path)
	// Unlike RemoteGitCommand, which is run against server who knows the directory where repos are located, LocalGitCommand is
	// run locally, therefore the knowledge about repos location should be provided explicitly by setting this field
	ReposDir       string
	repo           api.RepoName
	ensureRevision string
	args           []string
	exitStatus     int
}

func NewLocalGitCommand(repo api.RepoName, arg ...string) *LocalGitCommand {
	args := append([]string{git}, arg...)
	return &LocalGitCommand{
		command: exec.Command(git, arg...), // no need for including "git" in args here
		repo:    repo,
		args:    args,
	}
}

const NoReposDirErrorMsg = "No ReposDir provided, command cannot be run without it"

func (l *LocalGitCommand) DividedOutput(ctx context.Context) ([]byte, []byte, error) {
	if l.ReposDir == "" {
		log15.Error(NoReposDirErrorMsg)
		return nil, nil, errors.New(NoReposDirErrorMsg)
	}
	// cmd is a version of the command in LocalGitCommand with given context
	cmd := exec.CommandContext(ctx, git, l.Args()[1:]...) // stripping "git" itself
	var stderrBuf bytes.Buffer
	var stdoutBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	dir := protocol.NormalizeRepo(l.Repo())
	path := filepath.Join(l.ReposDir, filepath.FromSlash(string(dir)), ".git")
	cmd.Dir = path
	if cmd.Env == nil {
		// Do not strip out existing env when setting.
		cmd.Env = os.Environ()
	}
	cmd.Env = append(cmd.Env, "GIT_DIR="+path)

	err := cmd.Run()
	exitStatus := -10810
	if cmd.ProcessState != nil { // is nil if process failed to start
		exitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
	}
	l.exitStatus = exitStatus

	return stdoutBuf.Bytes(), stderrBuf.Bytes(), err
}

func (l *LocalGitCommand) Output(ctx context.Context) ([]byte, error) {
	stdout, _, err := l.DividedOutput(ctx)
	return stdout, err
}

func (l *LocalGitCommand) CombinedOutput(ctx context.Context) ([]byte, error) {
	stdout, stderr, err := l.DividedOutput(ctx)
	return append(stdout, stderr...), err
}

func (l *LocalGitCommand) DisableTimeout() {
	// No-op because there is no network request
}

func (l *LocalGitCommand) Repo() api.RepoName { return l.repo }

func (l *LocalGitCommand) Args() []string { return l.args }

func (l *LocalGitCommand) ExitStatus() int { return l.exitStatus }

func (l *LocalGitCommand) SetEnsureRevision(r string) { l.ensureRevision = r }

func (l *LocalGitCommand) EnsureRevision() string { return l.ensureRevision }

func (l *LocalGitCommand) StdoutReader(ctx context.Context) (io.ReadCloser, error) {
	output, err := l.CombinedOutput(ctx)
	return io.NopCloser(bytes.NewReader(output)), err
}

func (l *LocalGitCommand) String() string { return fmt.Sprintf("%q", l.Args()) }

// RemoteGitCommand represents a command to be executed remotely.
type RemoteGitCommand struct {
	repo           api.RepoName // the repository to execute the command in
	ensureRevision string
	args           []string
	noTimeout      bool
	exitStatus     int
	execFn         func(ctx context.Context, repo api.RepoName, op string, payload any) (resp *http.Response, err error)
}

// DividedOutput runs the command and returns its standard output and standard error.
func (c *RemoteGitCommand) DividedOutput(ctx context.Context) ([]byte, []byte, error) {
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
func (c *RemoteGitCommand) Output(ctx context.Context) ([]byte, error) {
	stdout, _, err := c.DividedOutput(ctx)
	return stdout, err
}

// CombinedOutput runs the command and returns its combined standard output and standard error.
func (c *RemoteGitCommand) CombinedOutput(ctx context.Context) ([]byte, error) {
	stdout, stderr, err := c.DividedOutput(ctx)
	return append(stdout, stderr...), err
}

func (c *RemoteGitCommand) DisableTimeout() {
	c.noTimeout = true
}

func (c *RemoteGitCommand) Repo() api.RepoName { return c.repo }

func (c *RemoteGitCommand) Args() []string { return c.args }

func (c *RemoteGitCommand) ExitStatus() int { return c.exitStatus }

func (c *RemoteGitCommand) SetEnsureRevision(r string) { c.ensureRevision = r }

func (c *RemoteGitCommand) EnsureRevision() string { return c.ensureRevision }

func (c *RemoteGitCommand) String() string { return fmt.Sprintf("%q", c.args) }

// StdoutReader returns an io.ReadCloser of stdout of c. If the command has a
// non-zero return value, Read returns a non io.EOF error. Do not pass in a
// started command.
func (c *RemoteGitCommand) StdoutReader(ctx context.Context) (io.ReadCloser, error) {
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
