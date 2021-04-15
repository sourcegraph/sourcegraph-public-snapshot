package workspace

import (
	"context"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

func runGitCmd(ctx context.Context, dir string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Env = []string{
		// Don't use the system wide git config.
		"GIT_CONFIG_NOSYSTEM=1",
		// And also not any other, because they can mess up output, change defaults, .. which can do unexpected things.
		"GIT_CONFIG=/dev/null",
		// Set user.name and user.email in the local repository. The user name and
		// e-mail will eventually be ignored anyway, since we're just using the Git
		// repository to generate diffs, but we don't want git to generate alarming
		// looking warnings.
		"GIT_AUTHOR_NAME=Sourcegraph",
		"GIT_AUTHOR_EMAIL=batch-changes@sourcegraph.com",
		"GIT_COMMITTER_NAME=Sourcegraph",
		"GIT_COMMITTER_EMAIL=batch-changes@sourcegraph.com",
	}
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Wrapf(err, "'git %s' failed: %s", strings.Join(args, " "), out)
	}
	return out, nil
}
