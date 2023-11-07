package run

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/root"
)

func GitCmd(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Env = append(os.Environ(),
		// Don't use the system wide git config.
		"GIT_CONFIG_NOSYSTEM=1",
		// And also not any other, because they can mess up output, change defaults, .. which can do unexpected things.
		"GIT_CONFIG=/dev/null")

	return InRoot(cmd)
}

func DockerCmd(args ...string) (string, error) {
	return InRoot(exec.Command("docker", args...))
}

type errorWithoutOutputer interface {
	ErrorWithoutOutput() string
}

type cmdInRootErr struct {
	err    error
	args   []string
	output string
}

func (e cmdInRootErr) Error() string {
	return fmt.Sprintf("'%s' failed: %s", strings.Join(e.args, " "), e.output)
}

func (e cmdInRootErr) ErrorWithoutOutput() string {
	return fmt.Sprintf("'%s' failed with %q", strings.Join(e.args, " "), e.err)
}

func (e cmdInRootErr) Unwrap() error { return e.err }

func InRoot(cmd *exec.Cmd) (string, error) {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return "", err
	}

	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), cmdInRootErr{err: err, args: cmd.Args, output: string(out)}
	}

	return string(out), nil
}

func BashInRoot(ctx context.Context, cmd string, env []string) (string, error) {
	c := exec.CommandContext(ctx, "bash", "-c", cmd)
	c.Env = env
	return InRoot(c)
}

func TrimResult(s string, err error) (string, error) {
	return strings.TrimSpace(s), err
}

func InteractiveInRoot(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return err
	}
	cmd.Dir = repoRoot
	return cmd.Run()
}
