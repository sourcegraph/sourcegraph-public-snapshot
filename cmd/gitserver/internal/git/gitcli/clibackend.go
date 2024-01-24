package gitcli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/executil"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ErrBadGitCommand is returned from the git CLI backend if the arguments provided
// are not allowed.
var ErrBadGitCommand = errors.New("bad git command, not allowed")

func NewBackend(logger log.Logger, rcf *wrexec.RecordingCommandFactory, dir common.GitDir, repoName api.RepoName) git.GitBackend {
	return &gitCLIBackend{
		logger:   logger,
		rcf:      rcf,
		dir:      dir,
		repoName: repoName,
	}
}

type gitCLIBackend struct {
	logger   log.Logger
	rcf      *wrexec.RecordingCommandFactory
	dir      common.GitDir
	repoName api.RepoName
}

func commandFailedError(err error, cmd wrexec.Cmder, stderr []byte) error {
	exitStatus := executil.UnsetExitStatus
	if cmd.Unwrap().ProcessState != nil { // is nil if process failed to start
		exitStatus = cmd.Unwrap().ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
	}

	return &CommandFailedError{
		Inner:      err,
		args:       cmd.Unwrap().Args,
		Stderr:     stderr,
		ExitStatus: exitStatus,
	}
}

type CommandFailedError struct {
	Stderr     []byte
	ExitStatus int
	Inner      error
	args       []string
}

func (e *CommandFailedError) Unwrap() error {
	return e.Inner
}

func (e *CommandFailedError) Error() string {
	return fmt.Sprintf("git command %v failed with status code %d (output: %q)", e.args, e.ExitStatus, e.Stderr)
}

const gitCommandTimeout = time.Minute

func (g *gitCLIBackend) gitCommand(ctx context.Context, args ...string) (wrexec.Cmder, context.CancelFunc, error) {
	var cancel context.CancelFunc = func() {}
	// If no deadline is set, use the default git command timeout.
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, gitCommandTimeout)
	}

	if !IsAllowedGitCmd(g.logger, args, g.dir) {
		blockedCommandExecutedCounter.Inc()
		return nil, cancel, ErrBadGitCommand
	}

	cmd := exec.Command("git", args...)
	g.dir.Set(cmd)

	return g.rcf.WrapWithRepoName(ctx, g.logger, g.repoName, cmd), cancel, nil
}

const maxStderrCapture = 1024

func (g *gitCLIBackend) runGitCommand(ctx context.Context, cmd wrexec.Cmder) (io.ReadCloser, error) {
	// Set up a limited buffer to capture stderr for error reporting.
	stderrBuf := bytes.NewBuffer(make([]byte, 0, maxStderrCapture))
	stderr := &limitWriter{W: stderrBuf, N: maxStderrCapture}
	cmd.Unwrap().Stderr = stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &cmdReader{
		ctx:        ctx,
		cmd:        cmd,
		ReadCloser: stdout,
		stderr:     stderrBuf,
		logger:     g.logger,
		git:        g,
		repoName:   g.repoName,
	}, nil
}

type cmdReader struct {
	io.ReadCloser
	ctx      context.Context
	cmd      wrexec.Cmder
	stderr   *bytes.Buffer
	buf      bytes.Buffer
	logger   log.Logger
	git      git.GitBackend
	repoName api.RepoName
}

func (rc *cmdReader) Read(p []byte) (n int, err error) {
	n, err = rc.ReadCloser.Read(p)
	writtenN, writeErr := rc.buf.Write(p[:n])
	if err == io.EOF {
		rc.ReadCloser.Close()

		if err := rc.cmd.Wait(); err != nil {
			if checkMaybeCorruptRepo(rc.ctx, rc.logger, rc.git, rc.repoName, rc.stderr.String()) {
				return n, common.ErrRepoCorrupted{Reason: rc.stderr.String()}
			}
			return n, commandFailedError(err, rc.cmd, rc.stderr.Bytes())
		}
	}
	if err == nil && writeErr != nil {
		return writtenN, writeErr
	}
	return n, err
}

// limitWriter is a io.Writer that writes to an W but discards after N bytes.
type limitWriter struct {
	W io.Writer // underling writer
	N int       // max bytes remaining
}

func (l *limitWriter) Write(p []byte) (int, error) {
	if l.N <= 0 {
		return len(p), nil
	}
	origLen := len(p)
	if len(p) > l.N {
		p = p[:l.N]
	}
	n, err := l.W.Write(p)
	l.N -= n
	if l.N <= 0 {
		// If we have written limit bytes, then we can include the discarded
		// part of p in the count.
		n = origLen
	}
	return n, err
}

// gitConfigMaybeCorrupt is a key we add to git config to signal that a repo may be
// corrupt on disk.
const gitConfigMaybeCorrupt = "sourcegraph.maybeCorruptRepo"

func checkMaybeCorruptRepo(ctx context.Context, logger log.Logger, git git.GitBackend, repo api.RepoName, stderr string) bool {
	if !stdErrIndicatesCorruption(stderr) {
		return false
	}

	logger = logger.With(log.String("repo", string(repo)), log.String("repo", string(repo)))
	logger.Warn("marking repo for re-cloning due to stderr output indicating repo corruption", log.String("stderr", stderr))

	// We set a flag in the config for the cleanup janitor job to fix. The janitor
	// runs every minute.
	err := git.Config().Set(ctx, gitConfigMaybeCorrupt, strconv.FormatInt(time.Now().Unix(), 10))
	if err != nil {
		logger.Error("failed to set maybeCorruptRepo config", log.Error(err))
	}

	return true
}

var (
	// objectOrPackFileCorruptionRegex matches stderr lines from git which indicate
	// that a repository's packfiles or commit objects might be corrupted.
	//
	// See https://github.com/sourcegraph/sourcegraph/issues/6676 for more
	// context.
	objectOrPackFileCorruptionRegex = lazyregexp.NewPOSIX(`^error: (Could not read|packfile) `)

	// objectOrPackFileCorruptionRegex matches stderr lines from git which indicate that
	// git's supplemental commit-graph might be corrupted.
	//
	// See https://github.com/sourcegraph/sourcegraph/issues/37872 for more
	// context.
	commitGraphCorruptionRegex = lazyregexp.NewPOSIX(`^fatal: commit-graph requires overflow generation data but has none`)
)

// stdErrIndicatesCorruption returns true if the provided stderr output from a git command indicates
// that there might be repository corruption.
func stdErrIndicatesCorruption(stderr string) bool {
	return objectOrPackFileCorruptionRegex.MatchString(stderr) || commitGraphCorruptionRegex.MatchString(stderr)
}
