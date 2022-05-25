package golang

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/generate"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
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

func Generate(ctx context.Context, args []string, progressBar bool, verbosity OutputVerbosityType) *generate.Report {
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
	rootDir, err := root.RepositoryRoot()
	if err != nil {
		return &generate.Report{Err: err}
	}

	// Grab the packages list
	pkgPaths, err := root.Run(run.Cmd(ctx, "go", "list", "./...")).Lines()
	if err != nil {
		return &generate.Report{Err: errors.Wrap(err, "go list ./...")}
	}

	// Run go generate on the packages list
	var goGenerateErr error
	if len(args) == 0 {
		// If no packages are given, go for everything but the exception.
		filtered := make([]string, 0, len(pkgPaths))
		for _, pkgPath := range pkgPaths {
			if !strings.Contains(pkgPath, "doc/cli/references") {
				filtered = append(filtered, pkgPath)
			}
		}
		if verbosity != QuietOutput {
			reportOut.WriteLine(output.Linef(output.EmojiInfo, output.StyleBold, "go generate ./... (excluding doc/cli/references)"))
		}
		goGenerateErr = runGoGenerate(ctx, filtered, progressBar, verbosity, &sb)
	} else {
		// Use the given packages.
		if verbosity != QuietOutput {
			reportOut.WriteLine(output.Linef(output.EmojiInfo, output.StyleBold, "go generate %s", strings.Join(args, " ")))
		}
		goGenerateErr = runGoGenerate(ctx, args, progressBar, verbosity, &sb)
	}

	if goGenerateErr != nil {
		return &generate.Report{Output: sb.String(), Err: errors.Wrap(err, "go generate")}
	}

	// Run goimports -w
	if verbosity != QuietOutput {
		reportOut.WriteLine(output.Linef(output.EmojiInfo, output.StyleBold, "goimports -w"))
	}
	if _, err := exec.LookPath("goimports"); err != nil {
		// Install goimports if not present
		err := run.Cmd(ctx, "go", "install", "golang.org/x/tools/cmd/goimports").
			Environ(os.Environ()).
			Env(map[string]string{
				// Install to local bin
				"GOBIN": filepath.Join(rootDir, ".bin"),
			}).
			Run().
			Stream(&sb)
		if err != nil {
			return &generate.Report{
				Output: sb.String(),
				Err:    errors.Wrap(err, "go install golang.org/x/tools/cmd/goimports returned an error"),
			}
		}

		err = root.Run(run.Cmd(ctx, "./.bin/goimports", "-w")).Stream(&sb)
		if err != nil {
			return &generate.Report{
				Output: sb.String(),
				Err:    errors.Wrap(err, "goimports -w"),
			}
		}
	} else {
		err = root.Run(run.Cmd(ctx, "goimports", "-w").Environ(os.Environ())).Stream(&sb)
		if err != nil {
			return &generate.Report{Output: sb.String(), Err: errors.Wrap(err, "goimports -w")}
		}
	}

	// Run go mod tidy
	if verbosity != QuietOutput {
		reportOut.WriteLine(output.Linef(output.EmojiInfo, output.StyleBold, "go mod tidy"))
	}

	err = root.Run(run.Cmd(ctx, "go", "mod", "tidy")).Stream(&sb)
	if err != nil {
		return &generate.Report{Output: sb.String(), Err: errors.Wrap(err, "go mod tidy")}
	}

	return &generate.Report{
		Output:   sb.String(),
		Duration: time.Since(start),
	}
}

func runGoGenerate(ctx context.Context, pkgPaths []string, progressBar bool, verbosity OutputVerbosityType, out io.Writer) (err error) {
	args := []string{"go", "generate"}
	if verbosity == VerboseOutput {
		args = append(args, "-x")
	}
	if progressBar {
		// If we want to display a progress bar we want the verbose output of `go
		// generate` so we can check which package has been generated.
		args = append(args, "-v")
	}
	args = append(args, pkgPaths...)

	if !progressBar {
		// If we don't want to display a progress bar we stream output to `out`
		// and are done.
		return root.Run(run.Cmd(ctx, args...)).Stream(out)
	}

	done := 0.0
	total := float64(len(pkgPaths))
	progress := std.Out.Progress([]output.ProgressBar{
		{Label: fmt.Sprintf("go generating %d packages", len(pkgPaths)), Max: total},
	}, nil)
	defer func() {
		if err != nil {
			progress.Destroy()
		} else {
			progress.Complete()
		}
	}()

	pkgMap := make(map[string]bool, len(pkgPaths))
	for _, pkg := range pkgPaths {
		pkgMap[strings.TrimPrefix(pkg, "github.com/sourcegraph/sourcegraph/")] = false
	}

	return root.Run(run.Cmd(ctx, args...)).StreamLines(func(line []byte) {
		if !bytes.HasSuffix(line, []byte(".go")) {
			return
		}

		dir := filepath.Dir(string(line))

		if current, ok := pkgMap[dir]; ok && !current {
			pkgMap[dir] = true

			if verbosity == VerboseOutput {
				progress.Writef("Generating %s...", dir)
			}

			done += 1.0
			progress.SetValue(0, done)
			progress.SetLabelAndRecalc(0, fmt.Sprintf("%d/%d packages generated", int(done), int(total)))
		}
	})
}
