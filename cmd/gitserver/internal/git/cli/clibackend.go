package cli

import (
	"context"
	"os/exec"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
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

func commandFailedError(err error, cmd wrexec.Cmder, out []byte) error {
	return errors.Wrapf(err, "git command %v failed (output: %q)", cmd.Unwrap().Args, out)
}

func (g *gitCLIBackend) gitCommand(ctx context.Context, args ...string) (wrexec.Cmder, context.CancelFunc, error) {
	// TODO: Make configurable
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	cmd := exec.Command("git", args...)
	g.dir.Set(cmd)

	if !gitdomain.IsAllowedGitCmd(g.logger, cmd.Args, cmd.Dir) {
		blockedCommandExecutedCounter.Inc()
		return nil, cancel, ErrBadGitCommand
	}

	return g.rcf.WrapWithRepoName(ctx, g.logger, g.repoName, cmd), cancel, nil
}
