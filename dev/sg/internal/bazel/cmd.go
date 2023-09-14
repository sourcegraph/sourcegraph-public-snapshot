package bazel

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
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
	bazelCmd := fmt.Sprintf("bazel build %s", strings.Join(args, " "))
	// have to execute bazel inside bash since there are some env vars that gets used by bazel
	cmd := exec.CommandContext(ctx, "bash", []string{"-c", bazelCmd}...)
	return run.InteractiveInRoot(cmd)
}

func Test(ctx context.Context, args ...string) error {
	announce("test", args...)
	bazelCmd := fmt.Sprintf("bazel test %s", strings.Join(args, " "))
	// have to execute bazel inside bash since there are some env vars that gets used by bazel
	cmd := exec.CommandContext(ctx, "bash", []string{"-c", bazelCmd}...)
	return run.InteractiveInRoot(cmd)
}

func Run(ctx context.Context, args ...string) error {
	announce("run", args...)
	bazelCmd := fmt.Sprintf("bazel run %s", strings.Join(args, " "))
	// have to execute bazel inside bash since there are some env vars that gets used by bazel
	cmd := exec.CommandContext(ctx, "bash", []string{"-c", bazelCmd}...)
	return run.InteractiveInRoot(cmd)
}
