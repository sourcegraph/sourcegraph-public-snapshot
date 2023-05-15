package main

import (
	"context"
	"os/exec"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/buf"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/generate"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/generate/golang"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/generate/proto"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
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
	out, err := exec.Command("git", "diff", "--name-only", "origin/master...").Output()
	if err != nil {
		panic(err)
	}

	// Check if output contains any .proto files
	if strings.Contains(string(out), ".proto") {
		println("Diff contains .proto changes!")
		return proto.Generate(ctx, verbose)
	} else {
		println("No .proto changes found")
		return nil
	}
}
