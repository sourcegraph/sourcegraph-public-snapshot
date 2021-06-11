package main

import (
	"os"
	"os/exec"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
)

func runGitCmd(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Env = append(os.Environ(),
		// Don't use the system wide git config.
		"GIT_CONFIG_NOSYSTEM=1",
		// And also not any other, because they can mess up output, change defaults, .. which can do unexpected things.
		"GIT_CONFIG=/dev/null")

	return runCommandInRoot(cmd)
}

func runDockerCmd(args ...string) (string, error) {
	return runCommandInRoot(exec.Command("docker", args...))
}

func runCommandInRoot(cmd *exec.Cmd) (string, error) {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return "", err
	}

	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrapf(err, "'%s' failed: %s", strings.Join(cmd.Args, " "), out)
	}

	return string(out), nil
}
