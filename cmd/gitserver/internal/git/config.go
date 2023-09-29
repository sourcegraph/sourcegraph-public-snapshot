package git

import (
	"context"
	"os/exec"
	"strings"
	"syscall"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/executil"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/gitserverfs"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func ConfigGet(rcf *wrexec.RecordingCommandFactory, reposDir string, dir common.GitDir, key string) (string, error) {
	cmd := exec.Command("git", "config", "--get", key)
	dir.Set(cmd)
	wrappedCmd := rcf.WrapWithRepoName(context.Background(), log.NoOp(), gitserverfs.RepoNameFromDir(reposDir, dir), cmd)
	out, err := wrappedCmd.Output()
	if err != nil {
		// Exit code 1 means the key is not set.
		var e *exec.ExitError
		if errors.As(err, &e) && e.Sys().(syscall.WaitStatus).ExitStatus() == 1 {
			return "", nil
		}
		return "", errors.Wrapf(executil.WrapCmdError(cmd, err), "failed to get git config %s", key)
	}
	return strings.TrimSpace(string(out)), nil
}

func ConfigSet(rcf *wrexec.RecordingCommandFactory, reposDir string, dir common.GitDir, key, value string) error {
	cmd := exec.Command("git", "config", key, value)
	dir.Set(cmd)
	wrappedCmd := rcf.WrapWithRepoName(context.Background(), log.NoOp(), gitserverfs.RepoNameFromDir(reposDir, dir), cmd)
	err := wrappedCmd.Run()
	if err != nil {
		return errors.Wrapf(executil.WrapCmdError(cmd, err), "failed to set git config %s", key)
	}
	return nil
}

func ConfigUnset(rcf *wrexec.RecordingCommandFactory, reposDir string, dir common.GitDir, key string) error {
	cmd := exec.Command("git", "config", "--unset-all", key)
	dir.Set(cmd)
	wrappedCmd := rcf.WrapWithRepoName(context.Background(), log.NoOp(), gitserverfs.RepoNameFromDir(reposDir, dir), cmd)
	out, err := wrappedCmd.CombinedOutput()
	if err != nil {
		// Exit code 5 means the key is not set.
		var e *exec.ExitError
		if errors.As(err, &e) && e.Sys().(syscall.WaitStatus).ExitStatus() == 5 {
			return nil
		}
		return errors.Wrapf(executil.WrapCmdError(cmd, err), "failed to unset git config %s: %s", key, string(out))
	}
	return nil
}
