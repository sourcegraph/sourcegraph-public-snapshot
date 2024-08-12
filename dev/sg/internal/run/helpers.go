package run

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

func GitCmd(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Env = append(os.Environ(),
		// Don't use the system wide git config.
		"GIT_CONFIG_NOSYSTEM=1",
		// And also not any other, because they can mess up output, change defaults, .. which can do unexpected things.
		"GIT_CONFIG=/dev/null")

	return InRoot(cmd, InRootArgs{})
}

func DockerCmd(args ...string) (string, error) {
	return InRoot(exec.Command("docker", args...), InRootArgs{})
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
	return fmt.Sprintf(`'%s' failed: err = "%s", output = "%s"`, strings.Join(e.args, " "), e.err.Error(), e.output)

}

func (e cmdInRootErr) ErrorWithoutOutput() string {
	return fmt.Sprintf("'%s' failed with %q", strings.Join(e.args, " "), e.err)
}

func (e cmdInRootErr) Unwrap() error { return e.err }

type InRootArgs struct {
	ExtractBazelError bool
}

func InRoot(cmd *exec.Cmd, args InRootArgs) (string, error) {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return "", err
	}

	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		output := string(out)
		if args.ExtractBazelError {
			// this is still experimental and currently only works for bazel errors
			output = bazelErrorExtractor(out)
		}
		return string(out), cmdInRootErr{err: err, args: cmd.Args, output: output}
	}

	return string(out), nil
}

// `(?m)` enables multiline mode, `^` matches the start of each line
var errorRegex = lazyregexp.New(`(?m)^ERROR:.*$`)

func bazelErrorExtractor(input []byte) string {
	// Find all matches
	matches := errorRegex.FindAll(input, -1)

	// Convert [][]byte to string
	var errorLines string
	for _, match := range matches {
		errorLines += string(match) + "\n"
	}

	return errorLines
}

func SplitOutputInRoot(cmd *exec.Cmd, stdout, stderr io.Writer) error {
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return err
	}
	cmd.Dir = repoRoot
	return cmd.Run()
}

type BashInRootArgs struct {
	Env               []string
	ExtractBazelError bool
}

func BashInRoot(ctx context.Context, cmd string, args BashInRootArgs) (string, error) {
	c := exec.CommandContext(ctx, "bash", "-c", cmd)
	c.Env = args.Env
	return InRoot(c, InRootArgs{
		ExtractBazelError: args.ExtractBazelError,
	})
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
