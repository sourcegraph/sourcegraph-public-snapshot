package golang

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/generate"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

type OutputVerbosityType int

const (
	VerboseOutput OutputVerbosityType = iota
	NormalOutput
	QuietOutput
)

func Generate(ctx context.Context, args []string, verbosity OutputVerbosityType) *generate.Report {
	start := time.Now()
	var sb strings.Builder
	reportOut := output.NewOutput(&sb, output.OutputOpts{
		ForceColor: true,
		ForceTTY:   true,
	})

	// Save working directory
	cwd, err := os.Getwd()
	if err != nil {
		return &generate.Report{Err: err}
	}
	defer func() {
		os.Chdir(cwd)
	}()
	root, err := root.RepositoryRoot()
	if err != nil {
		return &generate.Report{Err: err}
	}
	err = os.Chdir(root)
	if err != nil {
		return &generate.Report{Err: err}
	}

	// Grab the packages list
	cmd := exec.CommandContext(ctx, "go", "list", "./...")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &generate.Report{Err: errors.Wrap(err, "could not run go list ./...")}
	}

	// Run go generate on the packages list
	if len(args) == 0 {
		// If no packages are given, go for everything but the exception.
		pkgPaths := strings.Split(string(out), "\n")
		filtered := make([]string, 0, len(pkgPaths))
		for _, pkgPath := range pkgPaths {
			if !strings.Contains(pkgPath, "doc/cli/references") {
				filtered = append(filtered, pkgPath)
			}
		}
		if verbosity != QuietOutput {
			reportOut.WriteLine(output.Linef(output.EmojiInfo, output.StyleBold, "go generate ./... (excluding doc/cli/references)"))
		}
		err = runGoGenerate(filtered, verbosity, &sb)
	} else {
		// Use the given packages.
		if verbosity != QuietOutput {
			reportOut.WriteLine(output.Linef(output.EmojiInfo, output.StyleBold, "go generate %s", strings.Join(args, " ")))
		}
		err = runGoGenerate(args, verbosity, &sb)
	}

	if err != nil {
		return &generate.Report{Output: sb.String(), Err: errors.Wrap(err, "could not run go generate ./...")}
	}

	// Run goimports -w
	if verbosity != QuietOutput {
		reportOut.WriteLine(output.Linef(output.EmojiInfo, output.StyleBold, "goimports -w"))
	}
	if _, err := exec.LookPath("goimports"); err != nil {
		// Install goimports if not present
		cmd := exec.CommandContext(ctx, "go", "install", "golang.org/x/tools/cmd/goimports")
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, fmt.Sprintf("GOBIN=%s", filepath.Join(root, ".bin")))
		if verbosity == VerboseOutput {
			cmd.Stderr = os.Stderr
			cmd.Stdout = &sb
		}
		err := cmd.Run()
		if err != nil {
			return &generate.Report{Output: sb.String(), Err: errors.Wrap(err, "go install golang.org/x/tools/cmd/goimports returned an error")}
		}

		cmd = exec.CommandContext(ctx, "./.bin/goimports", "-w")
		if verbosity == VerboseOutput {
			cmd.Stderr = os.Stderr
			cmd.Stdout = &sb
		}
		err = cmd.Run()
		if err != nil {
			return &generate.Report{Output: sb.String(), Err: errors.Wrap(err, "goimports -w returned an error")}
		}
	} else {
		cmd = exec.CommandContext(ctx, "goimports", "-w")
		cmd.Env = os.Environ()
		if verbosity == VerboseOutput {
			cmd.Stderr = os.Stderr
			cmd.Stdout = &sb
		}
		err := cmd.Run()
		if err != nil {
			return &generate.Report{Output: sb.String(), Err: errors.Wrap(err, "goimports -w returned an error")}
		}
	}

	// Run go mod tidy
	if verbosity != QuietOutput {
		reportOut.WriteLine(output.Linef(output.EmojiInfo, output.StyleBold, "go mod tidy"))
	}
	cmd = exec.CommandContext(ctx, "go", "mod", "tidy")
	if verbosity == VerboseOutput {
		cmd.Stderr = os.Stderr
		cmd.Stdout = &sb
	}
	err = cmd.Run()
	if err != nil {
		return &generate.Report{Output: sb.String(), Err: errors.Wrap(err, "go mod tidy returned an error")}
	}

	return &generate.Report{
		Output:   sb.String(),
		Duration: time.Since(start),
	}
}

func runGoGenerate(pkgPaths []string, verbosity OutputVerbosityType, out io.Writer) error {
	args := []string{"generate"}
	if verbosity == VerboseOutput {
		args = append(args, "-x")
	}
	args = append(args, pkgPaths...)
	cmd := exec.Command("go", args...)
	if verbosity == VerboseOutput {
		cmd.Stderr = os.Stderr
		cmd.Stdout = out
	}

	err := cmd.Start()
	if err != nil {
		return errors.Errorf("go generate returned an error: %w", err)
	}

	err = cmd.Wait()
	if err != nil {
		return errors.Errorf("go generate returned an error: %w", err)
	}
	return nil
}
