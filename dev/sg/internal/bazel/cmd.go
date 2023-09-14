package bazel

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func announce(command string, args ...string) {
	std.Out.WriteLine(
		output.Linef(
			"Running",
			output.StyleYellow,
			fmt.Sprintf("bazel %s %s", command, strings.Join(args, " "))),
	)
}

func Build(ctx context.Context, args ...string) error {
	announce("build", args...)
	cmd := exec.CommandContext(ctx, "bazel", append([]string{"build"}, args...)...)
	_, err := run.InRoot(cmd)
	return err
}

func Test(ctx context.Context, args ...string) error {
	announce("test", args...)
	cmd := exec.CommandContext(ctx, "bazel", append([]string{"test"}, args...)...)
	_, err := run.InRoot(cmd)
	return err
}

func Run(ctx context.Context, args ...string) error {
	announce("run", args...)
	cmd := exec.CommandContext(ctx, "bazel", append([]string{"run"}, args...)...)
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return err
	}

	cmd.Dir = repoRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
