// Package linters defines all available linters.
package linters

import (
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"
)

// Targets lists all available linter targets. Each target consists of multiple linters.
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
		},
	},
	{
		Name: "go-custom",
		Help: "[WILL BE DEPRECATED] Custom checks for Go, will be migrated to the default go check set in the future",
		Linters: []lint.Runner{
			lintLoggingLibraries(),
			goModGuards(),
		},
	},
	{
		Name: "docsite",
		Help: "Check the code powering docs.sourcegraph.com for broken links and linting errors",
		Linters: []lint.Runner{
			lint.RunScript("Docsite lint", "dev/check/docsite.sh"),
		},
	},
	{
		Name: "docker",
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
			lint.RunScript("SVG Compression", "dev/check/svgo.sh"),
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
	{
		Name: "check-all-compat",
		Help: "[WILL BE DEPRECATED] - 1:1 compatibility with the legacy ./dev/check/all.sh script",
		Linters: []lint.Runner{
			goFmt,
			inlineTemplates,
			goEnterpriseImport,
			goDBConnImport,
			lintGoGenerate,
			goLint,
			noLocalHost,
			bashSyntax,
			shFmt,
			shellCheck,
			submodules,
			lintGoDirectives,
		},
	},
}
