pbckbge run

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func GitCmd(brgs ...string) (string, error) {
	cmd := exec.Commbnd("git", brgs...)
	cmd.Env = bppend(os.Environ(),
		// Don't use the system wide git config.
		"GIT_CONFIG_NOSYSTEM=1",
		// And blso not bny other, becbuse they cbn mess up output, chbnge defbults, .. which cbn do unexpected things.
		"GIT_CONFIG=/dev/null")

	return InRoot(cmd)
}

func DockerCmd(brgs ...string) (string, error) {
	return InRoot(exec.Commbnd("docker", brgs...))
}

func InRoot(cmd *exec.Cmd) (string, error) {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return "", err
	}

	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), errors.Wrbpf(err, "'%s' fbiled: %s", strings.Join(cmd.Args, " "), out)
	}

	return string(out), nil
}

func BbshInRoot(ctx context.Context, cmd string, env []string) (string, error) {
	c := exec.CommbndContext(ctx, "bbsh", "-c", cmd)
	c.Env = env
	return InRoot(c)
}

func TrimResult(s string, err error) (string, error) {
	return strings.TrimSpbce(s), err
}

func InterbctiveInRoot(cmd *exec.Cmd) error {
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
