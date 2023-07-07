package main

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/buf"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/generate"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/generate/golang"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/generate/proto"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var allGenerateTargets = generateTargets{
	{
		// Protocol Buffer generation runs before go, as otherwise
		// go mod tidy errors on generated protobuf go code directories.
		Name:   "buf",
		Help:   "Re-generate protocol buffer bindings using buf",
		Runner: generateProtoRunner,
		Completer: func() (options []string) {
			options, _ = buf.CodegenFiles()
			return
		},
	},
	{
		Name:   "go",
		Help:   "Run go generate [packages...] on the codebase",
		Runner: generateGoRunner,
		Completer: func() (options []string) {
			root, err := root.RepositoryRoot()
			if err != nil {
				return
			}
			options, _ = golang.FindFilesWithGenerate(root)
			return
		},
	},
}

func generateGoRunner(ctx context.Context, args []string) *generate.Report {
	if verbose {
		return golang.Generate(ctx, args, true, golang.VerboseOutput)
	} else if generateQuiet {
		return golang.Generate(ctx, args, true, golang.QuietOutput)
	} else {
		return golang.Generate(ctx, args, true, golang.NormalOutput)
	}
}

func generateProtoRunner(ctx context.Context, args []string) *generate.Report {
	// Always run in CI
	if os.Getenv("CI") == "true" {
		return proto.Generate(ctx, verbose)
	}

	// Check to see if any .proto files changed
	out, err := exec.Command("git", "diff", "--name-only", "main...HEAD").Output()
	if err != nil {
		return &generate.Report{Err: errors.Wrap(err, "git diff failed")} // should never happen
	}

	// Don't run buf gen if no .proto files changed or not in CI
	if !strings.Contains(string(out), ".proto") {
		return &generate.Report{Output: "No .proto files changed or not in CI. Skipping buf gen.\n"}
	}
	// Run buf gen by default
	return proto.Generate(ctx, verbose)
}
