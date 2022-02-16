package git

import (
	"context"
	"io"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// execSafe executes a Git subcommand iff it is allowed according to a allowlist.
//
// An error is only returned when there is a failure unrelated to the actual
// command being executed. If the executed command exits with a nonzero exit
// code, err == nil. This is similar to how http.Get returns a nil error for HTTP
// non-2xx responses.
//
// execSafe should NOT be exported. We want to limit direct git calls to this
// package.
func execSafe(ctx context.Context, repo api.RepoName, params []string) (stdout, stderr []byte, exitCode int, err error) {
	if Mocks.ExecSafe != nil {
		return Mocks.ExecSafe(params)
	}

	span, ctx := ot.StartSpanFromContext(ctx, "Git: execSafe")
	defer span.Finish()

	if len(params) == 0 {
		return nil, nil, 0, errors.New("at least one argument required")
	}

	if !gitdomain.IsAllowedGitCmd(params) {
		return nil, nil, 0, errors.Errorf("command failed: %q is not a allowed git command", params)
	}

	cmd := gitserver.DefaultClient.Command("git", params...)
	cmd.Repo = repo
	stdout, stderr, err = cmd.DividedOutput(ctx)
	exitCode = cmd.ExitStatus
	if exitCode != 0 && err != nil {
		err = nil // the error must just indicate that the exit code was nonzero
	}
	return stdout, stderr, exitCode, err
}

// execReader executes an arbitrary `git` command (`git [args...]`) and returns a
// reader connected to its stdout.
//
// execReader should NOT be exported. We want to limit direct git calls to this
// package.
func execReader(ctx context.Context, repo api.RepoName, args []string) (io.ReadCloser, error) {
	if Mocks.ExecReader != nil {
		return Mocks.ExecReader(args)
	}

	span, ctx := ot.StartSpanFromContext(ctx, "Git: ExecReader")
	span.SetTag("args", args)
	defer span.Finish()

	if !gitdomain.IsAllowedGitCmd(args) {
		return nil, errors.Errorf("command failed: %v is not a allowed git command", args)
	}
	cmd := gitserver.DefaultClient.Command("git", args...)
	cmd.Repo = repo
	return gitserver.StdoutReader(ctx, cmd)
}

// checkSpecArgSafety returns a non-nil err if spec begins with a "-", which
// could cause it to be interpreted as a git command line argument.
func checkSpecArgSafety(spec string) error {
	if strings.HasPrefix(spec, "-") {
		return errors.Errorf("invalid git revision spec %q (begins with '-')", spec)
	}
	return nil
}

func gitserverCmdFunc(repo api.RepoName) cmdFunc {
	return func(args []string) cmd {
		cmd := gitserver.DefaultClient.Command("git", args...)
		cmd.Repo = repo
		return cmd
	}
}

// cmdFunc is a func that creates a new executable Git command.
type cmdFunc func(args []string) cmd

// cmd is an executable Git command.
type cmd interface {
	Output(context.Context) ([]byte, error)
	String() string
}
