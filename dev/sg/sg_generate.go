package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	stdOut              = stdout.Out
	generateFlagSet     = flag.NewFlagSet("sg generate", flag.ExitOnError)
	generateVerboseFlag = generateFlagSet.Bool("v", false, "Display output from go generate")
	generateQuietFlag   = generateFlagSet.Bool("q", false, "Suppress all output but errors from go generate")

	generateCommand = &ffcli.Command{
		Name:       "generate",
		ShortUsage: "sg generate",
		FlagSet:    generateFlagSet,
		Exec:       generateExec,
	}
)

type generateVerbosityType int

const (
	generateVerbose generateVerbosityType = iota
	generateNormal
	generateQuiet
)

func generateExec(ctx context.Context, args []string) error {
	if *generateVerboseFlag && *generateQuietFlag {
		return errors.Errorf("-q and -v flags are exclusive")
	}
	if *generateVerboseFlag {
		return generateDo(ctx, generateVerbose)
	} else if *generateQuietFlag {
		return generateDo(ctx, generateQuiet)
	} else {
		return generateDo(ctx, generateNormal)
	}
}

func generateDo(ctx context.Context, verbosity generateVerbosityType) error {
	// Save working directory
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	defer func() {
		os.Chdir(cwd)
	}()
	root, err := root.RepositoryRoot()
	if err != nil {
		return err
	}
	err = os.Chdir(root)
	if err != nil {
		return err
	}

	// Grab the packages list
	cmd := exec.CommandContext(ctx, "go", "list", "./...")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "could not run go list ./...")
	}

	// Run go generate on the packages list
	if verbosity != generateQuiet {
		stdOut.WriteLine(output.Linef(output.EmojiInfo, output.StyleBold, "go generate ./... (excluding doc/cli/references)"))
	}
	pkgPaths := strings.Split(string(out), "\n")
	filtered := make([]string, 0, len(pkgPaths))
	for _, pkgPath := range pkgPaths {
		if !strings.Contains(pkgPath, "doc/cli/references") {
			filtered = append(filtered, pkgPath)
		}
	}
	err = generateGoGenerate(filtered, verbosity)
	if err != nil {
		return errors.Wrap(err, "could not run go generate ./...")
	}

	// Run goimports -w
	if verbosity != generateQuiet {
		stdOut.WriteLine(output.Linef(output.EmojiInfo, output.StyleBold, "goimports -w"))
	}
	if _, err := exec.LookPath("goimports"); err != nil {
		// Install goimports if not present
		cmd := exec.CommandContext(ctx, "go", "install", "golang.org/x/tools/cmd/goimports")
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, fmt.Sprintf("GOBIN=%s", filepath.Join(root, ".bin")))
		if verbosity == generateVerbose {
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
		}
		err := cmd.Run()
		if err != nil {
			return errors.Wrap(err, "go install golang.org/x/tools/cmd/goimports returned an error")
		}

		cmd = exec.CommandContext(ctx, "./.bin/goimports", "-w")
		if verbosity == generateVerbose {
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
		}
		err = cmd.Run()
		if err != nil {
			return errors.Wrap(err, "goimports -w returned an error")
		}
	} else {
		cmd = exec.CommandContext(ctx, "goimports", "-w")
		cmd.Env = os.Environ()
		if verbosity == generateVerbose {
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
		}
		err := cmd.Run()
		if err != nil {
			return errors.Wrap(err, "goimports -w returned an error")
		}
	}

	// Run go mod tidy
	if verbosity != generateQuiet {
		stdOut.WriteLine(output.Linef(output.EmojiInfo, output.StyleBold, "go mod tidy"))
	}
	cmd = exec.CommandContext(ctx, "go", "mod", "tidy")
	if verbosity == generateVerbose {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
	}
	err = cmd.Run()
	if err != nil {
		return errors.Wrap(err, "go mod tidy returned an error")
	}

	return nil
}

func generateGoGenerate(pkgPaths []string, verbosity generateVerbosityType) error {
	args := []string{"generate"}
	if verbosity == generateVerbose {
		args = append(args, "-x")
	}
	args = append(args, pkgPaths...)
	cmd := exec.Command("go", args...)
	if verbosity == generateVerbose {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
	}

	err := cmd.Start()
	if err != nil {
		return errors.Errorf("go generate %s returned an error: %w\n%s", err)
	}

	err = cmd.Wait()
	if err != nil {
		return errors.Errorf("go generate %s returned an error: %w\n%s", err)
	}
	return nil
}
