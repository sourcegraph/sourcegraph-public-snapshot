package run

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/sourcegraph/conc/pool"

	"github.com/sourcegraph/sourcegraph/dev/sg/root"
)

func outputPath() ([]byte, error) {
	// Get the output directory from Bazel, which varies depending on which OS
	// we're running against.
	cmd := exec.Command("bazel", "info", "output_path")
	return cmd.Output()
}

// binLocation returns the path on disk where Bazel is putting the binary
// associated with a given target.
func binLocation(target string) (string, error) {
	baseOutput, err := outputPath()
	if err != nil {
		return "", err
	}
	// Trim "bazel-out" because the next bazel query will include it.
	outputPath := strings.TrimSuffix(strings.TrimSpace(string(baseOutput)), "bazel-out")

	// Get the binary from the specific target.
	cmd := exec.Command("bazel", "cquery", target, "--output=files")
	baseOutput, err = cmd.Output()
	if err != nil {
		return "", err
	}
	binPath := strings.TrimSpace(string(baseOutput))

	return fmt.Sprintf("%s%s", outputPath, binPath), nil
}

func BazelCommands(ctx context.Context, parentEnv map[string]string, verbose bool, cmds ...BazelCommand) error {
	if len(cmds) == 0 {
		// no Bazel commands so we return
		return nil
	}

	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return err
	}

	var targets []string
	for _, cmd := range cmds {
		targets = append(targets, cmd.Target)
	}

	ibazel := newIBazel(repoRoot, targets...)

	p := pool.New().WithContext(ctx).WithCancelOnError()
	p.Go(func(ctx context.Context) error {
		return ibazel.Start(ctx, repoRoot)
	})

	for _, bc := range cmds {
		bc := bc
		p.Go(func(ctx context.Context) error {
			return bc.Start(ctx, repoRoot, parentEnv)
		})
	}

	return p.Wait()
}
