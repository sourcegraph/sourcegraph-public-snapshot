package main

import (
	"context"
	"flag"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/generate"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/generate/golang"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	generateGoFlagSet     = flag.NewFlagSet("sg generate go", flag.ExitOnError)
	generateGoVerboseFlag = generateGoFlagSet.Bool("v", false, "Display output from go generate")
	generateGoQuietFlag   = generateGoFlagSet.Bool("q", false, "Suppress all output but errors from go generate")
)

var allGenerateTargets = generateTargets{
	{
		Name:    "go",
		Help:    "Run go generate [packages...] on the codebase",
		FlagSet: generateGoFlagSet,
		Runner:  generateGoRunner,
	},
}

func generateGoRunner(ctx context.Context, args []string) *generate.Report {
	if *generateGoVerboseFlag && *generateGoQuietFlag {
		return &generate.Report{Err: errors.Errorf("-q and -v flags are exclusive")}
	}
	if *generateGoVerboseFlag {
		return golang.Generate(ctx, args, golang.VerboseOutput)
	} else if *generateGoQuietFlag {
		return golang.Generate(ctx, args, golang.QuietOutput)
	} else {
		return golang.Generate(ctx, args, golang.NormalOutput)
	}
}
