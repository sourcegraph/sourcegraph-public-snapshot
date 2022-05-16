package main

import (
	"context"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/generate"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/generate/golang"
)

var allGenerateTargets = generateTargets{
	{
		Name:   "go",
		Help:   "Run go generate [packages...] on the codebase",
		Runner: generateGoRunner,
	},
}

func generateGoRunner(ctx context.Context, args []string) *generate.Report {
	if verbose {
		return golang.Generate(ctx, args, golang.VerboseOutput)
	} else if generateQuiet {
		return golang.Generate(ctx, args, golang.QuietOutput)
	} else {
		return golang.Generate(ctx, args, golang.NormalOutput)
	}
}
