// Package linters defines all available linters.
package linters

import (
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"
)

// Targets lists all available linter targets. Each target consists of multiple linters.
//
// These should align with the names in 'enterprise/dev/ci/internal/ci/changed'
var Targets = []lint.Target{
	{
		Name: "urls",
		Help: "Check for broken urls in the codebase",
		Linters: []lint.Runner{
			lint.RunScript("Broken urls", "dev/check/broken-urls.bash"),
		},
	},
	{
		Name: "go",
		Help: "Check go code for linting errors, forbidden imports, generated files, etc",
		Linters: []lint.Runner{
			goFmt,
			lintGoGenerate,
			goLint,
			goDBConnImport,
			goEnterpriseImport,
			noLocalHost,
			lintGoDirectives,
			lintLoggingLibraries(),
			goModGuards(),
		},
	},
	{
		Name: "docs",
		Help: "Documentation checks",
		Linters: []lint.Runner{
			lint.RunScript("Docsite lint", "dev/check/docsite.sh"),
		},
	},
	{
		Name: "dockerfiles",
		Help: "Check Dockerfiles for Sourcegraph best practices",
		Linters: []lint.Runner{
			hadolint(),
			customDockerfileLinters(),
		},
	},
	{
		Name: "client",
		Help: "Check client code for linting errors, forbidden imports, etc",
		Linters: []lint.Runner{
			tsEnterpriseImport,
			inlineTemplates,
			lint.RunScript("Yarn duplicate", "dev/check/yarn-deduplicate.sh"),
			checkUnversionedDocsLinks(),
		},
	},
	{
		Name: "svg",
		Help: "Check svg assets",
		Linters: []lint.Runner{
			checkSVGCompression(),
		},
	},
	{
		Name: "shell",
		Help: "Check shell code for linting errors, formatting, etc",
		Linters: []lint.Runner{
			shFmt,
			shellCheck,
			bashSyntax,
		},
	},
}
