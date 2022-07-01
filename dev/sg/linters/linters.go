// Package linters defines all available linters.
package linters

import (
	"context"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
)

type Target = check.Category[*repo.State]

type linter = check.Check[*repo.State]

// Targets lists all available linter targets. Each target consists of multiple linters.
//
// These should align with the names in 'enterprise/dev/ci/internal/ci/changed'
var Targets = []Target{
	{
		Name:        "urls",
		Description: "Check for broken urls in the codebase",
		Checks: []*linter{
			runScript("Broken urls", "dev/check/broken-urls.bash"),
		},
	},
	{
		Name:        "go",
		Description: "Check go code for linting errors, forbidden imports, generated files, etc",
		Checks: []*linter{
			goFmt,
			goGenerateLinter,
			goLint(),
			goDBConnImport,
			goEnterpriseImport,
			noLocalHost,
			lintGoDirectives(),
			// lintLoggingLibraries(),
			goModGuards(),
			lintSGExit(),
		},
	},
	{
		Name:        "docs",
		Description: "Documentation checks",
		Checks: []*linter{
			runScript("Docsite lint", "dev/docsite.sh check"),
		},
	},
	{
		Name:        "dockerfiles",
		Description: "Check Dockerfiles for Sourcegraph best practices",
		Checks: []*linter{
			hadolint(),
			customDockerfileLinters(),
		},
	},
	{
		Name:        "client",
		Description: "Check client code for linting errors, forbidden imports, etc",
		Checks: []*linter{
			tsEnterpriseImport,
			inlineTemplates,
			runScript("Yarn duplicate", "dev/check/yarn-deduplicate.sh"),
			checkUnversionedDocsLinks(),
		},
	},
	{
		Name:        "svg",
		Description: "Check svg assets",
		Checks: []*linter{
			checkSVGCompression(),
		},
	},
	{
		Name:        "shell",
		Description: "Check shell code for linting errors, formatting, etc",
		Checks: []*linter{
			shFmt,
			shellCheck,
			bashSyntax,
		},
	},
}

// runScript creates check that runs the given script from the root of sourcegraph/sourcegraph.
func runScript(name string, script string) *linter {
	return &linter{
		Name: name,
		Check: func(ctx context.Context, out *std.Output, args *repo.State) error {
			return root.Run(run.Bash(ctx, script)).StreamLines(out.Write)
		},
	}
}

// runCheck creates a check that runs the given check func.
func runCheck(name string, check check.CheckAction[*repo.State]) *linter {
	return &linter{
		Name:  name,
		Check: check,
	}
}
