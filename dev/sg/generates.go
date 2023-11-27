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
	{
		Name:   "bazel",
		Help:   "Run the bazel target //dev:write_all_generated",
		Runner: generateBazelRunner,
	},
}

func generateBazelRunner(ctx context.Context, args []string) *generate.Report {
	if report := generate.RunScript("bazel run //dev:write_all_generated")(ctx, args); report.Err != nil {
		return report
	}

	return &generate.Report{}
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
	// If args are provided, assume the args are paths to buf configuration
	// files - so we just generate over specifiied configuration files.
	if len(args) > 0 {
		return proto.Generate(ctx, args, verbose)
	}

	// If we're not in CI, we check for proto file changes
	if os.Getenv("CI") != "true" {
		out, err := exec.Command("git", "diff", "--name-only", "main").Output()
		if err != nil {
			return &generate.Report{Err: errors.Wrap(err, "git diff failed")} // should never happen
		}

		if !strings.Contains(string(out), ".proto") {
			return &generate.Report{Output: "No .proto files changed or not in CI. Skipping buf gen.\n"}
		}
	}

	// By default, we will run buf generate in every directory with buf.gen.yaml
	bufGenFilePaths, err := buf.PluginConfigurationFiles()
	if err != nil {
		return &generate.Report{Err: errors.Wrapf(err, "finding plugin configuration files")}
	}

	// Run buf gen by default
	return proto.Generate(ctx, bufGenFilePaths, verbose)
}
